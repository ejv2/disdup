package output

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrChanTimeout = errors.New("output channel: timeout on send")
	ErrChanNil     = errors.New("output channel: nil output")
)

// chanSendTimeout attempts to send on channel c, but stops after timeout has
// elapsed. If the send succeeded, nil is returned, else ErrChanTimeout.
func chanSendTimeout[T any](c chan T, val T, timeout time.Duration) error {
	var timeoutchan <-chan time.Time
	if timeout == 0 {
		timeoutchan = nil
	} else {
		timeoutchan = time.After(timeout)
	}

	select {
	case c <- val:
		return nil
	case <-timeoutchan:
		return ErrChanTimeout
	}
}

// Channel outputs formatted messages to a channel, optionally with a timeout.
// Channel closes its output channel once the output is closed.
//
// If channel is nil, Channel.Open will return an error. If Timeout is zero, no
// timeout is enforced.
type Channel struct {
	Output  chan string
	Timeout time.Duration
}

func (c *Channel) Open(s *discordgo.Session) error {
	if c.Output == nil {
		return ErrChanNil
	}

	return nil
}

func (c *Channel) Write(m Message) {
	out := fmt.Sprintf("@%s (%s) #%s: %s", m.Author.Username, m.GuildName, m.ChannelName, m.PrettyContent)
	chanSendTimeout(c.Output, out, c.Timeout)
}

func (c *Channel) Close() error {
	close(c.Output)
	return nil
}

// RawChannel outputs a raw message object to the given channel, optionally
// with a timeout. Channel closes its output channel once the output is closed.
//
// If channel is nil, Channel.Open will return an error. If TImeout is zero, no
// timeout is enforced.
type RawChannel struct {
	Output  chan Message
	Timeout time.Duration
}

func (r *RawChannel) Open(s *discordgo.Session) error {
	if r.Output == nil {
		return ErrChanNil
	}

	return nil
}

func (r *RawChannel) Write(m Message) {
	chanSendTimeout(r.Output, m, r.Timeout)
}

func (r *RawChannel) Close() error {
	close(r.Output)
	return nil
}
