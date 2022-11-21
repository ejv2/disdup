package output_test

import (
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanv2/disdup/output"
)

var fakeSession = &discordgo.Session{}

var testMessages = []output.Message{
	// user1 (guild1 #chan1): Message 1
	{
		Message:       &discordgo.Message{ChannelID: "chan1.1", Author: &discordgo.User{ID: "a", Username: "user1"}},
		PrettyContent: "Message 1",
		ChannelName:   "chan1",
		GuildName:     "guild1",
	},
	// user1 (guild1 #chan2): Message 2
	{
		Message:       &discordgo.Message{ChannelID: "chan2.1", Author: &discordgo.User{ID: "a", Username: "user1"}},
		PrettyContent: "Message 2",
		ChannelName:   "chan2",
		GuildName:     "guild1",
	},
	// user1 (guild1 #chan2): Message 3
	{
		Message:       &discordgo.Message{ChannelID: "chan2.1", Author: &discordgo.User{ID: "a", Username: "user1"}},
		PrettyContent: "Message 3",
		ChannelName:   "chan2",
		GuildName:     "guild1",
	},
	// user2 (guild1 #chan2): Message 4
	{
		Message:       &discordgo.Message{ChannelID: "chan2.1", Author: &discordgo.User{ID: "b", Username: "user2"}},
		PrettyContent: "Message 4",
		ChannelName:   "chan2",
		GuildName:     "guild1",
	},
	// user1 (guild2 #chan1): Message 5
	{
		Message:       &discordgo.Message{ChannelID: "chan1.2", Author: &discordgo.User{ID: "b", Username: "user1"}},
		PrettyContent: "Message 5",
		ChannelName:   "chan1",
		GuildName:     "guild2",
	},
	// user1 (guild1 #chan1): Message 6
	{
		Message:       &discordgo.Message{ChannelID: "chan1.1", Author: &discordgo.User{ID: "b", Username: "user1"}},
		PrettyContent: "Message 6",
		ChannelName:   "chan1",
		GuildName:     "guild1",
	},
	// user1 (guild1 #chan1): Message 7
	{
		Message:       &discordgo.Message{ChannelID: "chan2.1", Author: &discordgo.User{ID: "b", Username: "user1"}},
		PrettyContent: "Message 7",
		ChannelName:   "chan2",
		GuildName:     "guild1",
	},
	// user1 (guild2 #chan1): Message 8
	{
		Message:       &discordgo.Message{ChannelID: "chan2.2", Author: &discordgo.User{ID: "b", Username: "user1"}},
		PrettyContent: "Message 8",
		ChannelName:   "chan2",
		GuildName:     "guild2",
	},
}

// WriteNopCloser is used to mask close calls to things we don't want to close.
// Write attempts after calling close discard data and return io.ErrShortWrite.
type WriteNopCloser struct {
	W io.Writer
}

func (w *WriteNopCloser) Write(p []byte) (n int, err error) {
	if w == nil {
		return 0, io.ErrShortWrite
	}
	return w.W.Write(p)
}

func (w *WriteNopCloser) Close() error {
	w.W = nil
	return nil
}
