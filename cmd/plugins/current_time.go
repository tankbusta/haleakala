package plugins

import (
	"fmt"
	"time"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"
)

const timeLayout = "01/02 15:04 MST"

type CurrentTimePlugin struct{}

func (s CurrentTimePlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "Return UTC (EVE) time", s.OnMessage)
}

// GetCurrentTime simply returns the current time of the server
func (s CurrentTimePlugin) OnMessage(ctx *muxer.Context) {
	ctx.Send(fmt.Sprintf("The time is %s", time.Now().UTC().Format(timeLayout)))
}

func (s CurrentTimePlugin) Name() string {
	return "time"
}
