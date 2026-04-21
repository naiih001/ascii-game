package math

type Vec2 struct {
	X int
	Y int
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}
