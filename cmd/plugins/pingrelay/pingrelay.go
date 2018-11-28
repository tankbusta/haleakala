package pingrelay

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"

	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
	zmq "github.com/pebbe/zmq4"
)

type relayConfig struct {
	ZMQ struct {
		Address        *string `json:"address"`
		EnableSecurity bool    `json:"enable_security"`
		ServerPublic   string  `json:"curve_server_public"`
		ClientPublic   string  `json:"curve_client_public"`
		ClientPrivate  string  `json:"curve_client_private"`
	} `json:"zmq"`
	PingsFrom     []string `json:"pings_from"`     // User(s) who sends the pings
	MaxSimilarity float64  `json:"max_similarity"` // The maximum similarity between the last ping from this user and the current ping based on the levenshtein measure
	RelayMessages []struct {
		ChannelIDs []string `json:"channel_ids"` // Group of Channel IDs to broadcast to
		Group      string   `json:"group"`       // Broadcast pings from this groups to the channel ids above
	} `json:"relay_groups"`
	SpamControl struct {
		Enable       bool     `json:"enabled"`
		TriggerWords []string `json:"words"`
	} `json:"spam_control"`
}

func init() {
	haleakala.Register("znc_relay", &plugin{})
}

type plugin struct{}

func (s plugin) Initialize(cfg haleakala.PluginConfigVars, ds *discordgo.Session, node storm.Node) (haleakala.IPlugin, error) {
	var rc relayConfig

	if err := json.Unmarshal(cfg, &rc); err != nil {
		return nil, err
	}

	if rc.ZMQ.Address == nil {
		// Do we have it as an environment variable?
		zmqenv := os.Getenv("BOT_ZMQ_ENDPOINT")
		if zmqenv == "" {
			return nil, errors.New("ping relay requires that address or the environment variable BOT_ZMQ_ENDPOINT to be set")
		}
		rc.ZMQ.Address = &zmqenv
	}

	sock, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}

	relay := &zncRelayPlugin{
		ds:           ds,
		stor:         node,
		sock:         sock,
		cfg:          &rc,
		stop:         make(chan bool, 1),
		wg:           &sync.WaitGroup{},
		cl:           createLookupTable(rc),
		messageStats: make(map[string]*PingStat),
		mu:           new(sync.RWMutex),
	}

	if err := relay.Start(); err != nil {
		return nil, err
	}

	return relay, nil
}

type zncRelayPlugin struct {
	ds           *discordgo.Session
	stor         storm.Node
	sock         *zmq.Socket
	stop         chan bool
	cfg          *relayConfig
	wg           *sync.WaitGroup
	mu           *sync.RWMutex
	cl           chanMap
	messageStats map[string]*PingStat
}

func (s *zncRelayPlugin) Destroy() error {
	close(s.stop)
	s.wg.Wait()

	return s.sock.Close()
}

func (s zncRelayPlugin) SupportsUnload() bool {
	return true
}

func (s zncRelayPlugin) Name() string {
	return "znc_relay"
}

func (s *zncRelayPlugin) clearCache(c *muxer.Context) {
	s.mu.Lock()
	for keyname := range s.messageStats {
		s.messageStats[keyname] = nil
		delete(s.messageStats, keyname)
	}
	s.messageStats = make(map[string]*PingStat)
	s.mu.Unlock()

	c.Send(":white_check_mark: Ping cache has been cleared successfully. Enjoy your Spam!")
}

func (s *zncRelayPlugin) InstallRoute(f haleakala.InstallFunc) error {
	f("ping_stats", "send statistics about current relayed pings", s.GetRelayStats)
	f("ping_config", "", haleakala.DefaultAdminMiddleware, s.UpdatePingConfig)
	f("ping_clear_cache", "", haleakala.DefaultAdminMiddleware, s.clearCache)
	return nil
}
