package config

// Config is the primary disdup configuration, optionally encoded in JSON
// format and loaded by the client code. It is passed to the main duplicator,
// which then uses it for reference.
type Config struct {
	// Token is the bot's authorization token for the Discord API
	Token string `json:"token"`
	// Name is the nickname the bot will assume upon being added to a guild
	Name string `json:"name"`
	// Guilds is a map of guild names or IDs to their associated
	// configuration.
	Guilds map[string]GuildConfig `json:"guilds"`
}

// GuildConfig represents the configuration for a single guild. It may be
// configured via either a name or guild ID, the ID taking precedence.
type GuildConfig struct {
	// Disable this guild? Disabled guilds will be entirely ignored for
	// duplication
	Disable bool `json:"disable"`
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
