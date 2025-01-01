package actions

import (
	"os"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/v4/process"
)

// NewRequireApp creates an app-dependent action branch.
func NewRequireApp(
	app string,
	do Action,
	fallback Action,
) Action {
	return NewRequireAppCustom(
		newAppRunningChecker(app).run,
		do,
		fallback,
	)
}

// NewRequireAppCustom creates an app-based actions branch based on a custom
// app-detection callback.
func NewRequireAppCustom(
	checkAppRunning func() bool,
	do Action,
	fallback Action,
) Action {
	return func() {
		action := do
		if !checkAppRunning() {
			action = fallback
		}
		if action != nil {
			action()
		}
	}
}

var osPathSep = string(os.PathSeparator)

type appRunningChecker struct {
	pid int32
	app string
	mx  sync.RWMutex
}

func newAppRunningChecker(app string) *appRunningChecker {
	arc := new(appRunningChecker)
	arc.app = app
	return arc
}

func (arc *appRunningChecker) checkCache() bool {
	arc.mx.RLock()
	defer arc.mx.RUnlock()
	if arc.pid != 0 {
		path := getPidPath(arc.pid)
		if matchAppPathPrefix(path, arc.app, osPathSep) {
			return true
		} else {
			arc.pid = 0
		}
	}
	return false
}

func (arc *appRunningChecker) run() bool {
	if arc.checkCache() {
		return true
	}

	// Scan processes.
	ps, _ := process.Processes()
	// We presume on average user-related apps are closer to the end than to the
	// start of the list of processes, thus we loop through in reverse.
	for i := len(ps) - 1; i >= 0; i-- {
		p := ps[i]
		pid := p.Pid
		path := getProcessPath(p)
		if matchAppPathPrefix(path, arc.app, osPathSep) {
			arc.mx.Lock()
			defer arc.mx.Unlock()
			arc.pid = pid
			return true
		}
	}

	return false
}

func getPidPath(pid int32) string {
	p, _ := process.NewProcess(pid)
	return getProcessPath(p)
}

func getProcessPath(p *process.Process) string {
	path, _ := p.Exe()
	return path
}

func matchAppPathPrefix(path, name, pathSep string) bool {
	return name != "" && (path == name || strings.HasPrefix(path, name+pathSep))
}
