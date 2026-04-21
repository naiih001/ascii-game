package entity

import gamemath "ascii-game/internal/math"

type Player struct {
	Entity
}

func NewPlayer(pos gamemath.Vec2) Player {
	return Player{
		Entity: Entity{
			Pos:    pos,
			Sprite: '@',
		},
	}
}
