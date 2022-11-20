package output_test

import (
	"io"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanv2/disdup/output"

	"testing"
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

var expectedCollationOutputs = []string{
	`
user1# (guild1 #chan1): Message 1
user1# (guild1 #chan2): Message 2
user1# (guild1 #chan2): Message 3
user2# (guild1 #chan2): Message 4
user1# (guild2 #chan1): Message 5
user1# (guild1 #chan1): Message 6
user1# (guild1 #chan2): Message 7
user1# (guild2 #chan2): Message 8
`,
	`
guild1 #chan1:
user1#: Message 1

guild1 #chan2:
user1#: Message 2
user1#: Message 3
user2#: Message 4

guild2 #chan1:
user1#: Message 5

guild1 #chan1:
user1#: Message 6

guild1 #chan2:
user1#: Message 7

guild2 #chan2:
user1#: Message 8
`,
	`
guild1 #chan1:
user1#: Message 1

guild1 #chan2:
user1#: Message 2
        Message 3
user2#: Message 4

guild2 #chan1:
user1#: Message 5

guild1 #chan1:
user1#: Message 6

guild1 #chan2:
user1#: Message 7

guild2 #chan2:
user1#: Message 8
`,
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

func testWriterOpenNil(t *testing.T) {
	var panicked bool

	defer func() {
		if !panicked {
			t.Error("Expected panic due to nil output")
		}
	}()
	defer func() {
		if err := recover(); err != output.ErrNilOutput {
			t.Error("Got wrong wrong error. Expected ErrNilOutput, got:", err)
			return
		}

		panicked = true
	}()

	w := output.Writer{Output: nil}
	w.Open(fakeSession)
}

func testWriterOpen(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Error("Unexpected panic from open:", err)
		}
	}()

	w := output.Writer{
		Output:  &WriteNopCloser{os.Stdout},
		Collate: 0,
	}
	w.Open(fakeSession)
}

func TestWriter_Open(t *testing.T) {
	t.Run("NilOutput", testWriterOpenNil)
	t.Run("Normal", testWriterOpen)
}

func testWrite(t *testing.T, collate, index int) {
	str := &strings.Builder{}
	wr := &WriteNopCloser{str}
	w := output.Writer{
		Output:  wr,
		Collate: collate,
	}
	w.Open(fakeSession)

	for _, msg := range testMessages {
		w.Write(msg)
	}

	lines, expect := strings.Split(str.String(), "\n"), strings.Split(expectedCollationOutputs[index], "\n")
	ind := 0
	for _, line := range lines {
		if !strings.HasSuffix(line, expect[ind]) {
			t.Errorf("Invalid write output (line %d)\nExpect:\n%s\n\nGot:\n%s\n", ind, expect[ind], line)
			return
		}
		ind++
	}
}

func TestWriter_Write(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		testWrite(t, 0, 0)
	})
	t.Run("CollateChannel", func(t *testing.T) {
		testWrite(t, output.WriterCollateChannel, 1)
	})
	t.Run("CollateUser", func(t *testing.T) {
		testWrite(t, output.WriterCollateUser, 2)
	})
}

func TestWriter_Close(t *testing.T) {
	str := &strings.Builder{}
	wr := &WriteNopCloser{str}
	out := output.Writer{Output: wr}

	if err := out.Open(fakeSession); err != nil {
		t.Fatal("Unexpected open error:", err)
	}

	if err := out.Close(); err != nil {
		t.Fatal("Unexpected close error:", err)
	}

	if !strings.Contains(str.String(), "disdup log closing") {
		t.Errorf("Didn't get close message, got: \"%s\"", str.String())
	}
}
