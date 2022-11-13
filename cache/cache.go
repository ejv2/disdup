package cache

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

// Generic errors.
var (
	ErrMissing     = errors.New("cache: entry not present")
	ErrNilProvider = errors.New("cache: attempted to create cache with nil provider")
)

// Cache represents a cache of Discord API data objects.
type Cache struct {
	provider     Provider
	channelCache map[string]*discordgo.Channel
	userCache    map[string]*discordgo.User
	guildCache   map[string]*discordgo.Guild
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
		provider:     p,
		channelCache: make(map[string]*discordgo.Channel),
		userCache:    make(map[string]*discordgo.User),
		guildCache:   make(map[string]*discordgo.Guild),
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
