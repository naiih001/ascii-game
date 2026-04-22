package proto

import (
	"testing"

	gamemath "ascii-game/internal/math"
	"ascii-game/internal/sim"
	"ascii-game/internal/world"
)

func TestSnapshotRoundTripIncludesHUDAndTrapState(t *testing.T) {
	original := sim.Snapshot{
		Tick: 42,
		Players: []sim.PlayerState{
			{
				ID:              7,
				Pos:             gamemath.Vec2{X: 12, Y: 8},
				Sprite:          '@',
				Health:          90,
				Shield:          35,
				MaxHP:           100,
				MaxSP:           50,
				HealthItemCount: 2,
				ShieldItemCount: 1,
			},
		},
		Traps: []sim.TrapState{
			{
				Pos:  gamemath.Vec2{X: 18, Y: 12},
				Type: world.TrapSpike,
			},
			{
				Pos:  gamemath.Vec2{X: 42, Y: 40},
				Type: world.TrapEMP,
			},
		},
	}

	decoded, err := UnmarshalSnapshot(MarshalSnapshot(original))
	if err != nil {
		t.Fatalf("round-trip snapshot failed: %v", err)
	}

	if decoded.Tick != original.Tick {
		t.Fatalf("tick = %d, want %d", decoded.Tick, original.Tick)
	}
	if len(decoded.Players) != 1 {
		t.Fatalf("player count = %d, want 1", len(decoded.Players))
	}
	player := decoded.Players[0]
	if player.Health != 90 || player.Shield != 35 || player.HealthItemCount != 2 || player.ShieldItemCount != 1 {
		t.Fatalf("decoded player = %+v", player)
	}
	if len(decoded.Traps) != 2 {
		t.Fatalf("trap count = %d, want 2", len(decoded.Traps))
	}
	if decoded.Traps[1].Type != world.TrapEMP {
		t.Fatalf("trap type = %d, want %d", decoded.Traps[1].Type, world.TrapEMP)
	}
}

func TestUseItemMessageRoundTrip(t *testing.T) {
	payload, err := MarshalUseItem(1)
	if err != nil {
		t.Fatalf("marshal use item: %v", err)
	}

	slot, err := UnmarshalUseItem(payload)
	if err != nil {
		t.Fatalf("unmarshal use item: %v", err)
	}

	if slot != 1 {
		t.Fatalf("slot = %d, want 1", slot)
	}
}
