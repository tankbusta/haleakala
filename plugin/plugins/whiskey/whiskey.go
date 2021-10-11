package whiskey

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/tankbusta/haleakala/plugin"

	"github.com/bwmarrin/discordgo"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func init() {
	p := whiskeyP{}
	plugin.Register(p.Name(), p)
}

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

type whiskeyP struct {
	plugin.BasePlugin
}

func (whiskeyP) Name() string { return "whiskey" }

func (whiskeyP) Commands() []discordgo.ApplicationCommand {
	return []discordgo.ApplicationCommand{
		{
			Name:        "whiskey",
			Description: "Gives you a shot of whiskey!",
			Type:        discordgo.ChatApplicationCommand,
		},
	}
}

func (whiskeyP) OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	whiskey := whiskies[r.Intn(len(whiskies))]
	year := commonYears[r.Intn(len(commonYears))]

	m := fmt.Sprintf("%s have a neat %s %d year on the house!", i.Member.Nick, whiskey, year)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: m,
		},
	})
}
