package plugin

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/uptrace/bun"
)

var (
	pluginsMu sync.RWMutex
	plugins   = make(map[string]IBasicPlugin)
)

// Register makes a plugin available to the system
func Register(name string, plugin IBasicPlugin) {
	pluginsMu.Lock()
	defer pluginsMu.Unlock()

	if plugin == nil {
		panic("cannot register a nil plugin")
	}

	if _, exists := plugins[name]; exists {
		panic(fmt.Sprintf("plugin %s already exists", name))
	}

	plugins[name] = plugin
}

func GetListOfPlugins() []string {
	pluginsMu.Lock()
	defer pluginsMu.Unlock()

	out := make([]string, 0)
	for name := range plugins {
		out = append(out, name)
	}

	return out
}

func GetPlugin(name string) IBasicPlugin {
	pluginsMu.Lock()
	defer pluginsMu.Unlock()

	if plug, exists := plugins[name]; exists {
		return plug
	}

	return nil
}

type PluginConfigVars json.RawMessage

type IBasicPlugin interface {
	// Name returns the name of the plugin
	Name() string

	Commands() []discordgo.ApplicationCommand

	OnLoad(s *discordgo.Session, db *bun.DB) error

	OnShutdown(s *discordgo.Session) error

	OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type BasePlugin struct{}

func (BasePlugin) Commands() []discordgo.ApplicationCommand { return nil }

func (BasePlugin) OnLoad(ds *discordgo.Session, db *bun.DB) error { return nil }

func (BasePlugin) OnShutdown(s *discordgo.Session) error { return nil }

func (BasePlugin) OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {}
