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
