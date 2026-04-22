package world

import (
	"sort"

	gamemath "ascii-game/internal/math"
)

const (
	defaultMapWidth  = 240
	defaultMapHeight = 120
)

type Map struct {
	Tiles  [][]rune
	Width  int
	Height int
	Traps  map[gamemath.Vec2]Trap
}

type TrapType byte

const (
	TrapSpike TrapType = iota + 1
	TrapEMP
	TrapAcid
)

type Trap struct {
	Pos         gamemath.Vec2
	Type        TrapType
	Damage      int
	ShieldDrain int
	Triggered   bool
	OneShot     bool
}

func NewDefaultMap() *Map {
	tiles := make([][]rune, defaultMapHeight)
	for y := 0; y < defaultMapHeight; y++ {
		row := make([]rune, defaultMapWidth)
		for x := 0; x < defaultMapWidth; x++ {
			if x == 0 || y == 0 || x == defaultMapWidth-1 || y == defaultMapHeight-1 {
				row[x] = '#'
				continue
			}

			row[x] = '.'
		}
		tiles[y] = row
	}

	carveSpawnPlaza(tiles, 2, 2, 18, 10)
	addVerticalBand(tiles, 28, 3, 36, 5)
	addVerticalBand(tiles, 63, 10, 42, 6)
	addVerticalBand(tiles, 104, 5, 58, 4)
	addVerticalBand(tiles, 150, 18, 70, 7)
	addVerticalBand(tiles, 196, 8, 54, 5)

	addHorizontalBand(tiles, 20, 16, 34, 6)
	addHorizontalBand(tiles, 45, 52, 28, 5)
	addHorizontalBand(tiles, 70, 84, 44, 8)
	addHorizontalBand(tiles, 95, 132, 32, 6)

	carveRoad(tiles, 12, 12, 220, 12)
	carveRoad(tiles, 20, 40, 218, 40)
	carveRoad(tiles, 16, 76, 226, 76)
	carveRoad(tiles, 18, 108, 210, 108)

	carveRoad(tiles, 24, 8, 24, 112)
	carveRoad(tiles, 82, 6, 82, 114)
	carveRoad(tiles, 138, 4, 138, 110)
	carveRoad(tiles, 206, 10, 206, 116)

	m := &Map{
		Tiles:  tiles,
		Width:  defaultMapWidth,
		Height: defaultMapHeight,
		Traps:  make(map[gamemath.Vec2]Trap),
	}

	populateDefaultTraps(m)

	return m
}

func (m *Map) InBounds(pos gamemath.Vec2) bool {
	return pos.X >= 0 && pos.X < m.Width && pos.Y >= 0 && pos.Y < m.Height
}

func (m *Map) IsWalkable(pos gamemath.Vec2) bool {
	if !m.InBounds(pos) {
		return false
	}

	return m.Tiles[pos.Y][pos.X] != '#'
}

func (m *Map) TileAt(pos gamemath.Vec2) rune {
	if !m.InBounds(pos) {
		return ' '
	}

	return m.Tiles[pos.Y][pos.X]
}

func (m *Map) AddTrap(pos gamemath.Vec2, trap Trap) {
	trap.Pos = pos
	m.Traps[pos] = trap
}

func (m *Map) TrapAt(pos gamemath.Vec2) (Trap, bool) {
	trap, ok := m.Traps[pos]
	return trap, ok
}

func (m *Map) SetTrap(trap Trap) {
	m.Traps[trap.Pos] = trap
}

func (m *Map) RemoveTrap(pos gamemath.Vec2) {
	delete(m.Traps, pos)
}

func (m *Map) ActiveTraps() []Trap {
	traps := make([]Trap, 0, len(m.Traps))
	for _, trap := range m.Traps {
		traps = append(traps, trap)
	}

	sort.Slice(traps, func(i, j int) bool {
		if traps[i].Pos.Y == traps[j].Pos.Y {
			return traps[i].Pos.X < traps[j].Pos.X
		}
		return traps[i].Pos.Y < traps[j].Pos.Y
	})

	return traps
}

func NewSpikeTrap() Trap {
	return Trap{
		Type:    TrapSpike,
		Damage:  20,
		OneShot: true,
	}
}

func NewEMPTrap() Trap {
	return Trap{
		Type:        TrapEMP,
		ShieldDrain: 30,
	}
}

func NewAcidTrap() Trap {
	return Trap{
		Type:        TrapAcid,
		Damage:      15,
		ShieldDrain: 15,
		OneShot:     true,
	}
}

func carveSpawnPlaza(tiles [][]rune, startX, startY, width, height int) {
	for y := startY; y < startY+height && y < len(tiles)-1; y++ {
		for x := startX; x < startX+width && x < len(tiles[y])-1; x++ {
			tiles[y][x] = '.'
		}
	}
}

func addVerticalBand(tiles [][]rune, x, startY, height, gap int) {
	for y := startY; y < startY+height && y < len(tiles)-1; y++ {
		if gap > 0 && (y-startY)%gap == 0 {
			continue
		}
		setWall(tiles, x, y)
		setWall(tiles, x+1, y)
	}
}

func addHorizontalBand(tiles [][]rune, y, startX, width, gap int) {
	for x := startX; x < startX+width && x < len(tiles[y])-1; x++ {
		if gap > 0 && (x-startX)%gap == 0 {
			continue
		}
		setWall(tiles, x, y)
		setWall(tiles, x, y+1)
	}
}

func carveRoad(tiles [][]rune, startX, startY, endX, endY int) {
	if startX == endX {
		if startY > endY {
			startY, endY = endY, startY
		}
		for y := startY; y <= endY; y++ {
			clearRoadTile(tiles, startX, y)
			clearRoadTile(tiles, startX+1, y)
		}
		return
	}

	if startY == endY {
		if startX > endX {
			startX, endX = endX, startX
		}
		for x := startX; x <= endX; x++ {
			clearRoadTile(tiles, x, startY)
			clearRoadTile(tiles, x, startY+1)
		}
	}
}

func clearRoadTile(tiles [][]rune, x, y int) {
	if y <= 0 || y >= len(tiles)-1 || x <= 0 || x >= len(tiles[y])-1 {
		return
	}

	tiles[y][x] = '.'
}

func setWall(tiles [][]rune, x, y int) {
	if y <= 0 || y >= len(tiles)-1 || x <= 0 || x >= len(tiles[y])-1 {
		return
	}

	tiles[y][x] = '#'
}

func populateDefaultTraps(m *Map) {
	for pos, trap := range map[gamemath.Vec2]Trap{
		{X: 18, Y: 12}:  NewSpikeTrap(),
		{X: 42, Y: 40}:  NewEMPTrap(),
		{X: 76, Y: 76}:  NewAcidTrap(),
		{X: 138, Y: 28}: NewSpikeTrap(),
		{X: 206, Y: 96}: NewEMPTrap(),
	} {
		if m.IsWalkable(pos) {
			m.AddTrap(pos, trap)
		}
	}
}
