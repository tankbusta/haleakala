package haleakala

import (
	"encoding/json"
	"fmt"

	"github.com/tankbusta/haleakala/muxer"
)

func (s *Context) ListLoadedPlugins(ctx *muxer.Context) {
	s.rwmu.RLock()
	defer s.rwmu.RUnlock()

	var hasPluginsThatCanBeUnloaded bool = false

	out := "The following plugins are currently loaded: \n```"
	for _, plug := range s.plugins {
		if _, ok := plug.(IPlugin); ok {
			out += fmt.Sprintf("* %s\n", plug.Name())
			hasPluginsThatCanBeUnloaded = true
		}
	}
	out += "```\n To unload a module, use !unload <name>"

	if hasPluginsThatCanBeUnloaded {
		ctx.Send(out)
	} else {
		ctx.Send("You dont have any plugins!")
	}
}

func (s *Context) UnloadPlugin(ctx *muxer.Context) {
	if len(ctx.Fields) != 2 {
		ctx.Send("Usage: !unload <plugin name>")
		return
	}
	pluginToUnload := ctx.Fields[1]

	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	// let's stop the plugin and unload any resources it may be using
	if p, ok := s.plugins[pluginToUnload]; ok {
		if advp, ok := p.(IPlugin); ok {
			if !advp.SupportsUnload() {
				ctx.Send(fmt.Sprintf(":octagonal_sign: the plugin `%s` does not support unloading via command", pluginToUnload))
				return
			}

			if err := advp.Destroy(); err != nil {
				ctx.Send(fmt.Sprintf(":octagonal_sign: the plugin `%s` failed to destroy", pluginToUnload))
			}
		}
		delete(s.plugins, p.Name())
	}

	// This can return an error, but it's only for non-existance
	s.mux.RemoveRoute(pluginToUnload)

	ctx.Send(fmt.Sprintf(":white_check_mark: plugin `%s` has been unloaded", pluginToUnload))
}

func (s *Context) LoadPlugin(ctx *muxer.Context) {
	if len(ctx.Fields) != 2 {
		ctx.Send("Usage: !load <plugin name>")
		return
	}
	pluginToLoad := ctx.Fields[1]

	// At this point, we dont need to register this plugin... we just need to call it's Initialize
	// function again and add it back to the list of loaded ones
	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	// Is this plugin already loaded?
	if _, ok := s.plugins[pluginToLoad]; ok {
		ctx.Send(fmt.Sprintf(":octagonal_sign: the plugin `%s` cannot be loaded because it already exists", pluginToLoad))
		return
	}

	pluginInit := GetPlugin(pluginToLoad)
	if pluginInit == nil {
		ctx.Send(fmt.Sprintf(":octagonal_sign: %s does not exist", pluginToLoad))
		return
	}

	var cfg json.RawMessage
	// Find our configuration for this plugin
	for _, plcfg := range s.cfg.Plugins {
		if plcfg.PluginName == pluginToLoad {
			cfg = plcfg.Config

			if !plcfg.Enabled {
				// logrus.WithFields(logrus.Fields{
				// 	"plugin": pluginToLoad,
				// }).Warn("Cannot load a disabled plugin")
				ctx.Send(fmt.Sprintf(":octagonal_sign: %s is turned off by the server admin", pluginToLoad))
				return
			}
		}
	}

	plugin, err := pluginInit.Initialize(PluginConfigVars(cfg), s.ds, s.createBucketForPlugin(pluginToLoad))
	if err != nil {
		// logrus.WithError(err).Error("An error occured while initializing the plugin", pluginToLoad)
		ctx.Send(fmt.Sprintf(":octagonal_sign: %s", err))
		return
	}

	if err := plugin.InstallRoute(s.InstallRoute); err != nil {
		// logrus.WithFields(logrus.Fields{
		// 	"plugin": plugin.Name(),
		// 	"error":  err,
		// }).Warn("Plugin could not be initialized properly")
		ctx.Send(fmt.Sprintf(":octagonal_sign: %s", err))
		return
	}

	// Keep track that this is loaded
	s.plugins[plugin.Name()] = plugin

	ctx.Send(fmt.Sprintf(":white_check_mark: plugin `%s` has been loaded!", pluginToLoad))
}
