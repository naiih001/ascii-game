package sim

import (
	"sort"

	"ascii-game/internal/entity"
	gamemath "ascii-game/internal/math"
	"ascii-game/internal/world"
)

type PlayerID uint64

type CommandType byte

const (
	CommandMove CommandType = iota + 1
)

type Command struct {
	Type     CommandType
	PlayerID PlayerID
	Delta    gamemath.Vec2
}

type PlayerState struct {
	ID     PlayerID
	Pos    gamemath.Vec2
	Sprite rune
}

type Snapshot struct {
	Tick    int64
	Players []PlayerState
}

type Simulation struct {
	world      *world.Map
	players    map[PlayerID]entity.Player
	commandBuf []Command
	nextID     PlayerID
	tick       int64
}

func New(worldMap *world.Map) *Simulation {
	return &Simulation{
		world:   worldMap,
		players: make(map[PlayerID]entity.Player),
		nextID:  1,
	}
}

func (s *Simulation) World() *world.Map {
	return s.world
}

func (s *Simulation) AddPlayer(pos gamemath.Vec2) PlayerID {
	id := s.nextID
	s.nextID++
	s.players[id] = entity.NewPlayer(pos)
	return id
}

func (s *Simulation) RemovePlayer(playerID PlayerID) {
	delete(s.players, playerID)
}

func (s *Simulation) QueueMove(playerID PlayerID, delta gamemath.Vec2) {
	s.commandBuf = append(s.commandBuf, Command{
		Type:     CommandMove,
		PlayerID: playerID,
		Delta:    delta,
	})
}

func (s *Simulation) Tick() {
	for _, cmd := range s.commandBuf {
		switch cmd.Type {
		case CommandMove:
			s.movePlayer(cmd.PlayerID, cmd.Delta)
		}
	}

	s.commandBuf = s.commandBuf[:0]
	s.tick++
}

func (s *Simulation) Snapshot() Snapshot {
	players := make([]PlayerState, 0, len(s.players))
	for id, player := range s.players {
		players = append(players, PlayerState{
			ID:     id,
			Pos:    player.Pos,
			Sprite: player.Sprite,
		})
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].ID < players[j].ID
	})

	return Snapshot{
		Tick:    s.tick,
		Players: players,
	}
}

func (s *Simulation) movePlayer(playerID PlayerID, delta gamemath.Vec2) {
	player, ok := s.players[playerID]
	if !ok {
		return
	}

	next := player.Pos.Add(delta)
	if !s.world.IsWalkable(next) {
		return
	}

	player.Pos = next
	s.players[playerID] = player
}
