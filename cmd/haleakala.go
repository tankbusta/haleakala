package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/cmd/plugins"

	"github.com/sirupsen/logrus"
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
}

func main() {
	log.SetOutput(logrus.StandardLogger().Out)

	if configPath == "" {
		logrus.Fatal("-config/BOT_CONFIG must be set")
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
		logrus.WithError(err).Fatal("Failed to start bot")
	}

	if showDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	sig := make(chan os.Signal, 1)
	finished := make(chan bool)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	if err := ctx.Start(); err != nil {
		logrus.WithError(err).Fatal("Failed to start bot")
		os.Exit(-1)
	}

	// Install the default plugins
	for _, plugin := range plugins.DefaultPlugins {
		if err := plugin.InstallRoute(ctx.InstallRoute); err != nil {
			if failOnRouteError {
				logrus.WithError(err).Fatalf("An error occured while loading plugin %s", plugin.Name())
			}

			logrus.WithError(err).Errorf("An error occured while loading plugin %s", plugin.Name())
		}
	}

	go func() {
		for _ = range sig {
			logrus.Warn("Stopping haleakala")
			ctx.Stop()
			finished <- true
		}
	}()

	<-finished
}
