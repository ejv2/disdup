# ``disdup`` - Discord message bouncer and duplicator

``disdup`` is a simple concurrent Discord message bouncer and transcriber which is able to transcribe plain discord messages to various formats for bouncing and forwarding. It is designed for self-hosting and for inviting into server channels or private message groups.

## Formats

``disdup`` supports bouncing to the following locations:

* Email (to a server or to mbox or maildir)
* IRC
* Plain logs

This is achieved using a plugin-based system, which has a component per-output which can speak the necessary language and send via the necessary channels (eg: format as an IRC message and speak to an IRC server).
