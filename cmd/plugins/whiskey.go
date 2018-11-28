package plugins

import (
	"fmt"

	"github.com/tankbusta/haleakala"
	"github.com/tankbusta/haleakala/muxer"
)

var whiskies = []string{
	"Macallan",
	"Glenlivet",
	"Yamazaki",
	"Nikka Taketsuru",
	"Hakushu",
	"Hibiki",
	"Glenfiddich",
	"Laphroaig",
	"Glenlivet",
	"Bowmore",
	"Jameson",
	"Black Bush",
	"Glenmorangie",
	"Jim Beam",
	"Wild Turkey",
	"Knob Creek",
	"Glenfarcles",
	"Bulleit",
	"Jack Daniels",
}

var commonYears = []int{
	5,
	8,
	12,
	18,
	21,
	25,
}

type WhiskeyPlugin struct{}

func (s WhiskeyPlugin) OnMessage(ctx *muxer.Context) {
	whiskey := whiskies[r.Intn(len(whiskies))]
	year := commonYears[r.Intn(len(commonYears))]

	m := fmt.Sprintf("%s have a neat %s %d year on the house!", ctx.MentionCreator(), whiskey, year)
	ctx.Send(m)
}

func (s WhiskeyPlugin) InstallRoute(f haleakala.InstallFunc) error {
	return f(s.Name(), "Pours you a whiskey from the finest collection available", s.OnMessage)
}

func (s WhiskeyPlugin) Name() string {
	return "whiskey"
}
