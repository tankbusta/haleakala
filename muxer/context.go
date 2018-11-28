package muxer

import (
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"
)

type Fields []string

func (s Fields) Get(key string) string {
	for _, f := range s {
		if f == key {
			return f
		}
	}
	return ""
}

func (s Fields) GetDefault(key, defaultValue string) string {
	if out := s.Get(key); out != "" {
		return out
	}
	return defaultValue
}

// Context holds a bit of extra data we pass along to route handlers
// This way processing some of this only needs to happen once.
type Context struct {
	Session *discordgo.Session
	Message *discordgo.Message

	Fields          Fields
	Content         string
	GuildID         string
	IsDirected      bool
	IsPrivate       bool
	HasPrefix       bool
	HasMention      bool
	HasMentionFirst bool
	FromUserID      string
	FromUser        string

	index    int8
	handlers HandlerChain
}

func (s *Context) MentionCreator() string {
	return fmt.Sprintf("<@%s>", s.Message.Author.ID)
}

func (s *Context) Next() {
	s.index++
	l := int8(len(s.handlers))
	for ; s.index < l; s.index++ {
		s.handlers[s.index](s)
	}
}

func (s *Context) Abort() {
	s.index = math.MaxInt8 / 2
}

func (s *Context) Send(message string) error {
	_, err := s.Session.ChannelMessageSend(s.Message.ChannelID, message)
	return err
}

func (s *Context) SendPrivately(message string) error {
	pmc, err := s.Session.UserChannelCreate(s.Message.Author.ID)
	if err != nil {
		// couldnt create message :(
		s.Send(fmt.Sprintf("Sorry %s, I tried to send you a PM but i couldn't", s.Message.Author.Username))
		return err
	}

	_, err = s.Session.ChannelMessageSend(pmc.ID, message)
	return err
}
