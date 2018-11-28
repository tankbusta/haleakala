package pingrelay

import (
	"errors"
	"regexp"
	"strings"
)

type (
	// IRCMessage defines the ZMQ payload from ZNC
	IRCMessage struct {
		Message string  `json:"message"`
		From    string  `json:"nick"`
		RecvAt  float64 `json:"now"`
	}

	// NCPing defines a NC. ping
	NCPing struct {
		// From indicates who sent the ping
		From string
		// ToGroup indicates where the ping was sent to
		ToGroup string
		// Ping indicates the contents of the ping
		Ping string
	}

	// pingCache keeps a record of the last ping sent by a user
	pingCache map[string]*NCPing // map[user]*NCPing
)

var (
	// XXX(tankbusta): Dont hardcode these
	errNotPing = errors.New("haleakala: not an NC. ping")
	ncdotregex = regexp.MustCompile(`^(?P<ping_from>.*?) \[(?P<group_name>.*?)\]: (?P<ping>.+)`)
)

// ParseNCMessage validates and parses an NC. ping based on the PL services code
func ParseNCMessage(ircmsg string) (*NCPing, error) {
	if !ncdotregex.MatchString(ircmsg) {
		return nil, errNotPing
	}

	match := ncdotregex.FindStringSubmatch(ircmsg)
	ping := &NCPing{}
	for i, name := range ncdotregex.SubexpNames() {
		switch name {
		case "ping_from":
			// Replace the underscore with a space
			ping.From = strings.Replace(match[i], "_", " ", -1)
		case "group_name":
			ping.ToGroup = strings.ToLower(match[i])
		case "ping":
			ping.Ping = match[i]
		}
	}
	return ping, nil
}
