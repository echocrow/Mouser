// Package vec implements a bacis 2D vector.
package vec

import "math"

// Vec2D is a two-dimensional vector.
type Vec2D struct {
	X float64
	Y float64
}

// Sub calculates a new Vec2D at v-u.
func (v Vec2D) Sub(u Vec2D) Vec2D {
	return Vec2D{
		v.X - u.X,
		v.Y - u.Y,
	}
}

// Len returns the positive length of the vector.
func (v Vec2D) Len() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2))
}

// Rot calculates a new Vec2D of v rotated by rad radians.
func (v Vec2D) Rot(rad float64) Vec2D {
	x, y := v.X, v.Y
	if (x == 0 && y == 0) || rad == 0 {
		return v
	}

	rad = posMod(rad, math.Pi*2)
	if rad == math.Pi/2 {
		return Vec2D{-y, x}
	} else if rad == math.Pi {
		return Vec2D{-x, -y}
	} else if rad == math.Pi*3/2 {
		return Vec2D{y, -x}
	}

	r := math.Sqrt(x*x + y*y)
	t := math.Atan2(y, x)

	t = t + rad

	x = math.Cos(t) * r
	y = math.Sin(t) * r

	p := 10e5
	x = math.Round(x*p) / p
	y = math.Round(y*p) / p

	return Vec2D{x, y}
}

func posMod(x, y float64) (remainder float64) {
	remainder = math.Mod(x, y)
	if remainder < 0 {
		remainder = remainder + y
	}
	return remainder
}
