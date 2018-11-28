package plugins

import (
	"fmt"
	"time"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"
)

var (
	startTime time.Time
	msg       = `Project Haleakala written by Thomas Le'Breau
	**Uptime**: %s`
)

func init() {
	startTime = time.Now()
}

type AboutPlugin struct{}

func (s AboutPlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "About this bot", s.OnMessage)
}

// OnMessage returns an about message
func (s AboutPlugin) OnMessage(ctx *muxer.Context) {
	ctx.Send(fmt.Sprintf(msg, time.Now().Sub(startTime).String()))
}

func (s AboutPlugin) Name() string {
	return "about"
}
