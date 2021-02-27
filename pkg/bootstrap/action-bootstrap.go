package bootstrap

import (
	"errors"
	"strings"

	"github.com/birdkid/mouser/pkg/actions"
	"github.com/birdkid/mouser/pkg/config"
)

const (
	toggleOnSuffix  = ":toggle:on"
	toggleOffSuffix = ":toggle:off"
)

type lazyAction struct {
	aRef config.ActionRef
	wip  bool
}

func newLazyAction(aRef config.ActionRef) lazyAction {
	return lazyAction{aRef: aRef}
}

type actionsRepo struct {
	r  map[string]*lazyAction
	as map[string]actions.Action
	s  config.Settings
}

func newActionsRepo(
	aRefs map[string]config.ActionRef,
	s config.Settings,
) actionsRepo {
	r := make(map[string]*lazyAction, len(aRefs))
	for name, aRef := range aRefs {
		la := newLazyAction(aRef)
		r[name] = &la
	}
	return actionsRepo{
		r:  r,
		as: make(map[string]actions.Action),
		s:  s,
	}
}

// get retrieves an action ref into an action.
func (ar actionsRepo) get(aRef config.ActionRef) (actions.Action, error) {
	aci := aRef.A
	switch ac := aci.(type) {
	case config.BasicAction:
		return ar.resolveActionName(ac.Name, ac.Args)
	case config.AppBranchAction:
		return ar.resolveAppBranchAction(ac)
	case nil:
		return nil, nil
	default:
		return nil, errors.New("invalid action type")
	}
}

func (ar actionsRepo) resolveActionName(
	name string,
	args []interface{},
) (actions.Action, error) {
	if a, ok := ar.as[name]; ok {
		return a, nil
	}

	if la, ok := ar.r[name]; ok {
		if la.wip {
			return nil, errors.New("circular action reference found")
		}
		la.wip = true
		aRef := la.aRef
		a, err := ar.get(aRef)
		if err != nil {
			return nil, err
		}
		ar.as[name] = a
		return a, nil
	}

	if baseName := ar.getToggleName(name); baseName != "" {
		if err := ar.resolveToggle(baseName); err != nil {
			return nil, err
		}
		return ar.as[name], nil
	}

	return actions.New(name, args...)
}

func (ar actionsRepo) getToggleName(name string) string {
	if b := strings.TrimSuffix(name, toggleOnSuffix); b != name {
		return b
	} else if b := strings.TrimSuffix(name, toggleOffSuffix); b != name {
		return b
	} else {
		return ""
	}
}

func (ar actionsRepo) resolveToggle(name string) error {
	a, err := ar.resolveActionName(name, nil)
	if err != nil {
		return err
	}
	on, off := actions.NewToggle(
		a,
		ar.s.Toggles.InitDelay.Duration(),
		ar.s.Toggles.RepeatDelay.Duration(),
	)
	onName := name + toggleOnSuffix
	offName := name + toggleOffSuffix
	ar.as[onName] = on
	ar.as[offName] = off
	return nil
}

func (ar actionsRepo) resolveAppBranchAction(
	ac config.AppBranchAction,
) (actions.Action, error) {
	branches := make(map[string]actions.Action, len(ac.Branches))
	for app, aRef := range ac.Branches {
		a, err := ar.get(aRef)
		if err != nil {
			return nil, err
		}
		branches[app] = a
	}

	fallback, err := ar.get(ac.Fallback)
	if err != nil {
		return nil, err
	}

	a := actions.NewAppBranch(branches, fallback)
	return a, nil
}

// gestureAction holds an action to be triggered by a matching gesture series.
type gestureAction struct {
	G gestureMatcher
	A actions.Action
}

func makeGestureAction(
	gac config.GestureAction,
	ar actionsRepo,
) (gestureAction, error) {
	gm, err := makeGestureMatcher(gac)
	if err != nil {
		return gestureAction{}, err
	}

	a, err := ar.get(gac.Action)
	if err != nil {
		return gestureAction{}, err
	}

	return gestureAction{G: gm, A: a}, nil
}