# lod2.zip

Contains the Go application powering [lod2.zip](https://lod2.zip/).

## Principles

* **Graceful degradation.** The server is self-reliant in that (at the moment) it is responsible for handling GitHub push webhooks to trigger a rebuild. If the server hard crashes, it will need manual SSH intervention to restart.
* **Self-sufficient.** Minimal external depenendencies reduce the chance of unexpected, uncontrolled issues and improve reliability.

## Files

Required files in the config directory (defaults to `~/.config/lod2/`):

* `keys/auth/privkey.jwk.json`: the private JWK used for signing authentication JWTs.
