package seat

import (
	"fmt"
	"strings"

	"github.com/tankbusta/haleakala/muxer"

	"github.com/sirupsen/logrus"
)

func (s *seatPlugin) sendCharacterInfo(groupid int, ctx *muxer.Context) {
	toons, err := s.client.GetGroup(groupid)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"discoid": ctx.FromUserID,
			"group":   groupid,
		}).Error("Failed to lookup group in SeAT")

		ctx.SendPrivately("Sorry :( I failed to make a request to SeAT")
		return
	}

	sz := len(toons.Data.Users)
	if toons == nil || sz == 0 {
		ctx.SendPrivately("How about you add some of your EVE Accounts to SeAT :)")
		return
	}

	users := make([]string, sz)
	for i, toon := range toons.Data.Users {
		users[i] = toon.Name
	}

	ctx.SendPrivately(fmt.Sprintf("The following EVE Characters are associated to you: %s", strings.Join(users, ", ")))
}
