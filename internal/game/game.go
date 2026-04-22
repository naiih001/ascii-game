package game

import (
	"fmt"
	"strings"

	"ascii-game/internal/entity"
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
	case input.ActionUseHealth:
		g.sendUseItem(entity.InventorySlotHealth)
	case input.ActionUseShield:
		g.sendUseItem(entity.InventorySlotShield)
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
	hudLines := renderHUD(localPlayer, snapshot.Tick, screenW)
	viewportH := screenH - len(hudLines)
	if screenW <= 0 || viewportH <= 0 {
		g.screen.Show()
		return
	}

	camera := gamemath.Vec2{
		X: localPlayer.Pos.X - (screenW / 2),
		Y: localPlayer.Pos.Y - (viewportH / 2),
	}

	if camera.X < 0 {
		camera.X = 0
	}
	if camera.Y < 0 {
		camera.Y = 0
	}
	maxCameraX := g.world.Width - screenW
	maxCameraY := g.world.Height - viewportH
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
	corpseStyle := tcell.StyleDefault.Foreground(tcell.ColorMaroon)
	spikeTrapStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	empTrapStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	acidTrapStyle := tcell.StyleDefault.Foreground(tcell.ColorYellowGreen)
	hudStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

	for sy := 0; sy < viewportH; sy++ {
		for sx := 0; sx < screenW; sx++ {
			worldPos := gamemath.Vec2{X: camera.X + sx, Y: camera.Y + sy}
			tile := g.world.TileAt(worldPos)
			drawRune, style := drawTile(tile, floorStyle, wallStyle)
			g.screen.DrawRune(sx, sy, drawRune, style)
		}
	}

	for _, trap := range snapshot.Traps {
		trapScreenX := trap.Pos.X - camera.X
		trapScreenY := trap.Pos.Y - camera.Y
		if trapScreenX >= 0 && trapScreenX < screenW && trapScreenY >= 0 && trapScreenY < viewportH {
			g.screen.DrawRune(trapScreenX, trapScreenY, drawTrap(trap.Type), trapStyle(trap.Type, spikeTrapStyle, empTrapStyle, acidTrapStyle))
		}
	}

	for _, player := range snapshot.Players {
		playerScreenX := player.Pos.X - camera.X
		playerScreenY := player.Pos.Y - camera.Y
		if playerScreenX >= 0 && playerScreenX < screenW && playerScreenY >= 0 && playerScreenY < viewportH {
			style := playerStyle
			if player.Health <= 0 {
				style = corpseStyle
			}
			g.screen.DrawRune(playerScreenX, playerScreenY, player.Sprite, style)
		}
	}

	for row, line := range hudLines {
		y := viewportH + row
		for x := 0; x < screenW; x++ {
			ch := ' '
			if x < len(line) {
				ch = rune(line[x])
			}
			g.screen.DrawRune(x, y, ch, hudStyle)
		}
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

func (g *Game) sendUseItem(slot int) {
	if err := g.client.SendUseItem(slot); err != nil {
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

func drawTrap(trapType world.TrapType) rune {
	switch trapType {
	case world.TrapSpike:
		return '^'
	case world.TrapEMP:
		return '~'
	case world.TrapAcid:
		return '!'
	default:
		return '?'
	}
}

func trapStyle(trapType world.TrapType, spikeStyle, empStyle, acidStyle tcell.Style) tcell.Style {
	switch trapType {
	case world.TrapSpike:
		return spikeStyle
	case world.TrapEMP:
		return empStyle
	case world.TrapAcid:
		return acidStyle
	default:
		return tcell.StyleDefault.Foreground(tcell.ColorWhite)
	}
}

func renderHUD(localPlayer sim.PlayerState, tick int64, width int) []string {
	if width < 24 {
		return []string{
			padHUDLine(fmt.Sprintf("HP %d/%d SP %d/%d", localPlayer.Health, localPlayer.MaxHP, localPlayer.Shield, localPlayer.MaxSP), width),
		}
	}

	info := fmt.Sprintf("POS %02d,%02d  TICK %d  MOVE arrows/hjkl  ITEMS 1/2  QUIT esc", localPlayer.Pos.X, localPlayer.Pos.Y, tick)
	lines := []string{
		padHUDLine(boxLine('=', width), width),
		padHUDLine(fmt.Sprintf("HP %s %3d/%3d", meter(localPlayer.Health, localPlayer.MaxHP, hudMeterWidth(width)), localPlayer.Health, localPlayer.MaxHP), width),
		padHUDLine(fmt.Sprintf("SP %s %3d/%3d", meter(localPlayer.Shield, localPlayer.MaxSP, hudMeterWidth(width)), localPlayer.Shield, localPlayer.MaxSP), width),
		padHUDLine(fmt.Sprintf("[1] Medkit x%d    [2] Shield x%d", localPlayer.HealthItemCount, localPlayer.ShieldItemCount), width),
		padHUDLine(info, width),
		padHUDLine(boxLine('=', width), width),
	}

	return lines
}

func hudMeterWidth(width int) int {
	meterWidth := width - 15
	if meterWidth < 10 {
		return 10
	}
	if meterWidth > 28 {
		return 28
	}
	return meterWidth
}

func meter(current, max, width int) string {
	if max <= 0 {
		return strings.Repeat(".", width)
	}

	filled := current * width / max
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	return strings.Repeat("#", filled) + strings.Repeat(".", width-filled)
}

func boxLine(ch rune, width int) string {
	if width <= 0 {
		return ""
	}

	return strings.Repeat(string(ch), width)
}

func padHUDLine(line string, width int) string {
	if len(line) >= width {
		return line[:width]
	}
	return line + strings.Repeat(" ", width-len(line))
}

func playerByID(snapshot sim.Snapshot, playerID sim.PlayerID) (sim.PlayerState, bool) {
	for _, player := range snapshot.Players {
		if player.ID == playerID {
			return player, true
		}
	}

	return sim.PlayerState{}, false
}
