package sounds

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"

	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

const StartEndFormatting = "```"

var HelpStringGlobal = ""

type Options struct {
	Bitrate            int    `yaml:"bitrate" json:"bitrate"`
	MaxQueueSize       int    `yaml:"max_queue" json:"max_queue"`
	AdminUserSnowflake string `yaml:"admin_user_id" json:"admin_user_id"`
	AudioPath          string `yaml:"audio_path" json:"audio_path"`
}

func (s *Options) validate() error {
	if s.Bitrate == 0 {
		s.Bitrate = 128
	}

	if s.MaxQueueSize == 0 {
		s.MaxQueueSize = 6
	}

	if s.AudioPath == "" {
		return errors.New("sound.audio_path must not be empty and set to a folder containing audio sounds")
	}

	return nil
}

func init() {
	haleakala.Register("sound", &soundPlugin{})
}

type soundPlugin struct{}

func (s soundPlugin) Initialize(cfg haleakala.PluginConfigVars, ds *discordgo.Session, _ storm.Node) (haleakala.IPlugin, error) {
	var opts Options

	if err := json.Unmarshal(cfg, &opts); err != nil {
		return nil, err
	}

	if err := opts.validate(); err != nil {
		return nil, err
	}

	for _, coll := range COLLECTIONS {
		coll.Load(opts.AudioPath)
	}

	return &SoundBotPlugin{
		queues:      make(map[string]chan *Play),
		ds:          ds,
		adminUserID: opts.AdminUserSnowflake,
	}, nil
}

type SoundBotPlugin struct {
	// unexported fields below
	opts Options
	ds   *discordgo.Session

	adminUserID string
	// Map of Guild id's to *Play channels, used for queuing and rate-limiting guilds
	queues map[string]chan *Play
	mu     sync.Mutex
}

func (s *SoundBotPlugin) Destroy() error {
	return nil
}

func (s *SoundBotPlugin) SupportsUnload() bool {
	return true
}
func (s *SoundBotPlugin) Name() string {
	return "sound"
}

func (s *SoundBotPlugin) dumpAvailableSounds(c *muxer.Context) {
	// Cache the generation of this message after one time...
	if HelpStringGlobal == "" {
		out := fmt.Sprintf("The following sound collections are available for your pleasure. To use them type `!%s <collection name> <optional sound>`. Some collections have multiple names, delimited by / but you only need one.\n\n", s.Name())
		for _, coll := range COLLECTIONS {
			out += fmt.Sprintf("**Collection: %s**\n%s", strings.Join(coll.Commands, "/"), StartEndFormatting)
			for _, sound := range coll.Sounds {
				out += fmt.Sprintf("* %s\n", sound.Name)
			}
			out += StartEndFormatting + "\n"
		}
		HelpStringGlobal = out
	}

	pmc, err := s.ds.UserChannelCreate(c.Message.Author.ID)
	if err != nil {
		// couldnt create message :(
		c.Send(fmt.Sprintf("Sorry %s, I tried to send you a PM but i couldn't", c.Message.Author.Username))
		return
	}

	s.ds.ChannelMessageSend(pmc.ID, HelpStringGlobal)
	c.Send("Sent a list of sounds to you via PM!")
}

func (s *SoundBotPlugin) OnMessage(c *muxer.Context) {
	if len(c.Fields) == 1 {
		s.dumpAvailableSounds(c)
		return
	}

	// We only want the fields past our command
	parts := c.Fields[1:]
	m := c.Message

	channel, err := s.ds.State.Channel(m.ChannelID)
	if channel == nil || err != nil {
		logrus.WithFields(logrus.Fields{
			"channel": m.ChannelID,
			"message": m.ID,
			"error":   err,
		}).Warning("Failed to grab channel")
		return
	}

	guild, err := s.ds.State.Guild(channel.GuildID)
	if guild == nil || err != nil {
		logrus.WithFields(logrus.Fields{
			"guild":   channel.GuildID,
			"channel": channel,
			"message": m.ID,
			"error":   err,
		}).Warning("Failed to grab guild")
		return
	}

	// Someone set us up the bomb
	if parts[0] == "airbomb" && s.adminUserID == c.FromUserID {
		var minBombs = 50
		// Allow for some configuration of how many "horns"
		if len(parts) == 2 {
			if minBombs, err = strconv.Atoi(parts[1]); err != nil {
				minBombs = 50
			}

			// cap it
			if minBombs > 100 {
				minBombs = 50
			}
		}

		c.Send(fmt.Sprintf(":ok_hand: %s", strings.Repeat(":trumpet:", minBombs)))

		play := s.createPlay(m.Author, guild, airhorn, nil)
		vc, err := s.ds.ChannelVoiceJoin(play.GuildID, play.ChannelID, true, true)
		if err != nil {
			c.Send(fmt.Sprintf(":cry: failed to have fun: %s", err.Error()))
			return
		}
		defer vc.Disconnect()

		for i := 0; i < minBombs; i++ {
			airhorn.Random().Play(vc)
		}

		return
	} else if parts[0] == "airbomb" && s.adminUserID != c.FromUserID {
		c.Send(":middle_finger: get rekt m8")
		return
	}

	// Find the collection for the command we got
	for _, coll := range COLLECTIONS {
		if scontains(parts[0], coll.Commands...) {
			// If they passed a specific sound effect, find and select that (otherwise play nothing)
			var sound *Sound
			if len(parts) > 1 {
				for _, s := range coll.Sounds {
					if parts[1] == s.Name {
						sound = s
					}
				}

				if sound == nil {
					return
				}
			}

			go s.enqueuePlay(m.Author, guild, coll, sound)
			return
		}
	}

	c.Send(fmt.Sprintf(":cry: Sorry I cant find the sound `%s`", strings.Join(parts, " ")))
}

func (s *SoundBotPlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "Play a sound to the discord voice channel that you're currently in", s.OnMessage)
}

// Attempts to find the current users voice channel inside a given guild
func (s *SoundBotPlugin) getCurrentVoiceChannel(user *discordgo.User, guild *discordgo.Guild) *discordgo.Channel {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			channel, _ := s.ds.State.Channel(vs.ChannelID)
			return channel
		}
	}
	return nil
}

func (s *SoundBotPlugin) createPlay(user *discordgo.User, guild *discordgo.Guild, coll *SoundCollection, sound *Sound) *Play {
	// Grab the users voice channel
	channel := s.getCurrentVoiceChannel(user, guild)
	if channel == nil {
		return nil
	}

	// Create the play
	play := &Play{
		GuildID:   guild.ID,
		ChannelID: channel.ID,
		UserID:    user.ID,
		Sound:     sound,
		Forced:    true,
	}

	// If we didn't get passed a manual sound, generate a random one
	if play.Sound == nil {
		play.Sound = coll.Random()
		play.Forced = false
	}

	// If the collection is a chained one, set the next sound
	if coll.ChainWith != nil {
		play.Next = &Play{
			GuildID:   play.GuildID,
			ChannelID: play.ChannelID,
			UserID:    play.UserID,
			Sound:     coll.ChainWith.Random(),
			Forced:    play.Forced,
		}
	}

	return play
}

func (s *SoundBotPlugin) playSound(play *Play, vc *discordgo.VoiceConnection) (err error) {
	logrus.WithFields(logrus.Fields{
		"play": play,
	}).Info("Playing sound")

	if vc == nil {
		vc, err = s.ds.ChannelVoiceJoin(play.GuildID, play.ChannelID, false, false)
		if err != nil {
			s.mu.Lock()
			delete(s.queues, play.GuildID)
			s.mu.Unlock()
			return err
		}
	}
	defer vc.Disconnect()

	// If we need to change channels, do that now
	if vc.ChannelID != play.ChannelID {
		vc.ChangeChannel(play.ChannelID, false, false)
		time.Sleep(time.Millisecond * 125)
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(time.Millisecond * 32)

	// Play the sound
	play.Sound.Play(vc)

	// If this is chained, play the chained sound
	if play.Next != nil {
		s.playSound(play.Next, vc)
	}

	// If there is another song in the queue, recurse and play that
	if len(s.queues[play.GuildID]) > 0 {
		play := <-s.queues[play.GuildID]
		s.playSound(play, vc)
		return nil
	}

	// If the queue is empty, delete it
	time.Sleep(time.Millisecond * time.Duration(play.Sound.PartDelay))

	s.mu.Lock()
	delete(s.queues, play.GuildID)
	s.mu.Unlock()

	return nil
}

func (s *SoundBotPlugin) enqueuePlay(user *discordgo.User, guild *discordgo.Guild, coll *SoundCollection, sound *Sound) {
	play := s.createPlay(user, guild, coll, sound)
	if play == nil {
		return
	}

	s.mu.Lock()
	_, exists := s.queues[guild.ID]
	if exists {
		if len(s.queues[guild.ID]) < s.opts.MaxQueueSize {
			s.queues[guild.ID] <- play
		}
		s.mu.Unlock()
	} else {
		s.queues[guild.ID] = make(chan *Play, s.opts.MaxQueueSize)
		s.mu.Unlock()
		s.playSound(play, nil)
	}
}
