package haleakala

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	Discord struct {
		Token string `yaml:"token" json:"token"`
	} `yaml:"discord" json:"discord"`
	DatabaseConfig string `json:"db_string"`
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
