// Package disdup implements a simple and programmable Discord message bouncer.
// It can be configured to duplicate messages from only certain guilds,
// channels or users and can convert messages to a variety of different
// formats.
package disdup

import (
	"errors"
	"fmt"
	"log"

	"github.com/ethanv2/disdup/cache"
	config "github.com/ethanv2/disdup/conf"
	"github.com/ethanv2/disdup/output"

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

// NewDuplicator initializes and starts running a new duplicator. As soon as
// this call completes, the duplicator is connected to Discord and serving
// events.
//
// NOTE: This call returns asynchronously. To wait for the duplicator to
// complete, use Duplicator.Run or Duplicator.Wait. It is the caller's
// responsibility to call close and to check for errors from the runner
// channel.
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
	dup.conn.AddHandler(dup.onMessage)
	dup.conn.AddHandler(dup.onJoin)

	if err = dup.conn.Open(); err != nil {
		return Duplicator{}, fmt.Errorf("duplicator: connection: %w", err)
	}

	// Open up outputs
	done, fail := make(chan struct{}, len(conf.Outputs)), make(chan error, 1)
	for _, output := range conf.Outputs {
		go func(out config.OutputConfig) {
			err := out.Output.Open(dup.conn)
			if err != nil {
				select {
				case fail <- err:
				default:
				}
			}

			done <- struct{}{}
		}(output)
	}
	for i := 0; i < cap(done); i++ {
		select {
		case err := <-fail:
			return Duplicator{}, fmt.Errorf("duplicator: output open: %w", err)
		case <-done:
		}
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

// updateNickname attempts to change the nickname of the bot in the guild `g`.
func (d Duplicator) updateNickname(g *discordgo.Guild) error {
	return d.conn.GuildMemberNickname(g.ID, "@me", d.conf.Name)
}

// onMessage is the event handler for a message creation event in any of the
// guilds of which the bot is a member.
func (d Duplicator) onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	c, err := d.cache.Channel(m.ChannelID)
	if err != nil {
		log.Println("[WARNING]: duplicator: onmessage: invalid channel:", err)
		return
	}
	g, err := d.cache.Guild(m.GuildID)
	if err != nil {
		log.Println("[WARNING]: duplicator: onmessage: invalid guild:", err)
		return
	}
	cont, err := m.ContentWithMoreMentionsReplaced(s)
	if err != nil {
		log.Println("[WARNING]: duplicator: onmessage: invalid message:", err)
		// Call stops on first invalid reference found, so some of the
		// message will be valid, so we should continue
	}

	if d.conf.MessageMatches(config.MessageMatcher{
		Author:  *m.Author,
		Channel: c,
		Guild:   g,
	}) {
		msg := output.Message{
			Message:       m.Message,
			PrettyContent: cont,
			ChannelName:   c.Name,
			GuildName:     g.Name,
		}

		for _, att := range m.Attachments {
			a, err := d.cache.Attachment(att)
			if err != nil {
				log.Println("[WARNING]: duplicator: attachment download failed:", err)
				continue
			}

			msg.Downloads = append(msg.Downloads, output.Attachment{
				Filename: a.Name,
				Type:     a.Type,
				Content:  a.Content,
			})
		}

		gconf := d.conf.FindGuild(m.GuildID, g.Name)
		for _, o := range d.conf.Outputs {
			go func(out config.OutputConfig) {
				// An empty output array means unconditionally output
				if len(gconf.Output) == 0 {
					out.Output.Write(msg)
					return
				}

				for _, name := range gconf.Output {
					if out.Name == name {
						out.Output.Write(msg)
					}
				}
			}(o)
		}
		// log.Printf("@%s in #%s: %s", msg.Author.Username, msg.ChannelName, msg.Content)
	}
}

// onJoin is the event handler for when the bot is added to a guild.
func (d Duplicator) onJoin(s *discordgo.Session, c *discordgo.GuildCreate) {
	if err := d.updateNickname(c.Guild); err != nil {
		d.err(err)
	}
}
