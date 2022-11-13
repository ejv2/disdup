package cache

import (
	"github.com/bwmarrin/discordgo"

	"testing"
)

type MockProvider struct{}

func (m MockProvider) Channel(channelID string) (c *discordgo.Channel, err error) {
	if channelID == "1234" {
		return &discordgo.Channel{
			ID:      "1234",
			Name:    "Testing Channel",
			GuildID: "9101112",
		}, nil
	}

	return &discordgo.Channel{}, ErrMissing
}

func (m MockProvider) User(userID string) (u *discordgo.User, err error) {
	if userID == "5678" {
		return &discordgo.User{
			ID:       "5678",
			Username: "Testing User",
		}, nil
	}

	return &discordgo.User{}, ErrMissing
}

func (m MockProvider) Guild(guildID string) (st *discordgo.Guild, err error) {
	if guildID == "9101112" {
		return &discordgo.Guild{
			ID:      "9101112",
			Name:    "Testing Server",
			OwnerID: "5678",
		}, nil
	}

	return &discordgo.Guild{}, ErrMissing
}

type FakeProvider struct{}

func (m FakeProvider) Channel(channelID string) (c *discordgo.Channel, err error) {
	return &discordgo.Channel{}, nil
}

func (m FakeProvider) User(userID string) (u *discordgo.User, err error) {
	return &discordgo.User{}, nil
}

func (m FakeProvider) Guild(guildID string) (st *discordgo.Guild, err error) {
	return &discordgo.Guild{}, nil
}

func testChannel(t *testing.T) {
	provider := MockProvider{}
	cache := NewCache(provider)

	c, err := cache.Channel("1234")
	if err != nil {
		t.Error("Unexpected error from channel retrieval:", err)
	}
	cexpect, _ := provider.Channel("1234")
	if c.ID != cexpect.ID {
		t.Error("Incorrect channel returned from retrieval")
	}

	cr, ok := cache.channelCache["1234"]
	if !ok {
		t.Error("Failed to insert channel into lookup cache")
		return
	}
	if cr.ID != cexpect.ID {
		t.Error("Incorrect channel inserted into cache map")
	}

	testchan := discordgo.Channel{
		ID:   "testcache",
		Name: "test channel",
	}
	cache.channelCache["testcache"] = &testchan
	if hc, err := cache.Channel("testcache"); hc.ID != testchan.ID || err != nil {
		t.Error("Failed to hit cache for cached channel value")
	}
}

func testChannelError(t *testing.T) {
	provider := MockProvider{}
	cache := NewCache(provider)

	_, err := cache.Channel("abcd")
	if err == nil {
		t.Error("Expected error from non-existent channel `abcd`")
		return
	}

	if _, ok := cache.channelCache["abcd"]; ok {
		t.Error("Channel cache contains non-existent channel `abcd`")
	}
}

func testUser(t *testing.T) {
	provider := MockProvider{}
	cache := NewCache(provider)

	u, err := cache.User("5678")
	if err != nil {
		t.Error("Unexpected error from user retrieval:", err)
	}
	uexpect, _ := MockProvider{}.User("5678")
	if u.ID != uexpect.ID {
		t.Error("Incorrect user returned from retrieval")
	}

	ur, ok := cache.userCache["5678"]
	if !ok {
		t.Error("Failed to insert user into lookup cache")
		return
	}
	if ur.ID != uexpect.ID {
		t.Error("Incorrect user inserted into cache map")
	}

	testuser := discordgo.User{
		ID:       "testuser",
		Username: "test user",
	}
	cache.userCache["testcache"] = &testuser
	if hc, err := cache.User("testcache"); hc.ID != testuser.ID || err != nil {
		t.Error("Failed to hit cache for cached user value")
	}
}

func testUserError(t *testing.T) {
	provider := MockProvider{}
	cache := NewCache(provider)

	_, err := cache.User("abcd")
	if err == nil {
		t.Error("Expected error from non-existent user `abcd`")
		return
	}

	if _, ok := cache.userCache["abcd"]; ok {
		t.Error("Channel cache contains non-existent user `abcd`")
	}
}

func testGuild(t *testing.T) {
	provider := MockProvider{}
	cache := NewCache(provider)

	g, err := cache.Guild("9101112")
	if err != nil {
		t.Error("Unexpected error from guild retrieval:", err)
	}
	gexpect, _ := MockProvider{}.Guild("9101112")
	if g.ID != gexpect.ID {
		t.Error("Incorrect guild returned from retrieval")
	}

	gr, ok := cache.guildCache["9101112"]
	if !ok {
		t.Error("Failed to insert user into lookup cache")
		return
	}
	if gr.ID != gexpect.ID {
		t.Error("Incorrect user inserted into cache map")
	}

	testguild := discordgo.Guild{
		ID:   "testguild",
		Name: "test guild",
	}
	cache.guildCache["testcache"] = &testguild
	if hc, err := cache.Guild("testcache"); hc.ID != testguild.ID || err != nil {
		t.Error("Failed to hit cache for cached guild value")
	}
}

func testGuildError(t *testing.T) {
	provider := MockProvider{}
	cache := NewCache(provider)

	_, err := cache.Guild("abcd")
	if err == nil {
		t.Error("Expected error from non-existent guild `abcd`")
		return
	}

	if _, ok := cache.guildCache["abcd"]; ok {
		t.Error("Guild cache contains non-existent user `abcd`")
	}
}

func TestRetrieval(t *testing.T) {
	t.Run("Channel", testChannel)
	t.Run("ChannelError", testChannelError)

	t.Run("User", testUser)
	t.Run("UserError", testUserError)

	t.Run("Guild", testGuild)
	t.Run("GuildError", testGuildError)
}
