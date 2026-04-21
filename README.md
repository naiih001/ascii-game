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
make game ADDR=your-public-host:your-public-port
```

## Make targets

- `make game` runs the multiplayer ASCII client
- `make server` runs the multiplayer server
- `make fmt` formats the Go code
- `make test` runs the Go test suite
- `make tidy` syncs module dependencies

## Controls

- Move: arrow keys or `hjkl`
- Quit: `Esc` or `Ctrl+C`

## Current slice

- Authoritative TCP server at 30 TPS
- Terminal client that renders server snapshots
- Async keyboard input forwarded as move commands
- Large tile map with visible collidable terrain
- Camera follows the local player from replicated state

## Railway deploy

Use Railway for this project's current raw TCP server shape.

1. Push this repo to GitHub.
2. In Railway, create a new project from the GitHub repo.
3. In the service settings, set the start command to:

```bash
go run ./cmd/server
```

4. Deploy the service.
5. In Railway service settings, create a TCP proxy for the service.
6. Railway will give you a public TCP endpoint like `something.proxy.rlwy.net:15140`.
7. Every player runs the client locally and points it at that public endpoint:

```bash
make game ADDR=something.proxy.rlwy.net:15140
```

Notes:

- The server now automatically listens on `0.0.0.0:$PORT` when Railway provides `PORT`.
- Local development still uses `127.0.0.1:7777` by default.
- Only the server is deployed to Railway. The `tcell` client still runs in each player's terminal.
