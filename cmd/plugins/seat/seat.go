package seat

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"
	"github.com/tankbusta/libseat"

	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
)

type (
	pluginConfig struct {
		APIEndpoint string `json:"server"`
		APIKey      string `json:"key"`
	}

	plugin     struct{}
	seatPlugin struct {
		client      *libseat.Client
		mapping     map[string]int // map[discord_id]group_id
		lastUpdated time.Time
		server      string
		mu          *sync.Mutex
	}
)

var (
	how2use                = "Usage: `!seat <command>`. The following commands are available: characters"
	registerMessage        = "Hi! Before using any of the !seat commands, you must tie your SeAT account to Discord. Please visit %s"
	cacheCanBeUpdatedEvery = time.Minute * 15
)

func init() {
	haleakala.Register("seat", &plugin{})
}

func (s plugin) Initialize(cfg haleakala.PluginConfigVars, ds *discordgo.Session, _ storm.Node) (haleakala.IPlugin, error) {
	var pcfg pluginConfig

	if err := json.Unmarshal(cfg, &pcfg); err != nil {
		return nil, err
	}

	if pcfg.APIEndpoint == "" {
		// Do we have it as an environment variable?
		seatsrv := os.Getenv("SEAT_API_SERVER")
		if seatsrv == "" {
			return nil, errors.New("Either SEAT_API_SERVER or seat.server must be set")
		}
		pcfg.APIEndpoint = seatsrv
	}

	if pcfg.APIKey == "" {
		// Do we have it as an environment variable?
		apikey := os.Getenv("SEAT_API_KEY")
		if apikey == "" {
			return nil, errors.New("Either SEAT_API_KEY or seat.key must be set")
		}
		pcfg.APIKey = apikey
	}

	client, err := libseat.NewClient(pcfg.APIEndpoint, pcfg.APIKey)
	if err != nil {
		return nil, err
	}

	plug := &seatPlugin{
		client:  client,
		mapping: make(map[string]int),
		server:  pcfg.APIEndpoint,
		mu:      &sync.Mutex{},
	}

	if err := plug.updateCache(); err != nil {
		return nil, errors.New("unable to update the discord mapping cache. skipping plugin init")
	}

	return plug, nil
}

func (s *seatPlugin) updateCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mapping, err := s.client.GetDiscordMapping()
	if err != nil {
		return err
	}

	for _, discordUser := range mapping {
		// sigh... seat treats discordIDs as integers when snowflakes should be treated
		// as strings
		discoid := strconv.FormatInt(discordUser.DiscordID, 10)
		s.mapping[discoid] = discordUser.SeatGroup
	}
	s.lastUpdated = time.Now()

	return nil
}

func (s seatPlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "Allows you to query your SeAT/EVE Information", s.OnMessage)
}

// OnMessage returns an about message
func (s *seatPlugin) OnMessage(ctx *muxer.Context) {
	parts := ctx.Fields
	if len(parts) <= 1 {
		ctx.SendPrivately(how2use)
		return
	}

	// Do we have the user id?
	groupID, ok := s.mapping[ctx.FromUserID]
	if !ok {
		// Update the cache...
		cacheWasLastUpdated := time.Now().Sub(s.lastUpdated)
		if cacheWasLastUpdated < cacheCanBeUpdatedEvery {
			tryAgain := cacheCanBeUpdatedEvery - cacheWasLastUpdated
			ctx.SendPrivately(fmt.Sprintf("Please wait %s before trying again", tryAgain.String()))
			return
		}

		if err := s.updateCache(); err != nil {
			ctx.SendPrivately("Sorry m8 try again later")
			return
		}

		// Try again after the cache update
		groupID, ok = s.mapping[ctx.FromUserID]
		if !ok {
			ctx.SendPrivately(fmt.Sprintf(registerMessage, s.server))
			return
		}
	}

	switch strings.ToLower(parts[1]) {
	case "characters", "chars":
		s.sendCharacterInfo(groupID, ctx)
	default:
		ctx.SendPrivately(how2use)
	}
}

func (s seatPlugin) SupportsUnload() bool {
	return true
}

func (s seatPlugin) Destroy() error {
	return nil
}

func (s seatPlugin) Name() string {
	return "seat"
}
