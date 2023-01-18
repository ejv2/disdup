package config_test

import (
	"github.com/bwmarrin/discordgo"
	config "github.com/ejv2/disdup/conf"

	"testing"
)

type Test struct {
	Name    string
	Expects []bool
	Config  config.Config
}

var (
	TestData = []Test{
		{"Zero value", []bool{false, false, false}, config.Config{Guilds: map[string]config.GuildConfig{}}},
		{"Disable", []bool{true, true, false}, config.Config{Guilds: map[string]config.GuildConfig{
			"a": {},
			"b": {},
			"c": {Disable: true},
		}}},
		{"Disable (not present)", []bool{true, false, false}, config.Config{Guilds: map[string]config.GuildConfig{
			"a": {},
		}}},
		{"Enable channel", []bool{true, true, false}, config.Config{Guilds: map[string]config.GuildConfig{
			"a": {
				// Should match; channel name matches
				EnabledChannels: []string{"a"},
			},
			"b": {
				// Should match; channel ID matches
				EnabledChannels: []string{"#b"},
			},
			"c": {
				// Should *not* match; neither channels match
				EnabledChannels: []string{"a", "#b"},
			},
		}}},
		{"Enable user", []bool{true, false, true}, config.Config{Guilds: map[string]config.GuildConfig{
			"a": {
				// Should match; first username matches
				EnabledUsers: []string{"Ethan Marshall", "abcdefg"},
			},
			"b": {
				// Should *not* match; neither names match
				EnabledUsers: []string{"Somebody else", "And somebody else"},
			},
			"c": {
				// Should match; second ID matches
				EnabledUsers: []string{"Apple user", "4206"},
			},
		}}},
		{"Precedence", []bool{false, false, false}, config.Config{Guilds: map[string]config.GuildConfig{
			// Should *not* match; disable overrides all other metrics
			"a": {
				Disable: true,
			},
			// Should *not* match; channel matches, but usernames do not
			"b": {
				Disable:         false,
				EnabledChannels: []string{"#b", "b"},
				EnabledUsers:    []string{"Not Cole Phelps", "Badge 1249"},
			},
			// Should *not* match; channel does not match, so username match is disregarded
			"c": {
				Disable:         false,
				EnabledChannels: []string{"d"},
				EnabledUsers:    []string{"Jay Irwin", "4206"},
			},
		}}},
	}
	TestMessages = []config.MessageMatcher{
		{
			Author:  discordgo.User{ID: "1234", Username: "Ethan Marshall"},
			Guild:   discordgo.Guild{ID: "a", Name: "a"},
			Channel: discordgo.Channel{ID: "#a", Name: "a"},
		},
		{
			Author:  discordgo.User{ID: "1247", Username: "Cole Phelps"},
			Guild:   discordgo.Guild{ID: "b", Name: "b"},
			Channel: discordgo.Channel{ID: "#b", Name: "Not the channel ID"},
		},
		{
			// I don't like this guy...
			Author:  discordgo.User{ID: "4206", Username: "Jay Irwin"},
			Guild:   discordgo.Guild{ID: "c", Name: "c"},
			Channel: discordgo.Channel{ID: "#c", Name: "c"},
		},
	}
)

func TestMatches(t *testing.T) {
	for _, test := range TestData {
		t.Run(test.Name, func(t *testing.T) {
			for i, msg := range TestMessages {
				if i >= len(TestMessages) {
					t.Log("WARNING: Not enough expects values for test:", test.Name)
					break
				}

				res := test.Config.MessageMatches(msg)
				if res != test.Expects[i] {
					t.Error(test.Name, "expected to get", test.Expects[i], "got", res, "for message", i)
				}
			}
		})
	}
}
