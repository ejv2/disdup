package output

import (
	"errors"
	"io"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Collation modes for collecting together consecutive alike messages. Each
// larger mode includes all those below it. It is impossible to, for instance,
// collate users without collating channels. Unknown collation modes are
// ignored.
//
// Constants are simply a linear sequence which we can test against inclusively
// using >= for a certain flag.
const (
	// Collate messages sent in the same channel of the same guild together.
	WriterCollateChannel = iota + 1
	// Collate messages sent by the same user in the same channel together.
	WriterCollateUser
)

// Writer error values.
var (
	ErrNilOutput = errors.New("output writer: use with nil output")
	ErrNotOpen   = errors.New("output writer: write before open")
)

// Writer outputs messages to an io.Writer, formatted with a timestamp, author
// and channel name.
//
// If collate is non-zero, it is a bitwise combination of one or more collation
// flags. See collation flags documentation for use.
type Writer struct {
	Output io.WriteCloser
	// Prefix will be prepended to each message log.
	Prefix string
	// Collate mode. See constants for documentation.
	Collate int
	lg      *log.Logger
	// ID of the last author
	lastAuthor string
	// Id of the last sent channel
	lastChannel string
}

func (w *Writer) Open(s *discordgo.Session) error {
	if w.Output == nil {
		panic(ErrNilOutput)
	}

	w.lg = log.New(w.Output, w.Prefix, log.LstdFlags)

	// Add an extra newline to the beginning of un-collated output to make
	// format consistent with collation messages
	if w.Collate == 0 {
		w.Output.Write([]byte("\n"))
	}

	return nil
}

func (w *Writer) Write(m Message) {
	if w.lg == nil {
		panic(ErrNotOpen)
	}

	if w.Collate >= WriterCollateChannel {
		if m.ChannelID != w.lastChannel {
			msg := "\n" + m.GuildName + " #" + m.ChannelName + ":\n"
			w.Output.Write([]byte(msg))
		}

		if w.Collate >= WriterCollateUser && m.Author.ID == w.lastAuthor && m.ChannelID == w.lastChannel {
			// Length of username plus three characters padding for alignment
			// This must be updated if output format changes!
			pref := strings.Repeat(" ", len([]rune(m.Author.Username))+3)
			w.lg.Printf("%s%s", pref, m.PrettyContent)
		} else {
			w.lg.Printf("%s: %s", m.Author, m.PrettyContent)
		}
	} else {
		w.lg.Printf("%s (%s #%s): %s", m.Author, m.GuildName, m.ChannelName, m.PrettyContent)
	}

	w.lastAuthor = m.Author.ID
	w.lastChannel = m.ChannelID
}

func (w *Writer) Close() error {
	w.lg.Println("disdup log closing")
	return w.Output.Close()
}
