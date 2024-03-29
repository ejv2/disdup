package output

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	gomail "github.com/Shopify/gomail"
	"github.com/bwmarrin/discordgo"
)

// Mailer initialization errors.
var (
	ErrBadServer      = errors.New("output mailer: invalid host format: expect hostname:port")
	ErrMailConnection = errors.New("output mailer: mail server connection")
)

// Reply detection modes. Modes are more broad the higher their number is, with
// MailerReplyChannel being the most broad and MailerReplyNone being the most
// restrictive. Use of unknown modes for the replies mode will cause a panic at
// initialization time.
const (
	// No messages are detected as replies.
	MailerReplyNone = iota
	// Only messages which were discord replies are replies.
	MailerReplyReplies
	// Messages by the same user in the same channel are replies.
	MailerReplyUser
	// Messages by any user in the same channel are replies.
	MailerReplyChannel
)

// Default configuration values for the output. Some values are set to these if
// they are their zero values at the time that Open is called.
const (
	MailerDefaultSubject = "[disdup] {author} in #{channel}"
	MailerDefaultFooter  = "This email was sent by Disdup. https://github.com/ejv2/disdup"
)

// Internal implementation constants.
const (
	// The interval at which the mailer will disconnect from the server to
	// free resources.
	mailerReconnectionInterval = 30 * time.Second
	// Body format string.
	// Templates are (in order):
	//   - Preamble
	//   - Message text
	//   - Remarks
	//   - Footer
	mailerBodyFormat = `%s

%s

--------
%s
%s`
	messageIDDomain = "noreply.disdup.io"
)

// formatSubject replaces formatting options documented in the Mailer struct in
// the SubjectFormat string.
func formatSubject(format string, msg Message) string {
	out := format

	out = strings.ReplaceAll(out, "{id}", msg.Message.ID)
	out = strings.ReplaceAll(out, "{author}", msg.Author.Username)
	out = strings.ReplaceAll(out, "{guild}", msg.GuildName)
	out = strings.ReplaceAll(out, "{guild_id}", msg.GuildID)
	out = strings.ReplaceAll(out, "{channel}", msg.ChannelName)
	out = strings.ReplaceAll(out, "{channel_id}", msg.ChannelID)
	out = strings.ReplaceAll(out, "{time}", time.Now().String())

	return out
}

// formatRemarks enumerates possible remarks and appends to a remarks string
// for stamping on outgoing emails. Each remark ends in a stop and a space to
// begin the next remark.
func formatRemarks(msg Message) string {
	b := &strings.Builder{}

	if len(msg.Attachments) > 0 {
		fmt.Fprintf(b, "This message had %d attachments, which are enclosed. ", len(msg.Attachments))
	}

	if len(msg.Embeds) > 0 {
		fmt.Fprintf(b, "This message had %d embeds, which are enclosed (if available). ", len(msg.Embeds))
	}

	return b.String()
}

// generateMessageID generates an RFC compatible unique message ID which will
// be used in outgoing mail.
func generateMessageID(msgID string) string {
	return "<" + msgID + "@" + messageIDDomain + ">"
}

// messageReplies stores the last message ID from any given channel and users
// in that channel, all indexed by channel ID.
type messageReplies struct {
	LastID   string
	LastUser map[string]string
}

// outMessage is a formatted message ready to be sent by the mailer. The
// original message is also sent over the channel to permit the storage of
// reply modes.
type outMessage struct {
	orig Message
	mail *gomail.Message
}

// A MailServer is the basic configuration for an SMTP server connection.
// Minimal details are supplied, which are the minimum required to connect to
// most servers.
type MailServer struct {
	// Full server address, in the format hostname:port
	Address  string
	Username string
	Password string
}

// AddrInfo parses the host and port from the supplied address.
func (m MailServer) AddrInfo() (host string, port int, err error) {
	asegs := strings.Split(m.Address, ":")
	if len(asegs) != 2 {
		err = ErrBadServer
		return
	}
	host = asegs[0]
	if host == "" {
		err = ErrBadServer
		return
	}
	tmpport, err := strconv.ParseInt(asegs[1], 10, 32)
	if err != nil || tmpport <= 0 {
		err = ErrBadServer
		return
	}
	port = int(tmpport)

	return
}

// Mailer outputs messages by sending an email message to a recipient. Emails
// can be configured with certain headers, specific handling for attachments
// and modes for collation into threads.
//
// For some features of Mailer to work correctly, an internal state must be
// maintained. As a result, a Mailer can only be used serially. This is handled
// internally and all Mailer methods are safe for concurrent use.
type Mailer struct {
	// To whom shall we send this email? This is the full email address,
	// including domain and/or port numbers.
	To string
	// From whom shall this email be sent? This is the full email address
	// which will appear in the From field.
	From string
	// A format string for the message. If empty, MailerDefaultSubject is
	// used. If a message is a reply via the rules specified, "Re: " is
	// prepended to the subject.
	// Format options are as follows:
	//  - {id}: the message snowflake id
	//  - {author}: the username (user#tag) author of the message
	//  - {guild}: the server name in which the message was sent
	//  - {guild_id}: the server id in which the message was sent
	//  - {channel}: the channel name in which the message was sent
	//  - {channel_id}: the channel id in which the message was sent
	//  - {time}: the message timestamp, formatted in standard email format
	SubjectFormat string
	// What messages shall be detected as replies and under which
	// circumstances? See associated constants for details.
	ReplyMode uint
	// Custom headers to attach to the email message.
	CustomHeaders map[string]string
	// Custom text to prepend to the beginning of the message body.
	Preamble string
	// Custom text to append to the end of the message body after a
	// separating line.
	Footer string
	// SMTP server and authentication settings.
	Server MailServer

	cancel  chan struct{}
	outtray chan outMessage

	// After init, the below are owned by the runner goroutine
	connected bool
	replies   map[string]messageReplies
	conn      *gomail.Dialer
	snd       gomail.SendCloser
}

func (m *Mailer) send(msg *gomail.Message) {
	var err error

	if !m.connected {
		m.snd, err = m.conn.Dial()
		if err != nil {
			// Drop this mail for now; retry again later
			log.Println("email failed to send", err)
			return
		}

		m.connected = true
	}

	err = gomail.Send(m.snd, msg)
	if err != nil {
		// Drop this mail for now; retry again later
		log.Println("email failed to send", err)
		return
	}
}

// run is the main runner method of this mailer. It runs until the Close()
// method is called one the Mailer. This is run concurrently to allow for the
// maintenance of SMTP connections. run also maintains the lookup tables of
// threads and senders, which should be considered owned by this goroutine.
func (m *Mailer) run() {
	timer := time.NewTimer(mailerReconnectionInterval)
	defer timer.Stop()
	defer func() {
		if m.connected {
			m.snd.Close()
		}
	}()

	for {
		select {
		case msg := <-m.outtray:
			timer.Stop()
			m.send(msg.mail)
			m.updateReplies(msg.orig)
			timer.Reset(mailerReconnectionInterval)
		case <-timer.C:
			if !m.connected {
				panic("mailer: disconnection interval reached while disconnected")
			}
			m.snd.Close()
			m.connected = false
		case <-m.cancel:
			if m.connected {
				m.snd.Close()
			}
			return
		}
	}
}

func (m *Mailer) updateReplies(msg Message) {
	switch m.ReplyMode {
	case MailerReplyReplies:
		return
	case MailerReplyChannel:
		umap := m.replies[msg.ChannelID].LastUser
		if umap == nil {
			umap = make(map[string]string)
		}
		m.replies[msg.ChannelID] = messageReplies{msg.ID, umap}
		fallthrough
	case MailerReplyUser:
		m.replies[msg.ChannelID].LastUser[msg.Author.ID] = msg.ID
	}
}

func (m *Mailer) lookupReply(msg Message) string {
	switch m.ReplyMode {
	case MailerReplyNone:
		return ""
	case MailerReplyReplies:
		if msg.ReferencedMessage == nil {
			return ""
		}
		return msg.ReferencedMessage.ID
	case MailerReplyUser:
		return m.replies[msg.ChannelID].LastID
	case MailerReplyChannel:
		return m.replies[msg.ChannelID].LastUser[msg.Author.ID]
	default:
		panic("mailer reply: unhandled or unknown reply mode")
	}
}

func (m *Mailer) Open(s *discordgo.Session) error {
	m.cancel = make(chan struct{})
	m.outtray = make(chan outMessage)

	host, port, err := m.Server.AddrInfo()
	if err != nil {
		return fmt.Errorf("output mailer: %w", ErrMailConnection)
	}
	if m.Footer == "" {
		m.Footer = MailerDefaultFooter
	}
	if m.SubjectFormat == "" {
		m.SubjectFormat = MailerDefaultSubject
	}

	m.conn = gomail.NewDialer(host, port, m.Server.Username, m.Server.Password)
	m.conn.StartTLSPolicy = gomail.MandatoryStartTLS

	// Make an initial connection to check for errors
	snd, err := m.conn.Dial()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrMailConnection, err.Error())
	}
	m.connected = true
	m.replies = make(map[string]messageReplies)
	m.snd = snd

	go m.run()
	return nil
}

// Write formats the incoming message for email and then hands off to the
// sender to send to the server.
func (m *Mailer) Write(msg Message) {
	mail := gomail.NewMessage()
	mail.SetHeader("To", m.To)
	mail.SetHeader("From", m.From)
	mail.SetHeader("Subject", formatSubject(m.SubjectFormat, msg))
	mail.SetHeader("Message-Id", generateMessageID(msg.ID))
	for hdr, val := range m.CustomHeaders {
		mail.SetHeader(hdr, val)
	}

	mail.SetBody("text/plain", fmt.Sprintf(mailerBodyFormat, m.Preamble, msg.PrettyContent, formatRemarks(msg), m.Footer))

	for i, att := range msg.Downloads {
		mail.AttachReader(att.Filename, &msg.Downloads[i])
	}

	if reply := m.lookupReply(msg); reply != "" {
		mail.SetHeader("In-Reply-To", generateMessageID(reply))
	}

	m.outtray <- outMessage{msg, mail}
}

func (m *Mailer) Close() error {
	close(m.cancel)
	return nil
}
