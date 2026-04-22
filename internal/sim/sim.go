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
	CommandUseItem
)

type Command struct {
	Type     CommandType
	PlayerID PlayerID
	Delta    gamemath.Vec2
	Slot     int
}

type PlayerState struct {
	ID              PlayerID
	Pos             gamemath.Vec2
	Sprite          rune
	Health          int
	Shield          int
	MaxHP           int
	MaxSP           int
	HealthItemCount int
	ShieldItemCount int
}

type TrapState struct {
	Pos  gamemath.Vec2
	Type world.TrapType
}

type Snapshot struct {
	Tick    int64
	Players []PlayerState
	Traps   []TrapState
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

func (s *Simulation) QueueUseItem(playerID PlayerID, slot int) {
	s.commandBuf = append(s.commandBuf, Command{
		Type:     CommandUseItem,
		PlayerID: playerID,
		Slot:     slot,
	})
}

func (s *Simulation) Tick() {
	for _, cmd := range s.commandBuf {
		switch cmd.Type {
		case CommandMove:
			s.movePlayer(cmd.PlayerID, cmd.Delta)
		case CommandUseItem:
			s.useItem(cmd.PlayerID, cmd.Slot)
		}
	}

	s.commandBuf = s.commandBuf[:0]
	s.tick++
}

func (s *Simulation) Snapshot() Snapshot {
	players := make([]PlayerState, 0, len(s.players))
	for id, player := range s.players {
		sprite := player.Sprite
		if !player.Alive() {
			sprite = '%'
		}

		players = append(players, PlayerState{
			ID:              id,
			Pos:             player.Pos,
			Sprite:          sprite,
			Health:          player.Stats.Health,
			Shield:          player.Stats.Shield,
			MaxHP:           player.Stats.MaxHP,
			MaxSP:           player.Stats.MaxSP,
			HealthItemCount: player.Inventory[entity.InventorySlotHealth].Quantity,
			ShieldItemCount: player.Inventory[entity.InventorySlotShield].Quantity,
		})
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].ID < players[j].ID
	})

	traps := make([]TrapState, 0, len(s.world.Traps))
	for _, trap := range s.world.ActiveTraps() {
		traps = append(traps, TrapState{
			Pos:  trap.Pos,
			Type: trap.Type,
		})
	}

	sort.Slice(traps, func(i, j int) bool {
		if traps[i].Pos.Y == traps[j].Pos.Y {
			return traps[i].Pos.X < traps[j].Pos.X
		}
		return traps[i].Pos.Y < traps[j].Pos.Y
	})

	return Snapshot{
		Tick:    s.tick,
		Players: players,
		Traps:   traps,
	}
}

func (s *Simulation) movePlayer(playerID PlayerID, delta gamemath.Vec2) {
	player, ok := s.players[playerID]
	if !ok || !player.Alive() {
		return
	}

	next := player.Pos.Add(delta)
	if !s.world.IsWalkable(next) {
		return
	}

	player.Pos = next
	if trap, ok := s.world.TrapAt(next); ok {
		s.applyTrap(&player, trap)
	}
	s.players[playerID] = player
}

func (s *Simulation) useItem(playerID PlayerID, slot int) {
	player, ok := s.players[playerID]
	if !ok {
		return
	}

	if !player.UseItem(slot) {
		return
	}

	s.players[playerID] = player
}

func (s *Simulation) applyTrap(player *entity.Player, trap world.Trap) {
	player.Stats.Damage(trap.Damage)
	player.Stats.DrainShield(trap.ShieldDrain)

	trap.Triggered = true
	if trap.OneShot {
		s.world.RemoveTrap(trap.Pos)
		return
	}

	s.world.SetTrap(trap)
}
