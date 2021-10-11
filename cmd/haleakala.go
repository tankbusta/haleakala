package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/tankbusta/haleakala"
	_ "github.com/tankbusta/haleakala/plugin/plugins"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configPath string
	showDebug  bool
)

func init() {
	flag.StringVar(&configPath, "config", os.Getenv("BOT_CONFIG"), "path to configuration file")
	flag.BoolVar(&showDebug, "debug", false, "enable verbose output")
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

	go func() {
		for _ = range sig {
			log.Warn().Msg("Received SIGINT, shutting down")
			ctx.Stop()
			finished <- true
		}
	}()

	<-finished
}
