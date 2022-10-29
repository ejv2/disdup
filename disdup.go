package disdup

import (
	_ "github.com/bwmarrin/discordgo"
)

type Duplicator struct {
	placeholder bool
}

func NewDuplicator() Duplicator {
	return Duplicator{true}
}
