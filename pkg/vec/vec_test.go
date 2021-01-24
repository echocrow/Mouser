package vec_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/birdkid/mouser/pkg/vec"
	"github.com/stretchr/testify/assert"
)

type v2d = vec.Vec2D

type mockEgine struct {
	pos v2d
}

func (e *mockEgine) GetMousePos() v2d {
	return e.pos
}

func TestVec2DSub(t *testing.T) {
	t.Parallel()
	tests := []struct {
		v    v2d
		u    v2d
		want v2d
	}{
		{v2d{0, 0}, v2d{0, 0}, v2d{0, 0}},
		{v2d{1, 0}, v2d{1, 0}, v2d{0, 0}},
		{v2d{2, 3}, v2d{5, 7}, v2d{-3, -4}},
		{v2d{7, 5}, v2d{3, 2}, v2d{4, 3}},
		{v2d{-1, -2}, v2d{3, 4}, v2d{-4, -6}},
		{v2d{-2, -5}, v2d{-3, -7}, v2d{1, 2}},
	}
	subtests := []struct {
		swap bool
		rev  bool
	}{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("Subtraction %d", i), func(t *testing.T) {
			t.Parallel()
			wantReg := tc.want
			for _, stc := range subtests {
				v, u := tc.v, tc.u
				want := wantReg
				if stc.swap {
					v.X, v.Y = v.Y, v.X
					u.X, u.Y = u.Y, u.X
					want.X, want.Y = want.Y, want.X
				}
				if stc.rev {
					v, u = u, v
					want.X, want.Y = -want.X, -want.Y
				}
				got := v.Sub(u)
				assert.Equal(t, want, got)
			}
		})
	}
}

func TestVec2DLen(t *testing.T) {
	t.Parallel()
	tests := []struct {
		v    v2d
		want float64
	}{
		{v2d{0, 0}, 0},
		{v2d{0, 1}, 1},
		{v2d{0, 42}, 42},
		{v2d{3, 4}, 5},
		{v2d{32, 255}, 257},
	}
	const (
		swap = 1 << iota
		invX
		invY
		stLen
	)
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("Lenth %d", i), func(t *testing.T) {
			t.Parallel()
			wantReg := tc.want
			for i := 0; i < stLen; i++ {
				swap := i & 1
				invX := i & 2
				invY := i & 4
				v := tc.v
				want := wantReg
				if i&swap != 0 {
					v.X, v.Y = v.Y, v.X
				}
				if i&invX != 0 {
					v.X = -v.X
				}
				if i&invY != 0 {
					v.Y = -v.Y
				}
				got := v.Len()
				assert.Equal(t, want, got)
			}
		})
	}
}

func TestVec2DRot(t *testing.T) {
	prec := .1e-5
	t.Parallel()
	tests := []struct {
		v       v2d
		rad     float64
		want    v2d
		precise bool
	}{
		{
			v2d{0, 0}, 0,
			v2d{0, 0}, true,
		},
		{
			v2d{12.34, -34.78}, 0,
			v2d{12.34, -34.78}, true,
		},
		{
			v2d{1, 0}, math.Pi / 2,
			v2d{0, 1}, true,
		},
		{
			v2d{2, 3}, math.Pi / 2,
			v2d{-3, 2}, true,
		},
		{
			v2d{4, 5}, math.Pi * 3 / 2,
			v2d{5, -4}, true,
		},
		{
			v2d{-11, -13}, math.Pi,
			v2d{11, 13}, true,
		},
		{
			v2d{1, 1}, math.Pi / 4,
			v2d{0, math.Sqrt2}, false,
		},
		{
			v2d{123, 0}, math.Pi / 3,
			v2d{123 * .5, 123 * math.Sqrt(3) / 2}, false,
		},
	}
	for i, tc := range tests {
		tc := tc
		for m := -2; m <= 2; m++ {
			m := m
			radD := float64(m) * math.Pi
			rad := tc.rad + radD
			deg := rad / (math.Pi) * 180
			t.Run(fmt.Sprintf("Rotation #%d (%.1fdeg)", i, deg), func(t *testing.T) {
				t.Parallel()
				want := tc.want
				flipped := m%2 == 1 || m%2 == -1
				if flipped {
					want = v2d{-want.X, -want.Y}
				}
				wantLen := tc.want.Len()
				got := tc.v.Rot(rad)
				gotLen := got.Len()
				if tc.precise {
					assert.Equal(t, want, got)
					assert.Equal(t, wantLen, gotLen)
				} else {
					assert.InDelta(t, want.X, got.X, prec)
					assert.InDelta(t, want.Y, got.Y, prec)
					assert.InDelta(t, wantLen, gotLen, prec)
				}
			})
		}
	}
}
