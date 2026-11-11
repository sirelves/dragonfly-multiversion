<h1 align="center">TM-Dragonfly</h1>

<p align="center">
  Minecraft: Bedrock server software in Go — one listener for every client from 1.16.0 to 1.26.60.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.26+-00ADD8?logo=go&logoColor=white" alt="Go 1.26+">
  <img src="https://img.shields.io/badge/bedrock-1.16.0%20→%201.26.60-9BE564" alt="Bedrock 1.16.0 to 1.26.60">
  <img src="https://img.shields.io/badge/license-MIT-brightgreen" alt="MIT License">
  <a href="https://t.me/TMDragonfly"><img src="https://img.shields.io/badge/telegram-@TMDragonfly-229ED9?logo=telegram&logoColor=white" alt="Telegram"></a>
</p>

## What is this?

A fork of [Dragonfly](https://github.com/df-mc/dragonfly) with full **multiversion** support:
every client from Bedrock **1.16.0 to 1.26.60** joins the same server, same port, no proxy. The
newest version runs natively; older ones are translated per client. Still the full Dragonfly API
underneath — build your game on top of it as a Go library.

## Features

- **Multiversion** — 1.16.0 … 1.26.60 on the same server and game port.
- **RakNet and NetherNet** — UDP gameplay with HTTP/HTTPS signaling for newer clients.
- **Config** — world mode (`current` or `flat`), render distance, difficulty, spawn protection.
- **Operators** — kept in `ops.yml`, with in-game commands: `weather`, `time set`, `gamemode`, `tp`, `give` (tab-complete items), `about`.
- **Protected packs** — drop a pack in `resources/` and it is served encrypted; the client's cached copy is unreadable.

## Getting started

Needs **Go 1.26+**.

```shell
git clone https://github.com/TrueTemm/TM-Dragonfly
cd TM-Dragonfly
go run .
```

A `config.toml` and `ops.yml` are written on first start. Type `stop` or press **ctrl+c** to shut
down. Release build: `go build -o tm-dragonfly .`

## Clients before 1.20.0

Protocols 407 (1.16.0) to 582 (1.19.80) live in `multiversion/legacy`. Each one
runs through the 1.20.0 path with its own block palette, item registry, biomes
and sounds, and a chain of per-release deltas then rewrites the wire, one
release at a time. The RakNet listener also speaks what those clients expect
before the packet layer: RakNet protocol 10, logins without
`RequestNetworkSettings` (before 1.19.30) and AES-CFB8 encryption (before
1.16.220).

- **Xbox Live.** These clients cannot authenticate anymore. With `AuthEnabled`
  on, clients below `Server.LegacyOfflineBelow` (589 by default) join without
  Xbox Live while everyone else still must log in. Their names are not
  verified, any XUID they claim is dropped, and they cannot be operators. Set
  `LegacyOfflineBelow = -1` to require Xbox Live from everyone.
- **World height.** Before 1.18.0 worlds span y 0–255: blocks, block entities
  and updates outside that range are not sent, and the column biome comes from
  y 64. Keep spawn and builds within it for those players.
- **What is translated.** Packets those clients have no decoder for are
  dropped; item tags in recipes become concrete items; smithing recipes are not
  offered below 1.19.80; entity metadata keys and flag bits are renumbered.
- **Not yet.** Command parameter types still use the 1.20.0 numbering, so
  argument hints may be off on the oldest clients.

### Verifying a protocol

`e2e/` drives real wire formats: every client and verifier is compiled against
the gophertunnel tag that shipped its protocol (`e2e/versions.txt`).

```shell
e2e/run.sh 407 545 582        # join, walk, chat; fails on any decode or server error
E2E_AUTH=1 e2e/run.sh 582 589 # the offline split: 582 gets in, 589 must log in
e2e/verify.sh 407             # every packet both ways against that tag's codec
./check.sh                    # build, vet, multiversion and server tests
```

Data provenance: block palettes, item registries, biomes and entities come from
the per-protocol dumps of IdotClub/BedrockData (LGPL-2.1, a fork of
pmmp/BedrockData) and block upgrade schemas from pmmp/BedrockBlockUpgradeSchema
(CC0); sound, particle, entity key and flag tables are evaluated from
CloudburstMC/Protocol's codecs by `cmd/mvgen`.

## Network ports

Clients use the same game address and port for every supported version. The default is `19132`.

| Protocol | Port | Purpose |
| --- | --- | --- |
| UDP | `19132` | RakNet and NetherNet game traffic |
| TCP | `19132` | NetherNet HTTP/HTTPS signaling |

Open both ports on the host firewall and any provider firewall:

```shell
sudo ufw allow 19132/udp
sudo ufw allow 19132/tcp
```

NetherNet is enabled by default with a shared UDP socket. If `SharedUDP` is disabled in
`config.toml`, its ICE UDP port defaults to `19133` and must also be opened. Set the NetherNet
certificate and private key paths in `config.toml` to enable HTTPS signaling. Port hints are
printed only when a NetherNet listener fails to open or stops with an error.

## Console commands

`help`, `list`, `status`, `version`, `about`, `weather`, `time set`, `op <nick>`, `deop <nick>`,
`say <message>`, `stop`. Operators also get the in-game commands listed above.

## License

MIT — see [LICENSE](LICENSE). Credits: [Dragonfly](https://github.com/df-mc/dragonfly).

Telegram: [@TMDragonfly](https://t.me/TMDragonfly)
