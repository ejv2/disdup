package config

import "github.com/ejv2/disdup/output"

// Config is the primary disdup configuration, optionally encoded in JSON
// format and loaded by the client code. It is passed to the main duplicator,
// which then uses it for reference.
type Config struct {
	// Token is the bot's authorization token for the Discord API
	Token string `json:"token"`
	// Name is the nickname the bot will assume upon being added to a guild
	Name string `json:"name"`
	// Guilds is a map of guild names or IDs to their associated
	// configuration. This is not an optional key: servers not configured
	// are ignored
	Guilds map[string]GuildConfig `json:"guilds"`
	// Outputs is a map of output names to the output interface which will
	// be used. On duplicator startup, all outputs have their "Open" method
	// called concurrently. On shutdown, all outputs have their "Close"
	// method called exactly once. Outputs are not automatically decoded
	// from json by default and must be initialized explicitly be the
	// caller.
	Outputs []OutputConfig `json:"-"`
}

// GuildConfig represents the configuration for a single guild. It may be
// configured via either a name or guild ID, the ID taking precedence. The zero
// value of this type is a valid configuration which duplicates all messages
// from a server.
type GuildConfig struct {
	// Disable this guild? Disabled guilds will be entirely ignored for
	// duplication. Guilds are enabled by default
	Disable bool `json:"disable"`
	// Output to the output with these names. If empty, all outputs are selected.
	Output []string `json:"output"`
	// EnabledChannels are which channels the bot will duplicate from. This
	// does not override "enable"; disabled guilds will still not be
	// duplicated. If empty, all channels are enabled. Channel names should
	// not include the leading '#'
	EnabledChannels []string `json:"enabled_channels"`
	// EnabledUsers are users which the bot will duplicate messages from.
	// This does not override "enabled" or "enabled_channels"; disabled
	// guilds or channels will still not be duplicated. If empty, all
	// users in all enabled channels will be duplicated. Users *must*
	// include the full tag (i.e: user#tag)
	EnabledUsers []string `json:"enabled_users"`
}

// OutputConfig represents one entry for an output handler which associates a
// name with an output interface.
//
// This is not represented using a map of string to output.Output, as we need
// to cycle through every output quite commonly, which is not as efficient
// using a map as with a slice.
type OutputConfig struct {
	// Name is the name by which this output can be referenced by other
	// configuration options.
	Name string
	// Output is the target for the output.
	Output output.Output
}
