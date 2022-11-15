package config

import (
	"github.com/bwmarrin/discordgo"
)

// MessageMatcher is a representation of a message better suited to matching
// against a config. It is used for the MessageMatches function.
type MessageMatcher struct {
	Author  discordgo.User
	Guild   discordgo.Guild
	Channel discordgo.Channel
}

// MessageMatches returns true if a message matches the criteria in Config,
// else returns false. A special MessageMatcher object is used to pass message
// info such that client code can do any lookups that it wishes, rather than
// passing the whole message and requiring lookups to happen here.
func (c Config) MessageMatches(match MessageMatcher) bool {
	// Guild checks
	g, ok := c.Guilds[match.Guild.ID]
	if !ok {
		g, ok = c.Guilds[match.Guild.Name]
		if !ok {
			return false
		}
	}
	if g.Disable {
		return false
	}

	// Channel checks
	if len(g.EnabledChannels) > 0 {
		found := false
		for _, elem := range g.EnabledChannels {
			if elem == match.Channel.ID || elem == match.Channel.Name {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	// User checks
	if len(g.EnabledUsers) > 0 {
		found := false
		for _, elem := range g.EnabledUsers {
			if elem == match.Author.ID || elem == match.Author.Username {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}
