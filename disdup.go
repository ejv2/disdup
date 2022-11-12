package disdup

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Duplicator errors.
var (
	ErrClosed = errors.New("duplicator: closed")
)

type Duplicator struct {
	conn *discordgo.Session

	cerr chan error
	stop chan struct{}
}

func NewDuplicator(token string) (Duplicator, error) {
	var dup Duplicator

	conn, err := discordgo.New("Bot " + token)
	if err != nil {
		return Duplicator{}, fmt.Errorf("duplicator: session creation: %w", err)
	}

	// Bot intents. Bot needs to:
	//  - Send messages
	//  - Read sent messages
	//  - All of the above in private channels/group chats
	conn.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentMessageContent | discordgo.IntentDirectMessages

	// Event handling.
	// Discordgo automatically dispatches events to the correct handler
	// based on method signature.
	conn.AddHandler(dup.onDisconnect)

	if err = conn.Open(); err != nil {
		return Duplicator{}, fmt.Errorf("duplicator: connection: %w", err)
	}

	dup = Duplicator{
		conn: conn,
		cerr: make(chan error),
		stop: make(chan struct{}),
	}
	return dup, nil
}

// Run runs the duplicator until an error occurs or the duplicator is
// instructed to stop.
func (d Duplicator) Run() error {
	return <-d.cerr
}

// Wait returns a channel over which a single error may be received on
// duplicator termination.
func (d Duplicator) Wait() chan error {
	return d.cerr
}

// Close terminates the duplicator. Any errors waiting to be received are
// discarded and all running goroutines terminate gracefully. It is safe to
// call Close after an error, although it is seldom necessary.
func (d Duplicator) Close() {
	select {
	case <-d.stop:
	default:
		close(d.stop)
	}
	d.conn.Close()
}

// err propagates an error to the client code, ensuring that this cannot block
// if an error was already reported. err may only block in the instance that
// the client code does not receive from the error channel correctly.
func (d Duplicator) err(err error) {
	select {
	case <-d.stop:
		return
	case d.cerr <- err:
		close(d.stop)
	}
}

// onDisconnect is the event handler for a disconnection event. Note that this
// is sent from discordgo, not remotely, so this can still handle the network
// dropping.
func (d Duplicator) onDisconnect(s *discordgo.Session, c *discordgo.Disconnect) {
	d.err(ErrClosed)
}
