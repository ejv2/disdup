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

	return Duplicator{
		conn: conn,
	}, nil
}

func (d Duplicator) Close() {
	d.conn.Close()
}
