// Package output is the collection of standard outputs for use with disdup. It
// mainly implements simple, reusable output components which can either be
// directly integrated into a user-facing application, or which can be used to
// form another, dissimilar component.
package output

import (
	"github.com/bwmarrin/discordgo"
)

// A Message is a superset of the discord message object with extra information
// retrieved and managed by disdup. Although messages are passed to outputs by
// reference, it should be assumed that they are immutable.
type Message struct {
	*discordgo.Message
	PrettyContent string
	ChannelName   string
	GuildName     string
}
