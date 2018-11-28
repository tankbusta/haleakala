package pingrelay

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/tankbusta/haleakala/muxer"

	"github.com/agext/levenshtein"
	zmq "github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
)

const (
	defaultMaxPingSimilarity = 0.75
	moarMessages             = zmq.Errno(syscall.EAGAIN)
	ESIEndpoint              = "esi.tech.ccp.is"
)

type (
	OnPingMessage func(channelID string, ping NCPing) error

	PingStat struct {
		TotalPings    int
		TotalDiscards int
		LastMessage   string
	}

	// unexported fields below
	chanMap map[string][]string // map[groupName][]DiscordChannelIDs
)

func (s chanMap) VisitGroup(groupName string) []string {
	if _, ok := s[groupName]; ok {
		return s[groupName]
	}

	return []string(nil)
}

func createLookupTable(cfg relayConfig) chanMap {
	m := make(chanMap)
	for _, topic := range cfg.RelayMessages {
		if _, ok := m[topic.Group]; !ok {
			m[topic.Group] = make([]string, 0)
		}

		for _, channel := range topic.ChannelIDs {
			m[topic.Group] = append(m[topic.Group], channel)
		}
	}
	return m
}

func (s *zncRelayPlugin) loop() {
	defer s.wg.Done()

	var b []byte
	var err error

	s.sock.SetRcvtimeo(1 * time.Second)
Loop:
	for {
		select {
		case _ = <-s.stop:
			break Loop
		default:
		}

		b, err = s.sock.RecvBytes(0)
		if err != nil {
			// EAGAIN/moarMessages means the recv timeout has been reached
			if err == moarMessages {
				continue
			}
			logrus.WithError(err).Error("failed to read messages from relay socket")
			continue
		}
		s.recvmsg(b)
	}
	logrus.Warn("ZNC -> ZMQ loop shutting down")
}

func (s *zncRelayPlugin) recvmsg(b []byte) {
	var msg IRCMessage
	if err := json.Unmarshal(b, &msg); err != nil {
		logrus.WithError(err).Error("Failed to decode message from relay socket")
		return
	}

	logrus.WithFields(logrus.Fields{
		"msg":  msg.Message,
		"from": msg.From,
		"recv": msg.RecvAt,
	}).Debug("Handling Message from Remote EP")

	// Only relay messages from "trusted" users
	if !StrSliceContainsStr(msg.From, s.cfg.PingsFrom) {
		logrus.WithFields(logrus.Fields{
			"from": msg.From,
			"msg":  msg.Message,
		}).Warn("Ping blocked. User is not trusted")
		return
	}

	// Is this an NC. Ping?
	presult, err := ParseNCMessage(msg.Message)
	// doesnt match the regex :(
	if err != nil || presult == nil {
		return
	}

	s.mu.Lock()
	if _, ok := s.messageStats[presult.From]; !ok {
		s.messageStats[presult.From] = &PingStat{}
	}
	s.mu.Unlock()

	userStats := s.messageStats[presult.From]
	// Check and see if it's similar to our last ping
	levSimilarity := levenshtein.Similarity(userStats.LastMessage, presult.Ping, nil)
	if levSimilarity >= s.cfg.MaxSimilarity {
		userStats.TotalDiscards++
		logrus.WithFields(logrus.Fields{
			"ping":       presult.Ping,
			"from":       presult.From,
			"similarity": levSimilarity,
		}).Warn("Discarding message as it is too similar")
		return
	}
	userStats.LastMessage = presult.Ping
	userStats.TotalPings++

	if s.cfg.SpamControl.Enable {
		if foundSpam := checkMessageForSpam(presult.Ping, s.cfg.SpamControl.TriggerWords); foundSpam {
			userStats.TotalDiscards++
			return
		}
	}

	for _, channel := range s.cl.VisitGroup(presult.ToGroup) {
		s.relayPing(channel, *presult)
	}
}

func (s *zncRelayPlugin) GetRelayStats(c *muxer.Context) {
	var out strings.Builder

	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.messageStats) == 0 {
		c.Send("No pings have been received!")
		return
	}

	out.WriteString("Ping Relay Statistics: \n\n```")
	for user, stats := range s.messageStats {
		out.WriteString(fmt.Sprintf("* %s sent %d pings and discarded %d\n", user, stats.TotalPings, stats.TotalDiscards))
	}
	out.WriteString("```")

	c.Send(out.String())
}

func (s *zncRelayPlugin) UpdatePingConfig(c *muxer.Context) {
	if len(c.Fields) != 4 {
		c.Send("!ping_config <command> <nick> <value>")
		return
	}

	// We only want the fields past our command
	command, nick, newvalue := c.Fields[1], c.Fields[2], c.Fields[3]
	fmt.Println(command, nick, newvalue)

}

// Start the ZMQ listener and listen for pings
func (s *zncRelayPlugin) Start() error {
	if s.cfg.ZMQ.EnableSecurity {
		zmq.AuthSetVerbose(true)
		if err := s.sock.ClientAuthCurve(s.cfg.ZMQ.ServerPublic,
			s.cfg.ZMQ.ClientPublic,
			s.cfg.ZMQ.ClientPrivate); err != nil {
			return err
		}
	}

	if err := s.sock.SetSubscribe(""); err != nil {
		return err
	}

	logrus.Infof("Connecting to ZMQ endpoint at %s", *s.cfg.ZMQ.Address)

	if err := s.sock.Connect(*s.cfg.ZMQ.Address); err != nil {
		return err
	}

	logrus.Debug("IRC Ping Relay connected to ZNC/Forwarder")

	s.wg.Add(1)
	go s.loop()
	return nil
}
