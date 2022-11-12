package cache_test

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ethanv2/disdup/cache"

	"testing"
)

var (
	panicsBeforeSetup int
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

	return &discordgo.Channel{}, cache.ErrMissing
}

func (m MockProvider) User(userID string) (u *discordgo.User, err error) {
	if userID == "5678" {
		return &discordgo.User{
			ID:       "5678",
			Username: "Testing User",
		}, nil
	}

	return &discordgo.User{}, cache.ErrMissing
}

func (m MockProvider) Guild(guildID string) (st *discordgo.Guild, err error) {
	if guildID == "9101112" {
		return &discordgo.Guild{
			ID:      "9101112",
			Name:    "Testing Server",
			OwnerID: "5678",
		}, nil
	}

	return &discordgo.Guild{}, cache.ErrMissing
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

func testPanicBeforeSetup(t *testing.T) {
	defer func() {
		err := recover()
		if err == cache.ErrNoProvider {
			panicsBeforeSetup++
			return
		}
		t.Error("Unexpected panic in setup:", err)
	}()

	cache.Channel("1234")
	cache.User("5678")
	cache.Guild("9101112")

	if panicsBeforeSetup != 3 {
		t.Error("Expected 3 panics due to invalid setup, got", panicsBeforeSetup)
	}
}

// testPanicOnSetProvider tests that SetProvider successfully detects an
// invalid reassign to the provider.
func testSetProvider(t *testing.T) {
	panicked := false
	prov := MockProvider{}

	// Should *not* panic
	cache.SetProvider(prov)
	// Should *not* panic
	cache.SetProvider(prov)
	// *Should* panic
	defer func() {
		if e := recover(); e != cache.ErrResetProvider {
			t.Error("Unexpected panic when re-setting provider to same value:", e)
			return
		}

		panicked = true
	}()
	cache.SetProvider(FakeProvider{})

	if !panicked {
		t.Error("SetProvider failed to panic when setting to new value")
	}
}

// TestSetup tests SetProvider and related invariants.
func TestSetProvider(t *testing.T) {
	t.Run("PanicBeforeSetProvider", testPanicBeforeSetup)
	t.Run("SetProvider", testSetProvider)
}

func testChannel(t *testing.T) {
	c, err := cache.Channel("1234")
	if err != nil {
		t.Error("Unexpected error from channel retrieval:", err)
	}
	cexpect, _ := MockProvider{}.Channel("1234")
	if c.ID != cexpect.ID {
		t.Error("Incorrect channel returned from retrieval")
	}

	cr, ok := (*cache.ChannelCache)["1234"]
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
	(*cache.ChannelCache)["testcache"] = &testchan
	if hc, err := cache.Channel("testcache"); hc.ID != testchan.ID || err != nil {
		t.Error("Failed to hit cache for cached channel value")
	}
}

func testChannelError(t *testing.T) {
	_, err := cache.Channel("abcd")
	if err == nil {
		t.Error("Expected error from non-existent channel `abcd`")
		return
	}

	if _, ok := (*cache.ChannelCache)["abcd"]; ok {
		t.Error("Channel cache contains non-existent channel `abcd`")
	}
}

func testUser(t *testing.T) {
	u, err := cache.User("5678")
	if err != nil {
		t.Error("Unexpected error from user retrieval:", err)
	}
	uexpect, _ := MockProvider{}.User("5678")
	if u.ID != uexpect.ID {
		t.Error("Incorrect user returned from retrieval")
	}

	ur, ok := (*cache.UserCache)["5678"]
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
	(*cache.UserCache)["testcache"] = &testuser
	if hc, err := cache.User("testcache"); hc.ID != testuser.ID || err != nil {
		t.Error("Failed to hit cache for cached user value")
	}
}

func testUserError(t *testing.T) {
	_, err := cache.User("abcd")
	if err == nil {
		t.Error("Expected error from non-existent user `abcd`")
		return
	}

	if _, ok := (*cache.UserCache)["abcd"]; ok {
		t.Error("Channel cache contains non-existent user `abcd`")
	}
}

func testGuild(t *testing.T) {
	g, err := cache.Guild("9101112")
	if err != nil {
		t.Error("Unexpected error from guild retrieval:", err)
	}
	gexpect, _ := MockProvider{}.Guild("9101112")
	if g.ID != gexpect.ID {
		t.Error("Incorrect guild returned from retrieval")
	}

	gr, ok := (*cache.GuildCache)["9101112"]
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
	(*cache.GuildCache)["testcache"] = &testguild
	if hc, err := cache.Guild("testcache"); hc.ID != testguild.ID || err != nil {
		t.Error("Failed to hit cache for cached guild value")
	}
}

func testGuildError(t *testing.T) {
	_, err := cache.Guild("abcd")
	if err == nil {
		t.Error("Expected error from non-existent guild `abcd`")
		return
	}

	if _, ok := (*cache.GuildCache)["abcd"]; ok {
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
