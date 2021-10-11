package haleakala

import (
	"fmt"
	"sync"

	"github.com/tankbusta/haleakala/plugin"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	// CurrentVersion is set by the linker and contains the Git Commit hash
	CurrentVersion = ""
	// BuildTime is set by the linker and includes the time in UTC on which the bot was compiled
	BuildTime = ""
)

// Context TODO
type Context struct {
	// unexported fields below
	cfg  *config
	ds   *discordgo.Session
	stop chan bool
	wg   *sync.WaitGroup

	// Plugin Stuff
	plugins map[string]plugin.IBasicPlugin
}

func New(configPath string) (*Context, error) {
	cfg, err := loadconfig(configPath)
	if err != nil {
		return nil, err
	}

	ds, err := discordgo.New(fmt.Sprintf("Bot %s", cfg.Discord.Token))
	if err != nil {
		return nil, err
	}

	return &Context{
		cfg:     cfg,
		ds:      ds,
		stop:    make(chan bool, 1),
		wg:      &sync.WaitGroup{},
		plugins: make(map[string]plugin.IBasicPlugin),
	}, nil
}

// Start connects to discord and starts listening for messages
func (s *Context) Start() error {
	slashCommands := make(map[string]plugin.IBasicPlugin)

	s.ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info().
			Int("version", r.Version).
			Msg("Bot is ready!")
	})

	s.ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if p, ok := slashCommands[i.ApplicationCommandData().Name]; ok {
			p.OnInteraction(s, i)
		}
	})

	if err := s.ds.Open(); err != nil {
		return err
	}

	for _, plugName := range plugin.GetListOfPlugins() {
		plug := plugin.GetPlugin(plugName)
		for _, cmd := range plug.Commands() {
			log.Info().
				Str("plugin", plugName).
				Msgf("Initializing command `%s`", cmd.Name)

			if _, err := s.ds.ApplicationCommandCreate(s.ds.State.User.ID, "896809463141511208", &cmd); err != nil {
				log.Error().
					Err(err).
					Str("plugin", plugName).
					Msgf("Failed to initialize command `%s`", cmd.Name)
				continue
			}

			slashCommands[cmd.Name] = plug
		}

		s.plugins[plugName] = plug
	}

	return nil
}

// Stop will end the bot safely
func (s *Context) Stop() error {
	close(s.stop)
	// Wait for all our main goroutines to exit..
	s.wg.Wait()

	for _, p := range s.plugins {
		if advp, ok := p.(plugin.IPlugin); ok {
			advp.Destroy()
		}
	}

	return s.ds.Close()
}
