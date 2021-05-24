package bootstrap

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/echocrow/Mouser/pkg/actions"
	"github.com/echocrow/Mouser/pkg/config"
	"github.com/echocrow/Mouser/pkg/log"
)

const (
	toggleSuffix    = ":toggle"
	toggleOnSuffix  = toggleSuffix + ":on"
	toggleOffSuffix = toggleSuffix + ":off"
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
func (ar actionsRepo) get(
	aRef config.ActionRef,
) (a actions.Action, name string, err error) {
	aci := aRef.A
	switch ac := aci.(type) {
	case config.BasicAction:
		a, err = ar.resolveActionName(ac.Name, ac.Args)
		name = ac.Name
	case config.AppBranchAction:
		a, err = ar.resolveAppBranchAction(ac)
		name = "(app-branch)"
	case nil:
		return nil, "(empty-action)", nil
	default:
		return nil, "", errors.New("invalid action type")
	}
	return
}

func (ar actionsRepo) getNested(aRef config.ActionRef) (actions.Action, error) {
	a, _, err := ar.get(aRef)
	return a, err
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
			return nil, fmt.Errorf("circular action reference at \"%s\"", name)
		}
		la.wip = true
		aRef := la.aRef
		a, err := ar.getNested(aRef)
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
	a, initDelay, repeatDelay, err := ar.resolveToggleBaseAction(name)
	if err != nil {
		return err
	}

	on, off := actions.NewToggle(a, initDelay, repeatDelay)
	onName := name + toggleOnSuffix
	offName := name + toggleOffSuffix
	ar.as[onName] = on
	ar.as[offName] = off
	return nil
}
func (ar actionsRepo) resolveToggleBaseAction(
	name string,
) (a actions.Action, initDelay, repeatDelay time.Duration, err error) {
	toggleName := name + toggleSuffix
	if la, ok := ar.r[toggleName]; ok {
		aci := la.aRef.A
		if ac, ok := aci.(config.ToggleAction); ok {
			a, err = ar.getNested(ac.Action)
			initDelay = ac.InitDelay.Duration()
			repeatDelay = ac.RepeatDelay.Duration()
		} else {
			err = fmt.Errorf("invalid toggle base action type at \"%s\"", toggleName)
		}
	} else {
		a, err = ar.resolveActionName(name, nil)
		initDelay = ar.s.Toggles.InitDelay.Duration()
		repeatDelay = ar.s.Toggles.RepeatDelay.Duration()
	}
	return
}

func (ar actionsRepo) resolveAppBranchAction(
	ac config.AppBranchAction,
) (actions.Action, error) {
	branches := make(map[string]actions.Action, len(ac.Branches))
	for app, aRef := range ac.Branches {
		a, err := ar.getNested(aRef)
		if err != nil {
			return nil, err
		}
		branches[app] = a
	}

	fallback, err := ar.getNested(ac.Fallback)
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

func newLoggedAction(
	a actions.Action,
	name string,
	logger log.Logger,
) actions.Action {
	return func() {
		logger.Printf(name)
		if a != nil {
			a()
		}
	}
}

func makeGestureAction(
	gac config.GestureAction,
	ar actionsRepo,
	logger log.Logger,
) (gestureAction, error) {
	gm, err := makeGestureMatcher(gac)
	if err != nil {
		return gestureAction{}, err
	}

	a, aName, err := ar.get(gac.Action)
	if err != nil {
		return gestureAction{}, err
	} else if logger != nil {
		a = newLoggedAction(a, aName, logger)
	}

	return gestureAction{G: gm, A: a}, nil
}
