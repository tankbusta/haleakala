package haleakala

import (
	"fmt"

	"github.com/asdine/storm"
)

func initStorage(cfg DatabaseConfig) (*storm.DB, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("database.path must not be empty")
	}

	return storm.Open(cfg.Path, storm.BoltOptions(0600, nil))
}

func (s *Context) createBucketForPlugin(pluginName string) storm.Node {
	return s.stor.From("plugins", pluginName)
}
