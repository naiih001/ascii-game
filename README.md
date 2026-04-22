# ASCII Game Prototype

Small Go ASCII game starter using `tcell`.

## Run

```bash
make tidy
make server
make game
```

For another client on the same machine:

```bash
make game
```

For clients connecting to a deployed server:

```bash
make game ADDR=your-public-host
```

## Make targets

- `make game` runs the multiplayer ASCII client
- `make server` runs the multiplayer server
- `make fmt` formats the Go code
- `make test` runs the Go test suite
- `make tidy` syncs module dependencies

## Controls

- Move: arrow keys or `hjkl`
- Use medkit: `1`
- Use shield battery: `2`
- Quit: `Esc` or `Ctrl+C`

## Current slice

- Authoritative TCP server at 30 TPS
- Terminal client that renders server snapshots
- Async keyboard input forwarded as move and item-use commands
- Large tile map with visible collidable terrain
- Camera follows the local player from replicated state
- Server-authoritative health, shield, inventory, and trap interactions
- Multi-line HUD with replicated HP/SP bars and inventory counts

## Railway deploy

Use Railway's normal HTTP service shape for this project. The game server now upgrades `GET /ws` to a WebSocket connection and keeps the existing binary protocol inside those frames.

This repo includes a checked-in [railway.json](/home/ryth/projects/ascii-game/railway.json) so Railway builds the dedicated server binary, starts the correct entrypoint, and exposes a simple `/healthz` endpoint for health checks.

1. Push this repo to GitHub.
2. In Railway, create a new project from the GitHub repo.
3. Deploy the service.
4. Deploy the service and use the generated Railway public domain such as `your-game.up.railway.app`.
5. Every player runs the client locally and points it at that public host:

```bash
make game ADDR=your-game.up.railway.app
```

Notes:

- Railway config builds `./cmd/server` into `./bin/server` and starts that binary.
- The server listens on `0.0.0.0:$PORT` when Railway provides `PORT` and serves WebSockets on `/ws`.
- Local development still uses `127.0.0.1:7777` by default.
- The local client automatically uses `ws://` for localhost addresses and `wss://` for non-local hosts.
- Only the server is deployed to Railway. The `tcell` client still runs in each player's terminal.
