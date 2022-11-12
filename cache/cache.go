package cache

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

// Generic errors.
var (
	ErrMissing = errors.New("cache: entry not present")
	ErrNoProvider = errors.New("cache: attempted use before setting provider")
	ErrResetProvider = errors.New("cache: attempted to re-set provider")
)

// Singleton package state.
var (
	provider     Provider
	channelCache = make(map[string]*discordgo.Channel)
	userCache    = make(map[string]*discordgo.User)
	guildCache   = make(map[string]*discordgo.Guild)
)

// Provider is a data provider for discord users and channels. This is mainly
// for testing and is designed for use with either a mock or
// *discordgo.Session.
type Provider interface {
	Channel(channelID string) (c *discordgo.Channel, err error)
	User(userID string) (u *discordgo.User, err error)
	Guild(guildID string) (st *discordgo.Guild, err error)
}

// SetProvider sets the provider which the cache is using for data. This must
// be called before the cache is used and may only be called once. SetProvider
// panics if called more than once with different providers.
func SetProvider(p Provider) {
	if provider != nil && provider != p {
		panic(ErrResetProvider)
	}
	provider = p
}

// Channel looks up and returns a channel's data from the discord API, or
// returns the cached value if already found. If the channel could not be
// found, error is returned from the discord API. Errors are not cached and
// failed lookups cause a new API hit.
func Channel(ID string) (discordgo.Channel, error) {
	if provider == nil {
		panic(ErrNoProvider)
	}

	if c, ok := channelCache[ID]; ok {
		return *c, nil
	}

	newchan, err := provider.Channel(ID)
	if err != nil {
		return discordgo.Channel{}, err
	}

	channelCache[ID] = newchan
	return *newchan, nil
}

// User looks up and returns a user's data from the discord API, or returns the
// cached value if already found. If the user could not be found, error is
// returned from the discord API. Errors are not cached and failed lookups
// cause a new API hit.
func User(ID string) (discordgo.User, error) {
	if provider == nil {
		panic(ErrNoProvider)
	}

	if u, ok := userCache[ID]; ok {
		return *u, nil
	}

	newuser, err := provider.User(ID)
	if err != nil {
		return discordgo.User{}, err
	}

	userCache[ID] = newuser
	return *newuser, nil
}

// Guild looks up and returns a guild's data from the discord API, or returns
// the cached value if already found. If the guild could not be found, error is
// returned from the discord API. Errors are not cached and failed lookups
// cause a new API hit.
func Guild(ID string) (discordgo.Guild, error) {
	if provider == nil {
		panic(ErrNoProvider)
	}

	if g, ok := guildCache[ID]; ok {
		return *g, nil
	}

	newuser, err := provider.Guild(ID)
	if err != nil {
		return discordgo.Guild{}, err
	}

	guildCache[ID] = newuser
	return *newuser, nil
}

// InvalidateChannel invalidates the cache entry for a given channel ID.
func InvalidateChannel(ID string) error {
	if _, ok := channelCache[ID]; !ok {
		return ErrMissing
	}

	delete(channelCache, ID)
	return nil
}

// InvalidateUser invalidates the cache entry for a given user ID.
func InvalidateUser(ID string) error {
	if _, ok := userCache[ID]; !ok {
		return ErrMissing
	}

	delete(userCache, ID)
	return nil
}

// InvalidateGuild invalidates the cache entry for a given guild ID.
func InvalidateGuild(ID string) error {
	if _, ok := guildCache[ID]; !ok {
		return ErrMissing
	}

	delete(guildCache, ID)
	return nil
}
