package entity

import (
	"testing"

	gamemath "ascii-game/internal/math"
)

func TestStatsDamageUsesShieldFirst(t *testing.T) {
	stats := Stats{
		Health: 100,
		Shield: 20,
		MaxHP:  100,
		MaxSP:  50,
	}

	stats.Damage(30)

	if stats.Health != 90 {
		t.Fatalf("health = %d, want 90", stats.Health)
	}
	if stats.Shield != 0 {
		t.Fatalf("shield = %d, want 0", stats.Shield)
	}
}

func TestPlayerUseItemConsumesInventoryAndCapsStats(t *testing.T) {
	player := NewPlayer(gamemath.Vec2{X: 1, Y: 1})
	player.Stats.Health = 80
	player.Stats.Shield = 40

	if !player.UseItem(InventorySlotHealth) {
		t.Fatal("expected health item use to succeed")
	}
	if player.Stats.Health != 100 {
		t.Fatalf("health = %d, want 100", player.Stats.Health)
	}
	if player.Inventory[InventorySlotHealth].Quantity != 2 {
		t.Fatalf("health item quantity = %d, want 2", player.Inventory[InventorySlotHealth].Quantity)
	}

	if !player.UseItem(InventorySlotShield) {
		t.Fatal("expected shield item use to succeed")
	}
	if player.Stats.Shield != 50 {
		t.Fatalf("shield = %d, want 50", player.Stats.Shield)
	}
	if player.Inventory[InventorySlotShield].Quantity != 1 {
		t.Fatalf("shield item quantity = %d, want 1", player.Inventory[InventorySlotShield].Quantity)
	}
}

func TestPlayerUseItemFailsWhenDead(t *testing.T) {
	player := NewPlayer(gamemath.Vec2{})
	player.Stats.Health = 0

	if player.UseItem(InventorySlotHealth) {
		t.Fatal("expected dead player item use to fail")
	}
}
