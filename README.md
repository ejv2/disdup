# ``disdup`` - Discord message bouncer and duplicator

Package ``disdup`` implements a simple concurrent Discord message bouncer and transcriber which is able to transcribe incoming discord messages to various formats for bouncing and forwarding. It is designed for either self-hosting or for embedding in larger applications.

``disdup`` is formatted primarily as an independent package, but also contains a pre-built CLI at cmd/disdup.

## Formats

``disdup`` supports bouncing to the following locations:

* Email (through smtp(s))
* IRC
* Plain logs

## Install (package)

To use the disdup package, simply import

```go
import "github.com/ethanv2/disdup"
```

Then run

```shell
go mod tidy
```

## Install (CLI)

To use the ``disdup`` CLI, navigate to cmd/disdup. Before running anything, copy the sample configs to their actual locations (remove ".sample" from the end of their names). Then, modify them to your liking. Supported outputs and their configuration are listed below.

To build the CLI, run

```shell
go build
```

Alternatively, just

```shell
go run
```

## Configuration

Configuration is parsed declaratively through the source code. The best way to see the details of the config is to read the source, or to run:

```shell
go doc -all conf
```

### Primary config

The primary config is located at ``disdup.conf``. It contains the bot's token and other bookkeeping details. It also contains a list of allowed guilds ("servers") and their properties. Guilds not listed in this configuration file will not be duplicated at all. Guilds may be specified by name or by ID (which can be copied from the Discord UI).

Each guild can have zero or more ``enabled_channels``. If no enabled channels are listed, all channels are enabled. Else, only the channels listed by name or ID will be duplicated from. This does not override the guild being disabled.

Each guild can have zero or more ``enabled_users``. If no enabled users are listed, all users are enabled. Else, only the users listed by full username "name#tag" or ID will be duplicated from. This does not override the ``enabled_channels``, nor the guild being disabled.

Each guild also has an associated set of outputs, which are the names of the outputs specified in the outputs config file. See the next section for details.

### Outputs

The outputs config is a map between output names and their properties. In the sample config, one output is declared named "print". These names may be referenced by the ``output`` array in a guild's configuration. Every output has an associated ``type`` and a list of ``args``.

Available output ``type``s are as follows:

* "stdout": logs all messages to standard output in a known fashion. Can be collated by channel or by user and channel. Has a configurable prefix to denote output from this specific output.
* "command": runs a command with configurable arguments whenever a message is received. Arguments can contain formatting directives which pass information about a message to the command.
* "mail": send an email containing the message contents, attachments, etc. to a specific mailbox.

Outputs also take an object called ``args``. These are specific to each output. Unknown options are ignored, but some outputs require that some args are provided. For instance, "command" requires that a "cmd" key for the command be provided.
