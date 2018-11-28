package haleakala

import (
	"encoding/json"
	"io/ioutil"
)

type statusChange struct {
	Statuses    []string `yaml:"statuses" json:"statuses"`
	EverySecond int      `yaml:"every_seconds" json:"every_seconds"`
}

type PluginConfig struct {
	PluginName string          `yaml:"name" json:"name"`
	Enabled    bool            `yaml:"enabled" json:"enabled"`
	Config     json.RawMessage `yaml:"config" json:"config"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

type config struct {
	Discord struct {
		Username      string       `yaml:"username" json:"username"`
		Password      string       `yaml:"password" json:"password"`
		Token         string       `yaml:"token" json:"token"`
		StatusChanger statusChange `yaml:"status" json:"status"`
		AdminChannels []string     `yaml:"admin_channels" json:"admin_channels"`
	} `yaml:"discord" json:"discord"`
	Plugins        []PluginConfig `yaml:"plugins" json:"plugins"`
	DatabaseConfig DatabaseConfig `json:"database"`
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
