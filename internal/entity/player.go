package entity

import gamemath "ascii-game/internal/math"

type Stats struct {
	Health int
	Shield int
	MaxHP  int
	MaxSP  int
}

func (s *Stats) Alive() bool {
	return s.Health > 0
}

func (s *Stats) Damage(amount int) {
	if amount <= 0 {
		return
	}

	if s.Shield > 0 {
		absorb := amount
		if absorb > s.Shield {
			absorb = s.Shield
		}
		s.Shield -= absorb
		amount -= absorb
	}

	s.Health -= amount
	if s.Health < 0 {
		s.Health = 0
	}
}

func (s *Stats) DrainShield(amount int) {
	if amount <= 0 {
		return
	}

	s.Shield -= amount
	if s.Shield < 0 {
		s.Shield = 0
	}
}

func (s *Stats) HealHP(amount int) {
	if amount <= 0 {
		return
	}

	s.Health += amount
	if s.Health > s.MaxHP {
		s.Health = s.MaxHP
	}
}

func (s *Stats) RepairSP(amount int) {
	if amount <= 0 {
		return
	}

	s.Shield += amount
	if s.Shield > s.MaxSP {
		s.Shield = s.MaxSP
	}
}

type ItemType byte

const (
	ItemNone ItemType = iota
	ItemHealth
	ItemShield
)

const (
	InventorySlotHealth = 0
	InventorySlotShield = 1
	InventorySlotCount  = 2
)

type Item struct {
	Type     ItemType
	Quantity int
}

type Player struct {
	Entity
	Stats     Stats
	Inventory [InventorySlotCount]Item
}

func NewPlayer(pos gamemath.Vec2) Player {
	return Player{
		Entity: Entity{
			Pos:    pos,
			Sprite: '@',
		},
		Stats: Stats{
			Health: 100,
			Shield: 50,
			MaxHP:  100,
			MaxSP:  50,
		},
		Inventory: [InventorySlotCount]Item{
			{Type: ItemHealth, Quantity: 3},
			{Type: ItemShield, Quantity: 2},
		},
	}
}

func (p *Player) Alive() bool {
	return p.Stats.Alive()
}

func (p *Player) UseItem(slot int) bool {
	if !p.Alive() || slot < 0 || slot >= len(p.Inventory) {
		return false
	}

	item := &p.Inventory[slot]
	if item.Quantity <= 0 {
		return false
	}

	switch item.Type {
	case ItemHealth:
		p.Stats.HealHP(25)
	case ItemShield:
		p.Stats.RepairSP(15)
	default:
		return false
	}

	item.Quantity--
	return true
}
