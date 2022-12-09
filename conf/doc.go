// Package config contains structures which serve as Disdup's configuration .
// These can be either filled in by the user of the package, or unmarshalled
// from JSON using the JSON tags provided. Not all tags (such as Outputs)
// support being read from JSON. As a result, it is the caller's responsibility
// to fill out these fields.
//
// The zero-values of each type are designed to be valid configs, although they
// may not function particularly well. However, the primary configuration
// "struct Config" requires that the Token field at least be valid. If nothing
// else is provided, Disdup will receive and silently discard incoming events
// from Discord.
//
// Package config also implements the simple algorithm for checking if a
// message is supposed to be duplicated using details from the message.
package config
