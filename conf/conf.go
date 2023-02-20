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
	Guilds map[string]*GuildConfig `json:"guilds"`
	// Outputs is a map of output names to the output interface which will
	// be used. On duplicator startup, all outputs have their "Open" method
	// called concurrently. On shutdown, all outputs have their "Close"
	// method called exactly once. Outputs are not automatically decoded
	// from json by default and must be initialized explicitly be the
	// caller.
	Outputs []OutputConfig `json:"-"`
}

// Guild registers a new guild for duplication and enables it by default. The
// parameter `nameid` may be the name or ID of the guild, with the ID taking
// precedence. If a guild with the same name or ID has already been registered,
// Guild panics.
func (c *Config) Guild(nameid string) *GuildConfig {
	if _, ok := c.Guilds[nameid]; ok {
		panic("disdup config: duplicate registration of guild ID " + nameid)
	}

	cfg := &GuildConfig{}
	c.Guilds[nameid] = cfg
	return cfg
}

// Use adds an output to the output array. If an output with the same name has
// already been registered, Use panics. A pointer to the target config is
// returned for use in function chaining.
func (c *Config) Use(name string, output output.Output) *Config {
	for _, elem := range c.Outputs {
		if elem.Name == name {
			panic("disdup config: duplicate use of output name " + name)
		}
	}

	c.Outputs = append(c.Outputs, OutputConfig{Name: name, Output: output})
	return c
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

// Disable marks the guild as disabled. No messages shall be duplicated from
// it, regardless of other configuration.
func (g *GuildConfig) Disabled() *GuildConfig {
	g.Disable = true
	return g
}

// Use configures an output for use on this guild. If this is not called at
// least once, all outputs are enabled. No error checking is done on this value
// at configuration time. If the output does not exist, it is ignored and no
// output is produced.
func (g *GuildConfig) Use(name string) *GuildConfig {
	g.Output = append(g.Output, name)
	return g
}

// UseAll undoes all calls to Use and enabled all outputs on this guild. This
// is not necessary to call unless Use has been called at least once and must
// be undone.
func (g *GuildConfig) UseAll() *GuildConfig {
	g.Output = nil
	return g
}

// Channel enables a channel with the name or id `nameid`.
func (g *GuildConfig) Channel(nameid string) *GuildConfig {
	g.EnabledChannels = append(g.EnabledChannels, nameid)
	return g
}

// User enables a user with the name or id `nameid`.
func (g *GuildConfig) User(nameid string) *GuildConfig {
	g.EnabledUsers = append(g.EnabledUsers, nameid)
	return g
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
