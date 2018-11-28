package pingrelay_forwarder

import (
	"encoding/json"

	"github.com/tankbusta/haleakala"

	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
)

func init() {
	haleakala.Register("znc_relay_forwarder", &forwarderPlugin{})
}

type forwarderPlugin struct{}

func (s forwarderPlugin) Initialize(cfg haleakala.PluginConfigVars, _ *discordgo.Session, _ storm.Node) (haleakala.IPlugin, error) {
	var rc Config

	if err := json.Unmarshal(cfg, &rc); err != nil {
		return nil, err
	}

	forwarder, err := New(rc)
	if err != nil {
		return nil, err
	}

	if err := forwarder.Start(); err != nil {
		return nil, err
	}

	return &pingRelayForwarderPlugin{
		gw: forwarder,
	}, nil
}

type pingRelayForwarderPlugin struct {
	gw *Context
}

func (s *pingRelayForwarderPlugin) Destroy() error {
	if s.gw != nil {
		e := s.gw.Stop()
		s.gw = nil
		return e
	}
	return nil
}

func (s pingRelayForwarderPlugin) SupportsUnload() bool {
	return true
}

func (s pingRelayForwarderPlugin) Name() string {
	return "znc_relay_forwarder"
}

func (s *pingRelayForwarderPlugin) InstallRoute(f haleakala.InstallFunc) error {
	// dont need to install anything here
	return nil
}
