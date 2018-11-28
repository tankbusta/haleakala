package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"

	"github.com/sirupsen/logrus"
)

var (
	baseURL   = "https://esi.tech.ccp.is/latest/status/?datasource=%s"
	esiClient = &http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
)

type StatusInto struct {
	StartTime     time.Time `json:"start_time"`
	Players       int       `json:"players"`
	ServerVersion string    `json:"server_version"`
}

type EVEStatusPlugin struct{}

func (s EVEStatusPlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "Return the status of the EVE cluster", s.OnMessage)
}

func (s EVEStatusPlugin) Name() string {
	return "evestatus"
}

// OnMessage returns the status of the cluster
func (s EVEStatusPlugin) OnMessage(ctx *muxer.Context) {
	var resp *http.Response
	var info StatusInto
	var server string

	if len(ctx.Fields) == 2 {
		server = ctx.Fields[1]
	}

	switch strings.ToLower(server) {
	case "tranquility", "singularity":
		break
	case "sisi":
		server = "singularity"
	case "tq":
		fallthrough
	default:
		server = "tranquility"
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(baseURL, server), nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"server": server,
			"error":  err,
		}).Error("Error creating EVE status request")
		goto FailedToFetch
	}
	req.Header.Add("User-Agent", "Thomas Le'Breau in EVE Slack")

	if resp, err = http.DefaultClient.Do(req); err != nil {
		logrus.WithError(err).Error("Error making request for EVE Status")
		goto FailedToFetch
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		logrus.WithError(err).Error("Error decoding body of EVE status")
		goto FailedToFetch
	}

	ctx.Send(fmt.Sprintf(":white_check_mark: **Online**\n**Players:** %d\n**Version:** %s\n", info.Players, info.ServerVersion))
	return

FailedToFetch:
	ctx.Send(":exclamation: Offline")
	return
}
