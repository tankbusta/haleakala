package haleakala

import (
	"encoding/json"
	"io/ioutil"
)

type PluginConfig struct {
	PluginName string          `yaml:"name" json:"name"`
	Enabled    bool            `yaml:"enabled" json:"enabled"`
	Config     json.RawMessage `yaml:"config" json:"config"`
}

type config struct {
	Discord struct {
		Token         string   `yaml:"token" json:"token"`
		AdminChannels []string `yaml:"admin_channels" json:"admin_channels"`
	} `yaml:"discord" json:"discord"`
	Plugins []PluginConfig `yaml:"plugins" json:"plugins"`
}

func loadconfig(path string) (*config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	out := &config{}
	if err := json.Unmarshal(b, out); err != nil {
		return nil, err
	}

	return out, nil
}
