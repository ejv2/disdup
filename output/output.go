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

// An Output is a destination for messages from Disdup. It has a very similar
// interface to os.File and io.ReadCloser, mainly for familiarity with existing
// APIs.
//
// Open is called at duplicator startup once. It is called concurrently with
// other outputs, so it may not duplicate state. It is responsible for
// initialising the state of the output. If error returned is not nil, disdup
// startup is aborted and the error is propagated to the client.
//
// Write is called whenever a matching incoming message event is received. For
// more information on available information, see the documentation for the
// Message struct. You are free to do any operation in Write, but it is best
// not to block for too long. as no new message events can be processed until
// all outputs for the current one have completed.
//
// Close is called exactly once upon the dropping of the output by disdup. If
// it throws an error, the rest of the close callbacks will be called before
// the error is propagated to the client code.
type Output interface {
	Open(s *discordgo.Session) error
	Write(m Message)
	Close() error
}
