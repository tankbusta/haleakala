package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/cmd/plugins"
)

var (
	configPath       string
	showDebug        bool
	showPlugins      bool
	failOnRouteError bool
)

func init() {
	flag.StringVar(&configPath, "config", os.Getenv("BOT_CONFIG"), "path to configuration file")
	flag.BoolVar(&showDebug, "debug", false, "enable verbose output")
	flag.BoolVar(&showPlugins, "show-plugins", false, "list plugins and exit")
	flag.BoolVar(&failOnRouteError, "route-failure-fatal", false, "route/plugin issues are considered a fatal error")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if showDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func main() {
	if configPath == "" {
		log.Fatal().Msg("-config/BOT_CONFIG must be set")
	}

	if showPlugins {
		fmt.Println("Available Plugins:")
		for _, plugin := range plugins.DefaultPlugins {
			fmt.Printf("* %s\n", plugin.Name())
		}

		for _, plugin := range haleakala.GetListOfPlugins() {
			fmt.Printf("* %s\n", plugin)
		}
		os.Exit(0)
	}

	ctx, err := haleakala.New(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create bot")
	}

	sig := make(chan os.Signal, 1)
	finished := make(chan bool)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	if err := ctx.Start(); err != nil {
		log.Error().Err(err).Msg("failed to start bot")
		os.Exit(-1)
	}

	// Install the default plugins
	for _, plugin := range plugins.DefaultPlugins {
		if err := plugin.InstallRoute(ctx.InstallRoute); err != nil {
			if failOnRouteError {
				log.Fatal().
					Err(err).
					Msgf("An error occured while loading plugin %s", plugin.Name())
			}

			log.Error().
				Err(err).
				Msgf("An error occured while loading plugin %s", plugin.Name())
		}
	}

	go func() {
		for _ = range sig {
			log.Warn().Msg("Received SIGINT, shutting down")
			ctx.Stop()
			finished <- true
		}
	}()

	<-finished
}
