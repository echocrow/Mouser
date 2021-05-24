package bootstrap

import (
	"errors"

	"github.com/echocrow/Mouser/pkg/config"
	"github.com/echocrow/Mouser/pkg/hotkeys/gestures"
)

// gestureMatcher describes a gestures series validator.
type gestureMatcher interface {
	matches([]gestures.Gesture) bool
}

// singleGestureTailMatch matches the last gesture of a gesture series.
type singleGestureTailMatch struct {
	G gestures.Gesture
}

// matches checks whether gm matches the given gesture series.
func (gm singleGestureTailMatch) matches(gs []gestures.Gesture) bool {
	return gestures.EndsIn(gs, gm.G)
}

// gestureTailMatch matches the last few gestures of a gesture series.
type gestureTailMatch struct {
	Gs []gestures.Gesture
}

// matches checks whether gm matches the given gesture series.
func (gm gestureTailMatch) matches(gs []gestures.Gesture) bool {
	l := len(gm.Gs)
	gsl := len(gs)
	return gsl >= l && gestures.Match(gs[gsl-l:gsl], gm.Gs)
}

// fullMatch matches the all gestures of a gesture series.
type fullMatch struct {
	Gs []gestures.Gesture
}

// matches checks whether gm matches the given gesture series.
func (gm fullMatch) matches(gs []gestures.Gesture) bool {
	return gestures.Match(gs, gm.Gs)
}

func makeGestures(gStrs []string) []gestures.Gesture {
	gs := make([]gestures.Gesture, len(gStrs))
	for i, gStr := range gStrs {
		gs[i] = gestures.Gesture(gStr)
	}
	return gs
}

func makeGestureMatcher(
	gac config.GestureAction,
) (gestureMatcher, error) {
	gs := makeGestures(gac.Gesture)
	l := len(gs)
	if l == 0 {
		return nil, errors.New("missing gestures in gesture action config")
	}
	if gac.Exact {
		return fullMatch{gs}, nil
	} else if l == 1 {
		return singleGestureTailMatch{gs[0]}, nil
	} else {
		return gestureTailMatch{gs}, nil
	}
}
