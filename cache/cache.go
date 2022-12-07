package cache

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// Generic errors.
var (
	ErrMissing     = errors.New("cache: entry not present")
	ErrNilProvider = errors.New("cache: attempted to create cache with nil provider")
	ErrIO          = errors.New("cache: attachment download: I/O error")
	ErrRequest     = errors.New("cache: attachment download: network request failed")
	ErrGetFailed   = errors.New("cache: attachment download: http error")
)

// Cache represents a cache of Discord API data objects.
type Cache struct {
	provider        Provider
	channelCache    map[string]*discordgo.Channel
	userCache       map[string]*discordgo.User
	guildCache      map[string]*discordgo.Guild
	attachmentCache map[string]*Attachment
}

// An Attachment is a generic representation for an attachment downloaded from
// the Discord API.
type Attachment struct {
	Name, Type string
	Content    []byte
}

// Provider is a data provider for discord users and channels. This is mainly
// for testing and is designed for use with either a mock or
// *discordgo.Session.
type Provider interface {
	Channel(channelID string) (c *discordgo.Channel, err error)
	User(userID string) (u *discordgo.User, err error)
	Guild(guildID string) (st *discordgo.Guild, err error)
}

// NewCache creates a new cache object with provider p.
func NewCache(p Provider) *Cache {
	if p == nil {
		panic(ErrNilProvider)
	}

	return &Cache{
		provider:        p,
		channelCache:    make(map[string]*discordgo.Channel),
		userCache:       make(map[string]*discordgo.User),
		guildCache:      make(map[string]*discordgo.Guild),
		attachmentCache: make(map[string]*Attachment),
	}
}

// Channel looks up and returns a channel's data from the discord API, or
// returns the cached value if already found. If the channel could not be
// found, error is returned from the discord API. Errors are not cached and
// failed lookups cause a new API hit.
func (c *Cache) Channel(ID string) (discordgo.Channel, error) {
	if ch, ok := c.channelCache[ID]; ok {
		return *ch, nil
	}

	newchan, err := c.provider.Channel(ID)
	if err != nil {
		return discordgo.Channel{}, err
	}

	c.channelCache[ID] = newchan
	return *newchan, nil
}

// User looks up and returns a user's data from the discord API, or returns the
// cached value if already found. If the user could not be found, error is
// returned from the discord API. Errors are not cached and failed lookups
// cause a new API hit.
func (c *Cache) User(ID string) (discordgo.User, error) {
	if u, ok := c.userCache[ID]; ok {
		return *u, nil
	}

	newuser, err := c.provider.User(ID)
	if err != nil {
		return discordgo.User{}, err
	}

	c.userCache[ID] = newuser
	return *newuser, nil
}

// Guild looks up and returns a guild's data from the discord API, or returns
// the cached value if already found. If the guild could not be found, error is
// returned from the discord API. Errors are not cached and failed lookups
// cause a new API hit.
func (c *Cache) Guild(ID string) (discordgo.Guild, error) {
	if g, ok := c.guildCache[ID]; ok {
		return *g, nil
	}

	newuser, err := c.provider.Guild(ID)
	if err != nil {
		return discordgo.Guild{}, err
	}

	c.guildCache[ID] = newuser
	return *newuser, nil
}

// Attachment looks up and returns the content and info for a remote attachment
// from the Discord API. Lookups from the same url are guaranteed not to cause
// an API hit. Errors are not cached and the attachment is assumed to not
// exist.
func (c *Cache) Attachment(at *discordgo.MessageAttachment) (Attachment, error) {
	if a, ok := c.attachmentCache[at.URL]; ok {
		return *a, nil
	}

	ret := Attachment{
		Name: at.Filename,
		Type: at.ContentType,
	}

	r, err := http.Get(at.URL)
	if err != nil {
		return ret, fmt.Errorf("%w: %s", ErrRequest, err.Error())
	}
	if r.StatusCode != 200 {
		return ret, ErrGetFailed
	}
	defer r.Body.Close()

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return ret, fmt.Errorf("%w: %s", ErrIO, err.Error())
	}
	ret.Content = buf

	c.attachmentCache[at.URL] = &ret
	return ret, nil
}

// InvalidateChannel invalidates the cache entry for a given channel ID.
func (c *Cache) InvalidateChannel(ID string) error {
	if _, ok := c.channelCache[ID]; !ok {
		return ErrMissing
	}

	delete(c.channelCache, ID)
	return nil
}

// InvalidateUser invalidates the cache entry for a given user ID.
func (c *Cache) InvalidateUser(ID string) error {
	if _, ok := c.userCache[ID]; !ok {
		return ErrMissing
	}

	delete(c.userCache, ID)
	return nil
}

// InvalidateGuild invalidates the cache entry for a given guild ID.
func (c *Cache) InvalidateGuild(ID string) error {
	if _, ok := c.guildCache[ID]; !ok {
		return ErrMissing
	}

	delete(c.guildCache, ID)
	return nil
}
