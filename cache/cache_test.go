package cache

import (
	"errors"
	"strconv"
	"time"

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

	return nil, ErrMissing
}

func (m MockProvider) User(userID string) (u *discordgo.User, err error) {
	if userID == "5678" {
		return &discordgo.User{
			ID:       "5678",
			Username: "Testing User",
		}, nil
	}

	return nil, ErrMissing
}

func (m MockProvider) Guild(guildID string) (st *discordgo.Guild, err error) {
	if guildID == "9101112" {
		return &discordgo.Guild{
			ID:      "9101112",
			Name:    "Testing Server",
			OwnerID: "5678",
		}, nil
	}

	return nil, ErrMissing
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
	uexpect, _ := provider.User("5678")
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
	gexpect, _ := provider.Guild("9101112")
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

func testAttachment(t *testing.T) {
	url := "https://imgs.xkcd.com/comics/circuit_diagram.png"
	provider := MockProvider{}
	cache := NewCache(provider)

	att := &discordgo.MessageAttachment{
		ID:          "12345ABCDEF",
		URL:         url,
		ProxyURL:    url,
		Filename:    "circuit_diagram.png",
		ContentType: "image/png",
	}
	_, err := cache.Attachment(att)
	if err != nil {
		t.Fatalf("Unexpected error from known good URL: %s", err.Error())
	}

	ret, ok := cache.attachmentCache[url]
	if !ok {
		t.Errorf("Cache did not insert attachment correctly to cache map")
	}
	firstTime := ret.LastReference

	// After new reference, the current time should be later than
	cache.Attachment(att)
	nret := cache.attachmentCache[url]
	if nret.LastReference.Sub(firstTime) <= 0 {
		t.Errorf("Bad timing value after new reference\nexpect time after: %v\ngot: %v", firstTime, nret.LastReference)
	}
}

func testAttachmentFailure(t *testing.T) {
	provider := MockProvider{}
	cache := NewCache(provider)
	cases := []struct {
		URL    string
		Expect error
	}{
		{"https://example.com/notexist.png", ErrGetFailed},
		{"http://doesnotexist.gov.uk/", ErrRequest},
	}

	for _, c := range cases {
		att := &discordgo.MessageAttachment{
			ID:          "doesn't matter",
			Filename:    "who cares",
			URL:         c.URL,
			ProxyURL:    c.URL,
			ContentType: "image/png",
		}

		_, err := cache.Attachment(att)
		if err == nil {
			t.Errorf("%s: unexpected successful fetch, expected failure", c.URL)
		}
		if !errors.Is(err, c.Expect) {
			t.Errorf("%s: wrong error\nexpect: %s\ngot: %s", c.URL, c.Expect.Error(), err.Error())
		}
		if _, ok := cache.attachmentCache[c.URL]; ok {
			t.Errorf("%s: inserted into cache despite error in download", c.URL)
		}
	}
}

func TestAttachment(t *testing.T) {
	t.Run("Success", testAttachment)
	t.Run("Failure", testAttachmentFailure)
}

// Tests cleaning the cache based on last reference time.
func testCacheCleanRef(t *testing.T) {
	c := NewCache(MockProvider{})

	// Attachment referenced 24 hours in the future - will not be deleted
	c.attachmentCache["0"] = &Attachment{
		Name:          "0",
		LastReference: time.Now().Add(time.Hour * 24),
	}
	// Attachment last referenced two deletion cycles ago - *will* be deleted
	c.attachmentCache["1"] = &Attachment{
		Name:          "1",
		LastReference: time.Now().Add(-2 * AttachmentLifetime),
	}
	c.Clean()

	if _, ok := c.attachmentCache["0"]; !ok {
		t.Error("element '0' was wrongfully removed from cache")
	}
	if _, ok := c.attachmentCache["1"]; ok {
		t.Error("element '1' was wrongfully saved from removal from cache")
	}
}

// Tests cleaning the cache based on the count in the cache.
func testCacheCleanLeak(t *testing.T) {
	c := NewCache(MockProvider{})

	// 100 excess elements - should be pruned down to the prune threshold
	for i := int64(2); i < AttachmentPruneThreshold+100; i++ {
		str := strconv.FormatInt(i, 10)
		c.attachmentCache[str] = &Attachment{Name: str}
	}
	c.Clean()

	if len(c.attachmentCache) > AttachmentPruneThreshold {
		t.Errorf("expected cache to reduce size to len() = %d, got len() = %d", AttachmentPruneThreshold, len(c.attachmentCache))
	}
}

func TestCache_Clean(t *testing.T) {
	t.Run("Time", testCacheCleanRef)
	t.Run("Count", testCacheCleanLeak)
}
