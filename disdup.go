package disdup

import (
	"errors"
	"fmt"

	"github.com/ethanv2/disdup/cache"
	config "github.com/ethanv2/disdup/conf"

	"github.com/bwmarrin/discordgo"
)

// Duplicator errors.
var (
	ErrClosed = errors.New("duplicator: closed")
)

type Duplicator struct {
	conn  *discordgo.Session
	cache *cache.Cache
	conf  config.Config

	cerr chan error
	stop chan struct{}
}

func NewDuplicator(conf config.Config) (Duplicator, error) {
	var err error
	dup := Duplicator{
		conf: conf,
		cerr: make(chan error),
		stop: make(chan struct{}),
	}

	dup.conn, err = discordgo.New("Bot " + conf.Token)
	if err != nil {
		return Duplicator{}, fmt.Errorf("duplicator: session creation: %w", err)
	}

	// Bot intents. Bot needs to:
	//  - Send messages
	//  - Read sent messages
	//  - All of the above in private channels/group chats
	//  - See when a guild is added, changes name etc.
	dup.conn.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentMessageContent | discordgo.IntentDirectMessages | discordgo.IntentGuilds

	// Set up cache based on current discord session
	dup.cache = cache.NewCache(dup.conn)

	// Event handling.
	// Discordgo automatically dispatches events to the correct handler
	// based on method signature.
	dup.conn.AddHandler(dup.onDisconnect)
	dup.conn.AddHandler(dup.onMessage)

	if err = dup.conn.Open(); err != nil {
		return Duplicator{}, fmt.Errorf("duplicator: connection: %w", err)
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
