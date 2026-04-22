package sim

import (
	"testing"

	"ascii-game/internal/entity"
	gamemath "ascii-game/internal/math"
	"ascii-game/internal/world"
)

func TestMovePlayerTriggersAndRemovesOneShotTrap(t *testing.T) {
	worldMap := &world.Map{
		Tiles: [][]rune{
			[]rune("...."),
			[]rune("...."),
		},
		Width:  4,
		Height: 2,
		Traps:  make(map[gamemath.Vec2]world.Trap),
	}
	trapPos := gamemath.Vec2{X: 2, Y: 0}
	worldMap.AddTrap(trapPos, world.NewSpikeTrap())

	simulation := New(worldMap)
	playerID := simulation.AddPlayer(gamemath.Vec2{X: 1, Y: 0})

	simulation.QueueMove(playerID, gamemath.Vec2{X: 1, Y: 0})
	simulation.Tick()

	playerState, ok := playerByID(simulation.Snapshot(), playerID)
	if !ok {
		t.Fatal("expected player in snapshot")
	}
	if playerState.Pos != trapPos {
		t.Fatalf("player pos = %+v, want %+v", playerState.Pos, trapPos)
	}
	if playerState.Health != 100 || playerState.Shield != 30 {
		t.Fatalf("stats after trap = hp:%d sp:%d, want hp:100 sp:30", playerState.Health, playerState.Shield)
	}
	if _, ok := worldMap.TrapAt(trapPos); ok {
		t.Fatal("expected one-shot trap to be removed")
	}
}

func TestUseItemChangesReplicatedState(t *testing.T) {
	worldMap := &world.Map{
		Tiles: [][]rune{
			[]rune("...."),
		},
		Width:  4,
		Height: 1,
		Traps:  make(map[gamemath.Vec2]world.Trap),
	}

	simulation := New(worldMap)
	playerID := simulation.AddPlayer(gamemath.Vec2{X: 0, Y: 0})

	player := simulation.players[playerID]
	player.Stats.Health = 60
	player.Inventory[entity.InventorySlotHealth].Quantity = 1
	simulation.players[playerID] = player

	simulation.QueueUseItem(playerID, entity.InventorySlotHealth)
	simulation.Tick()

	playerState, ok := playerByID(simulation.Snapshot(), playerID)
	if !ok {
		t.Fatal("expected player in snapshot")
	}
	if playerState.Health != 85 {
		t.Fatalf("health = %d, want 85", playerState.Health)
	}
	if playerState.HealthItemCount != 0 {
		t.Fatalf("health item count = %d, want 0", playerState.HealthItemCount)
	}
}

func TestReusableTrapPersistsAndDrainsShield(t *testing.T) {
	worldMap := &world.Map{
		Tiles: [][]rune{
			[]rune("...."),
			[]rune("...."),
		},
		Width:  4,
		Height: 2,
		Traps:  make(map[gamemath.Vec2]world.Trap),
	}
	trapPos := gamemath.Vec2{X: 1, Y: 1}
	worldMap.AddTrap(trapPos, world.NewEMPTrap())

	simulation := New(worldMap)
	playerID := simulation.AddPlayer(gamemath.Vec2{X: 1, Y: 0})

	simulation.QueueMove(playerID, gamemath.Vec2{X: 0, Y: 1})
	simulation.Tick()

	playerState, ok := playerByID(simulation.Snapshot(), playerID)
	if !ok {
		t.Fatal("expected player in snapshot")
	}
	if playerState.Health != 100 || playerState.Shield != 20 {
		t.Fatalf("stats after emp = hp:%d sp:%d, want hp:100 sp:20", playerState.Health, playerState.Shield)
	}
	if _, ok := worldMap.TrapAt(trapPos); !ok {
		t.Fatal("expected reusable trap to remain")
	}
}

func playerByID(snapshot Snapshot, id PlayerID) (PlayerState, bool) {
	for _, player := range snapshot.Players {
		if player.ID == id {
			return player, true
		}
	}

	return PlayerState{}, false
}
