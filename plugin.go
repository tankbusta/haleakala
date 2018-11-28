package haleakala

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
)

var (
	pluginsMu sync.RWMutex
	plugins   = make(map[string]IPluginInitalizer)
)

// Register makes a plugin available to the system
func Register(name string, plugin IPluginInitalizer) {
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
	for name, _ := range plugins {
		out = append(out, name)
	}

	return out
}

func GetPlugin(name string) IPluginInitalizer {
	pluginsMu.Lock()
	defer pluginsMu.Unlock()

	if plug, exists := plugins[name]; exists {
		return plug
	}

	return nil
}

type PluginConfigVars json.RawMessage

type IPluginInitalizer interface {
	Initialize(PluginConfigVars, *discordgo.Session, storm.Node) (IPlugin, error)
}

// IPlugin defines an interface that plugins can use to embedd into the bot process
// that are "long running" or maintain some level of state throughout the lifecycle of the bot
type IPlugin interface {
	Destroy() error
	// SupportsUnload returns a boolean if this plugin should be allowed to be !unloaded via a command
	SupportsUnload() bool

	IBasicPlugin
}

type IBasicPlugin interface {
	// Name returns the name of the plugin
	Name() string
	// InstallRoute TODO
	InstallRoute(InstallFunc) error
}
