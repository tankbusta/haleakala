package haleakala

import (
	"fmt"

	"github.com/tankbusta/haleakala/database"
	"github.com/tankbusta/haleakala/plugin"
	"github.com/uptrace/bun"

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
	cfg     *config
	ds      *discordgo.Session
	db      *bun.DB
	plugins map[string]plugin.IBasicPlugin
}

func New(configPath string) (*Context, error) {
	cfg, err := loadconfig(configPath)
	if err != nil {
		return nil, err
	}

	db, err := database.GetDatabase(cfg.DatabaseConfig)
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
		db:      db,
		plugins: make(map[string]plugin.IBasicPlugin),
	}, nil
}

// Start connects to discord and starts listening for messages
func (s *Context) Start() error {
	slashCommands := make(map[string]plugin.IBasicPlugin)
	ready := make(chan struct{}, 1)

	s.ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info().
			Int("version", r.Version).
			Msg("Bot is ready!")

		// Notify that we're ready to start initializing commands and complete the bot's
		// initialization. Without this, we can panic while creating slash commands.
		ready <- struct{}{}
	})

	s.ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if p, ok := slashCommands[i.ApplicationCommandData().Name]; ok {
			p.OnInteraction(s, i)
		}
	})

	if err := s.ds.Open(); err != nil {
		return err
	}

	// TODO: A timeout of sorts
	<-ready
	// ready channel is no longer needed
	close(ready)

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

		if err := plug.OnLoad(s.ds, s.db); err != nil {
			log.Error().
				Err(err).
				Msgf("Failed to initialize plugin `%s`", plugName)
		}

		s.plugins[plugName] = plug
	}

	return nil
}

// Stop will end the bot safely
func (s *Context) Stop() error {
	for _, p := range s.plugins {
		log.Warn().
			Msgf("Unloading plugin `%s`", p.Name())

		if err := p.OnShutdown(s.ds); err != nil {
			log.Error().
				Err(err).
				Msgf("Failed to shutting down plugin `%s`", p.Name())
		}
	}

	return s.ds.Close()
}
