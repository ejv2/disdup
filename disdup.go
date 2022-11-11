package disdup

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Duplicator struct {
	conn *discordgo.Session
}

func NewDuplicator(token string) (Duplicator, error) {
	conn, err := discordgo.New("Bot " + token)
	if err != nil {
		return Duplicator{}, fmt.Errorf("duplicator: session creation: %w", err)
	}

	conn.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentMessageContent | discordgo.IntentDirectMessages

	if err = conn.Open(); err != nil {
		return Duplicator{}, fmt.Errorf("duplicator: connection: %w", err)
	}

	return Duplicator{
		conn: conn,
	}, nil
}

func (d Duplicator) Close() {
	d.conn.Close()
}
