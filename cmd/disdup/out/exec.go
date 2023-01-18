package out

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ejv2/disdup/output"
)

// Possible executor init errors.
var (
	ErrEmptyCommand = errors.New("output executor: empty command")
)

// formatArgs replaces formatting options documented in the Executor struct in
// the arguments provided. The operation is performed in place on the passed
// slice.
func formatArgs(args []string, msg output.Message) []string {
	ret := make([]string, 0, len(args))

	for _, arg := range args {
		arg = strings.ReplaceAll(arg, "{id}", msg.Message.ID)
		arg = strings.ReplaceAll(arg, "{author}", msg.Author.Username)
		arg = strings.ReplaceAll(arg, "{guild}", msg.GuildName)
		arg = strings.ReplaceAll(arg, "{guild_id}", msg.GuildID)
		arg = strings.ReplaceAll(arg, "{channel}", msg.ChannelName)
		arg = strings.ReplaceAll(arg, "{channel_id}", msg.ChannelID)
		arg = strings.ReplaceAll(arg, "{content}", msg.PrettyContent)
		arg = strings.ReplaceAll(arg, "{time}", time.Now().Format(time.RFC822))

		ret = append(ret, arg)
	}

	return ret
}

// Executor reads in all incoming messages and executes a given program with
// configurable arguments. Arguments may contain formatting directives to pass
// information about messages to the executing program.
//
// If an empty command is passed, Executor.Open returns ErrEmptyCommand.
type Executor struct {
	// Command is the path to or name of the program to execute on message
	// send.
	Command string
	// Args are command line arguments to provide to the program. Simple
	// format substitution is supported for each, replacing the following:
	//   - {id}: unique message ID
	//   - {author}: name#tag of the author of the message
	//   - {guild}: the name of the guild in which the message was sent
	//   - {channel}: the name of the channel in which the message was sent
	//   - {content}: the formatted content of the message
	//   - {time}: approximate timestamp of the message's send, formatted according to RFC822
	//
	// Arguments are guaranteed to be formatted as correct command line
	// arguments, with the same restrictions as per usual via exec.Command.
	Args []string

	procwg sync.WaitGroup
}

func (e *Executor) Open(s *discordgo.Session) error {
	if e.Command == "" {
		return ErrEmptyCommand
	}

	return nil
}

func (e *Executor) Write(m output.Message) {
	e.procwg.Add(1)
	defer e.procwg.Done()

	args := formatArgs(e.Args, m)
	cmd := exec.Command(e.Command, args...)

	// For some reason, this is overriden by default
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("executor: %s %v: command failed to execute", e.Command, args)
	}
}

func (e *Executor) Close() error {
	e.procwg.Wait()
	return nil
}
