package haleakala

import (
	"math/rand"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/rs/zerolog/log"
	"github.com/tankbusta/haleakala/muxer"

	"github.com/bwmarrin/discordgo"
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
	mux  *muxer.Mux
	rwmu *sync.RWMutex
	wg   *sync.WaitGroup
	stor *storm.DB

	// Plugin Stuff
	plugins map[string]IBasicPlugin
}

type InstallFunc func(command, description string, middlewares ...muxer.HandlerFunc) error

func New(configPath string) (*Context, error) {
	cfg, err := loadconfig(configPath)
	if err != nil {
		return nil, err
	}

	// Set the default AdminMiddleware to admin only channels
	DefaultAdminMiddleware = AllowOnCertainChannels(cfg.Discord.AdminChannels)

	db, err := initStorage(cfg.DatabaseConfig)
	if err != nil {
		return nil, err
	}

	ds, err := discordgo.New(cfg.Discord.Username, cfg.Discord.Password, cfg.Discord.Token)
	if err != nil {
		return nil, err
	}

	return &Context{
		cfg:     cfg,
		ds:      ds,
		stop:    make(chan bool, 1),
		mux:     muxer.New(),
		wg:      &sync.WaitGroup{},
		rwmu:    &sync.RWMutex{},
		plugins: make(map[string]IBasicPlugin),
		stor:    db,
	}, nil
}

// installRoute installs a new command into the bot
func (s *Context) InstallRoute(command, description string, middlewares ...muxer.HandlerFunc) error {
	s.mux.Route(command, description, middlewares...)
	return nil
}

// InitializePlugin registers a plugin with the bot & router
func (s *Context) InitializePlugin(f IPlugin) error {
	return nil
}

func (s *Context) loadPluginsFromConfig() error {
	for _, plug := range s.cfg.Plugins {
		if !plug.Enabled {
			log.Warn().
				Str("plugin", plug.PluginName).
				Msg("Skipping plugin load as it's not enabled")
			continue
		}

		found := GetPlugin(plug.PluginName)
		if found == nil {

			log.Warn().
				Str("plugin", plug.PluginName).
				Msg("Plugin not found")
			continue
		}

		plugin, err := found.Initialize(PluginConfigVars(plug.Config), s.ds, s.createBucketForPlugin(plug.PluginName))
		if err != nil {
			log.Error().
				Err(err).
				Str("plugin", plug.PluginName).
				Msg("Plugin could not be initialized properly")
			return err
		}

		if err := plugin.InstallRoute(s.InstallRoute); err != nil {
			log.Error().
				Err(err).
				Str("plugin", plug.PluginName).
				Msg("Plugin could not be initialized properly")
			return err
		}

		s.rwmu.Lock()
		s.plugins[plugin.Name()] = plugin
		s.rwmu.Unlock()

		log.Warn().
			Str("plugin", plug.PluginName).
			Msg("Plugin loaded successfully!")
	}
	return nil
}

// Start connects to discord and starts listening for messages
func (s *Context) Start() error {
	if err := s.loadPluginsFromConfig(); err != nil {
		return err
	}

	s.ds.AddHandler(s.mux.OnMessageCreate)
	if err := s.ds.Open(); err != nil {
		return err
	}

	// Fetch our user record to ensure we've successfully logged in
	if _, err := s.ds.User("@me"); err != nil {
		return err
	}

	s.mux.Route("help", "Get a current list of all commands", s.mux.Help)

	// Plugin Management Commands
	s.mux.Route("plugins", "", AllowOnCertainChannels(s.cfg.Discord.AdminChannels), s.ListLoadedPlugins)
	s.mux.Route("unload", "", AllowOnCertainChannels(s.cfg.Discord.AdminChannels), s.UnloadPlugin)
	s.mux.Route("load", "", AllowOnCertainChannels(s.cfg.Discord.AdminChannels), s.LoadPlugin)

	if len(s.cfg.Discord.StatusChanger.Statuses) >= 1 {
		s.wg.Add(1)
		go s.changeStatus()
	}

	return nil
}

func (s *Context) changeStatus() {
	defer s.wg.Done()
	if s.cfg.Discord.StatusChanger.EverySecond == 0 {
		s.cfg.Discord.StatusChanger.EverySecond = 60
	}

	ticker := time.NewTicker(time.Duration(s.cfg.Discord.StatusChanger.EverySecond) * time.Second)
	defer ticker.Stop()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	statuses := s.cfg.Discord.StatusChanger.Statuses

Loop:
	for {
		select {
		case <-s.stop:
			break Loop
		case <-ticker.C:
		}
		selectedStatus := statuses[r.Intn(len(statuses))]

		log.Debug().
			Str("status", selectedStatus).
			Msg("Changing game status")

		if err := s.ds.UpdateStatus(0, selectedStatus); err != nil {
			log.Error().Err(err).Msgf("Failed to change game status to `%s`", selectedStatus)
		}
	}
}

// Stop will end the bot safely
func (s *Context) Stop() error {
	close(s.stop)
	// Wait for all our main goroutines to exit..
	s.wg.Wait()

	for _, p := range s.plugins {
		if advp, ok := p.(IPlugin); ok {
			advp.Destroy()
		}
	}

	return s.ds.Close()
}
