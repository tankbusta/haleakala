package plugin

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// DevChatID is the discord channel ID where dev bot messages are posted
//
// TODO: Make this configurable
const DevChatID = "896809463141511212"

func DiscordLogEvent(ds *discordgo.Session, msg string, args ...interface{}) error {
	_, err := ds.ChannelMessageSend(DevChatID, fmt.Sprintf(msg, args...))
	return err
}

func GetUnderlyingUserIdentifier(i *discordgo.Interaction) string {
	// If the interaction is coming from a DM, i.User will be populated
	// according to Discord/Discordgo documentation
	if i.User != nil {
		return i.User.Username
	}

	// If the interaction is coming from a guild chat, i.Member will be populated
	if i.Member != nil {
		// The nickname of the user as told by their specific guild server-profile
		if i.Member.Nick != "" {
			return i.Member.Nick
		}

		return i.Member.User.Username
	}

	// Who sent us this message??
	return ""
}
