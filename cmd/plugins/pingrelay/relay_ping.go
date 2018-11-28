package pingrelay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

const timeLayout = "01/02 15:04 MST"

var (
	characterLookupCache   = map[string]int{}
	characterLookupCacheMu = sync.Mutex{}
)

type esiCharacterSearchResp struct {
	Character []int `json:"character"`
}

// relayPing is called whenever we get a message from the relayer
func (s *zncRelayPlugin) relayPing(channelID string, ping NCPing) error {
	var pingUserEVECharacterID int
	var resp esiCharacterSearchResp

	now := time.Now().UTC().Format(timeLayout)

	// XXX(tankbusta): dont hardcode this
	if ping.From == "CCP KenZoku" {
		goto ContinuePing
	}

	characterLookupCacheMu.Lock()
	defer characterLookupCacheMu.Unlock()

	if fromCharacterID, ok := characterLookupCache[ping.From]; !ok {
		vals := url.Values{}
		vals.Add("categories", "character")
		vals.Add("datasource", "tranquility")
		vals.Add("language", "en-us")
		vals.Add("search", ping.From)
		vals.Add("strict", "true")
		vals.Add("user_agent", "Thomas Le'Breau")

		searchURL := url.URL{
			Scheme:   "https",
			Host:     ESIEndpoint,
			Path:     "/latest/search",
			RawQuery: vals.Encode(),
		}

		logrus.WithFields(logrus.Fields{
			"esi_url": searchURL.String(),
			"from":    ping.From,
			"url":     searchURL.String(),
		}).Debug("Making request to ESI to get character ID")

		res, err := http.Get(searchURL.String())
		if err != nil || res.StatusCode != 200 {
			logrus.WithFields(logrus.Fields{
				"err":  err,
				"code": res.StatusCode,
				"url":  searchURL.String(),
			}).Error("Failed to get character ID for portrait")
			goto ContinuePing
		}

		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			logrus.WithError(err).Error("Failed to decode ESI response")
			goto ContinuePing
		}

		if len(resp.Character) > 0 {
			// we'll use the first character ID and there should only be 1 anyways since we're in "strict" mode
			pingUserEVECharacterID = resp.Character[0]
			characterLookupCache[ping.From] = pingUserEVECharacterID
		}
	} else {
		pingUserEVECharacterID = fromCharacterID
	}

ContinuePing:
	author := &discordgo.MessageEmbedAuthor{
		Name: ping.From,
	}

	if pingUserEVECharacterID != 0 {
		author.IconURL = fmt.Sprintf("https://imageserver.eveonline.com/Character/%d_256.jpg", pingUserEVECharacterID)
		author.URL = fmt.Sprintf("https://zkillboard.com/character/%d/", pingUserEVECharacterID)
	}

	if _, err := s.ds.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("@everyone %s", ping.Ping),
		Embed: &discordgo.MessageEmbed{
			Author:      author,
			Color:       0x00ff00,
			Title:       "Fleet Ping",
			Description: ping.Ping,
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "EVE Time",
					Value:  now,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Group",
					Value:  ping.ToGroup,
					Inline: true,
				},
			},
		},
	}); err != nil {
		logrus.WithError(err).WithField("channel_id", channelID).Error("Failed to send ping")
	}

	return nil
}
