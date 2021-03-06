package haleakala

import (
	"github.com/tankbusta/haleakala/muxer"

	"github.com/sirupsen/logrus"
)

var DefaultAdminMiddleware muxer.HandlerFunc = nil

func AllowOnCertainChannels(chanids []string) muxer.HandlerFunc {
	return func(c *muxer.Context) {
		if StrSliceContainsStr(c.Message.ChannelID, chanids) {
			c.Next()
		} else {
			logrus.WithFields(logrus.Fields{
				"channel_id": c.Message.ChannelID,
				"from":       c.Message.Author.Username,
			}).Warn("Discarded command request due to incorrect channel")
			c.Abort()
		}
	}
}
