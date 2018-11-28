// Package muxer provides a simple Discord message route multiplexer that
// parses messages and then executes a matching registered handler, if found.
// dgMux can be used with both Disgord and the DiscordGo library.
// based on the code from https://github.com/bwmarrin/disgord/tree/master/x
package muxer

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Route holds information about a specific message route handler
type Route struct {
	Pattern     string      // match pattern that should trigger this route handler
	Description string      // short description of this route
	Help        string      // detailed help string for this route
	Run         HandlerFunc // route handler function to call
	Handlers    HandlerChain
}

// HandlerFunc is the function signature required for a message route handler.
type HandlerFunc func(*Context)

type HandlerChain []HandlerFunc

var errPrefixNotFound = errors.New("muxer: prefix not found. could not remove")

// Mux is the main struct for all dgMux methods.
type Mux struct {
	l       *sync.Mutex
	Routes  []*Route
	Default *Route
	Prefix  string
}

// New returns a new Discord message route mux
func New() *Mux {
	return &Mux{
		Prefix: "!",
		l:      &sync.Mutex{},
	}
}

// Route allows you to register a route
func (m *Mux) Route(pattern, desc string, handlers ...HandlerFunc) (*Route, error) {
	m.l.Lock()

	r := &Route{
		Pattern:     pattern,
		Description: desc,
		Handlers:    handlers,
	}
	m.Routes = append(m.Routes, r)
	m.l.Unlock()

	return r, nil
}

// RemoveRoute allows you to unregister a route from the muxer
func (m *Mux) RemoveRoute(pattern string) error {
	m.l.Lock()
	defer m.l.Unlock()

	for i, r := range m.Routes {
		if r.Pattern == pattern {
			// Go ahead and remove the route!
			copy(m.Routes[i:], m.Routes[i+1:])
			m.Routes[len(m.Routes)-1] = nil // or the zero value of T
			m.Routes = m.Routes[:len(m.Routes)-1]
			return nil
		}
	}
	return errPrefixNotFound
}

// FuzzyMatch attepts to find the best route match for a givin message.
func (m *Mux) FuzzyMatch(msg string) (*Route, []string) {
	// Tokenize the msg string into a slice of words
	fields := strings.Fields(msg)

	// no point to continue if there's no fields
	if len(fields) == 0 {
		return nil, nil
	}

	// Search though the command list for a match
	var r *Route
	var rank int
	var fk int

	for fk, fv := range fields {
		for _, rv := range m.Routes {

			// If we find an exact match, return that immediately.
			if rv.Pattern == fv {
				return rv, fields[fk:]
			}

			// Some "Fuzzy" searching...
			if strings.HasPrefix(rv.Pattern, fv) {
				if len(fv) > rank {
					r = rv
					rank = len(fv)
				}
			}
		}
	}
	return r, fields[fk:]
}

// Match returns an exact match for the route
func (m *Mux) Match(msg string) (*Route, []string) {
	// Tokenize the msg string into a slice of words
	fields := strings.Fields(msg)

	// no point to continue if there's no fields
	if len(fields) == 0 {
		return nil, nil
	}

	for _, rv := range m.Routes {
		if rv.Pattern == fields[0] {
			return rv, fields[0:]
		}
	}

	return nil, nil
}

// OnMessageCreate is a DiscordGo Event Handler function.  This must be
// registered using the DiscordGo.Session.AddHandler function.  This function
// will receive all Discord messages and parse them for matches to registered
// routes.
func (m *Mux) OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	var err error

	// Ignore all messages created by the Bot account itself
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	// Fetch the channel for this Message
	var c *discordgo.Channel
	c, err = ds.State.Channel(mc.ChannelID)
	if err != nil {
		// Try fetching via REST API
		c, err = ds.Channel(mc.ChannelID)
		if err != nil {
			log.Printf("unable to fetch Channel for Message")
			return
		}
		// Attempt to add this channel into our State
		err = ds.State.ChannelAdd(c)
		if err != nil {
			log.Printf("error updating State with Channel")
		}
	}

	ctx := &Context{
		Session:    ds,
		Message:    mc.Message,
		Content:    strings.TrimSpace(mc.Content),
		GuildID:    c.GuildID,
		IsPrivate:  c.Type == discordgo.ChannelTypeDM || c.Type == discordgo.ChannelTypeGroupDM,
		index:      -1,
		FromUserID: mc.Author.ID,
		FromUser:   mc.Author.Username,
	}

	// Detect Private Message
	if ctx.IsPrivate {
		ctx.IsDirected = true
	}

	// Detect @name or @nick mentions
	if !ctx.IsDirected {

		// Detect if Bot was @mentioned
		for _, v := range mc.Mentions {
			if v.ID == ds.State.User.ID {
				ctx.IsDirected, ctx.HasMention = true, true
				reg := regexp.MustCompile(fmt.Sprintf("<@!?(%s)>", ds.State.User.ID))

				// Was the @mention the first part of the string?
				if reg.FindStringIndex(ctx.Content)[0] == 0 {
					ctx.HasMentionFirst = true
				}

				// strip bot mention tags from content string
				ctx.Content = reg.ReplaceAllString(ctx.Content, "")
				break
			}
		}
	}

	// Detect prefix mention
	if !ctx.IsDirected && len(m.Prefix) > 0 {
		if strings.HasPrefix(ctx.Content, m.Prefix) {
			ctx.IsDirected, ctx.HasPrefix, ctx.HasMentionFirst = true, true, true
			ctx.Content = strings.TrimPrefix(ctx.Content, m.Prefix)
		}
	}

	// For now, if we're not specifically mentioned we do nothing.
	// later I might add an option for global non-mentioned command words
	if !ctx.IsDirected {
		return
	}

	// Try to find the exact command out of the message.
	r, fl := m.Match(ctx.Content)
	if r != nil {
		ctx.Fields = fl
		ctx.handlers = r.Handlers
		ctx.Next()
		return
	}

	// If no command match was found, call the default.
	// Ignore if only @mentioned in the middle of a message
	if m.Default != nil && (ctx.HasMentionFirst) {
		// TODO: This could use a ratelimit
		// or should the ratelimit be inside the cmd handler?..
		// In the case of "talking" to another bot, this can create an endless
		// loop.  Probably most common in private messages.
		m.Default.Run(ctx)
	}

}
