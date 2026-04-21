package entity

import gamemath "ascii-game/internal/math"

type Entity struct {
	Pos    gamemath.Vec2
	Sprite rune
}
