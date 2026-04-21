package game

import (
	"fmt"

	"ascii-game/internal/input"
	gamemath "ascii-game/internal/math"
	"ascii-game/internal/netclient"
	"ascii-game/internal/render"
	"ascii-game/internal/sim"
	"ascii-game/internal/world"
	"github.com/gdamore/tcell/v2"
)

type Game struct {
	screen        *render.Screen
	world         *world.Map
	client        *netclient.Client
	localPlayerID sim.PlayerID
	input         chan input.Action
	running       bool
	runErr        error
}

func New(serverAddr string) (*Game, error) {
	screen, err := render.New()
	if err != nil {
		return nil, err
	}

	client, err := netclient.Connect(serverAddr)
	if err != nil {
		screen.Close()
		return nil, err
	}

	g := &Game{
		screen:        screen,
		world:         world.NewDefaultMap(),
		client:        client,
		localPlayerID: client.LocalPlayerID(),
		input:         make(chan input.Action, 32),
		running:       true,
	}

	go input.Poll(screen.Raw(), g.input)

	return g, nil
}

func (g *Game) Close() {
	if g.client != nil {
		_ = g.client.Close()
	}
	g.screen.Close()
}

func (g *Game) handleAction(action input.Action) {
	switch action {
	case input.ActionMoveUp:
		g.sendMove(gamemath.Vec2{X: 0, Y: -1})
	case input.ActionMoveDown:
		g.sendMove(gamemath.Vec2{X: 0, Y: 1})
	case input.ActionMoveLeft:
		g.sendMove(gamemath.Vec2{X: -1, Y: 0})
	case input.ActionMoveRight:
		g.sendMove(gamemath.Vec2{X: 1, Y: 0})
	case input.ActionQuit:
		g.running = false
	case input.ActionResize:
	}
}

func (g *Game) Update() {
	if err := g.client.Err(); err != nil {
		g.runErr = err
		g.running = false
	}
}

func (g *Game) Render() {
	g.screen.Clear()
	snapshot := g.client.Snapshot()
	localPlayer, ok := playerByID(snapshot, g.localPlayerID)
	if !ok && len(snapshot.Players) == 0 {
		g.drawWaitingScreen("Waiting for first snapshot from server...")
		return
	}
	if !ok {
		g.screen.Show()
		return
	}

	screenW, screenH := g.screen.Size()
	if screenW <= 0 || screenH <= 1 {
		g.screen.Show()
		return
	}

	camera := gamemath.Vec2{
		X: localPlayer.Pos.X - (screenW / 2),
		Y: localPlayer.Pos.Y - ((screenH - 1) / 2),
	}

	if camera.X < 0 {
		camera.X = 0
	}
	if camera.Y < 0 {
		camera.Y = 0
	}
	maxCameraX := g.world.Width - screenW
	maxCameraY := g.world.Height - (screenH - 1)
	if maxCameraX < 0 {
		maxCameraX = 0
	}
	if maxCameraY < 0 {
		maxCameraY = 0
	}
	if camera.X > maxCameraX {
		camera.X = maxCameraX
	}
	if camera.Y > maxCameraY {
		camera.Y = maxCameraY
	}

	floorStyle := tcell.StyleDefault.Foreground(tcell.ColorReset)
	wallStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorOlive)
	playerStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	hudStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

	for sy := 0; sy < screenH-1; sy++ {
		for sx := 0; sx < screenW; sx++ {
			worldPos := gamemath.Vec2{X: camera.X + sx, Y: camera.Y + sy}
			tile := g.world.TileAt(worldPos)
			drawRune, style := drawTile(tile, floorStyle, wallStyle)
			g.screen.DrawRune(sx, sy, drawRune, style)
		}
	}

	for _, player := range snapshot.Players {
		playerScreenX := player.Pos.X - camera.X
		playerScreenY := player.Pos.Y - camera.Y
		if playerScreenX >= 0 && playerScreenX < screenW && playerScreenY >= 0 && playerScreenY < screenH-1 {
			g.screen.DrawRune(playerScreenX, playerScreenY, player.Sprite, playerStyle)
		}
	}

	status := fmt.Sprintf("POS %02d,%02d  TICK %d  MOVE arrows/hjkl  QUIT esc", localPlayer.Pos.X, localPlayer.Pos.Y, snapshot.Tick)
	for x := 0; x < screenW; x++ {
		ch := ' '
		if x < len(status) {
			ch = rune(status[x])
		}
		g.screen.DrawRune(x, screenH-1, ch, hudStyle)
	}

	g.screen.Show()
}

func (g *Game) Err() error {
	return g.runErr
}

func (g *Game) sendMove(delta gamemath.Vec2) {
	if err := g.client.SendMove(delta); err != nil {
		g.runErr = err
		g.running = false
	}
}

func (g *Game) drawWaitingScreen(message string) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for x, ch := range message {
		g.screen.DrawRune(x, 0, ch, style)
	}
	g.screen.Show()
}

func drawTile(tile rune, floorStyle, wallStyle tcell.Style) (rune, tcell.Style) {
	switch tile {
	case '#':
		return 'X', wallStyle
	default:
		return '.', floorStyle
	}
}

func playerByID(snapshot sim.Snapshot, playerID sim.PlayerID) (sim.PlayerState, bool) {
	for _, player := range snapshot.Players {
		if player.ID == playerID {
			return player, true
		}
	}

	return sim.PlayerState{}, false
}
