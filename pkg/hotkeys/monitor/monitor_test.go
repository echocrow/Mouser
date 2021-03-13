package monitor_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/monitor"
	"github.com/birdkid/mouser/pkg/hotkeys/monitor/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEngineFiddler func(initOk, runOk bool)

func newMockEngine() (*mocks.Engine, mockEngineFiddler) {
	e := new(mocks.Engine)
	eInitOk := true
	eRunOk := true
	setOks := func(initOk, runOk bool) {
		eInitOk = initOk
		eRunOk = runOk
	}
	e.On("Init").Return(func() bool {
		return eInitOk
	})
	e.On("Start", mock.Anything).Return(
		func(*monitor.Monitor) error {
			if !eRunOk {
				return errors.New("some error")
			} else {
				return nil
			}
		},
	)
	e.On("Stop").Return()
	e.On("Deinit").Return(func() bool {
		return eInitOk
	})
	e.On("SetLogCb", mock.Anything).Return()
	return e, setOks
}

func newMockMonitor(e monitor.Engine) *monitor.Monitor {
	if e == nil {
		e, _ = newMockEngine()
	}
	m := monitor.New(nil, e)
	return m
}

func TestMonitorStartStop(t *testing.T) {
	t.Parallel()
	type monitorActions []struct {
		stop, eInitOk, eRunOk, wantErr                bool
		initCount, startCount, stopCount, deinitCount int
	}
	tests := []struct {
		name    string
		actions monitorActions
	}{
		{
			"starts and stops",
			monitorActions{
				{false, true, true, false, 1, 1, 0, 0},
				{true, true, true, false, 1, 1, 1, 1},
				{false, true, true, false, 2, 2, 1, 1},
				{true, true, true, false, 2, 2, 2, 2},
			},
		},
		{
			"fails on double start",
			monitorActions{
				{false, true, true, false, 1, 1, 0, 0},
				{false, true, true, true, 1, 1, 0, 0},
			},
		},
		{
			"fails on cold stop",
			monitorActions{
				{true, true, true, true, 0, 0, 0, 0},
			},
		},
		{
			"fails with faulty engine start",
			monitorActions{
				{false, false, false, true, 1, 0, 0, 0},
			},
		},
		{
			"deinits after faulty engine run",
			monitorActions{
				{false, true, false, true, 1, 1, 0, 1},
			},
		},
		{
			"fails with faulty engine stop",
			monitorActions{
				{false, true, true, false, 1, 1, 0, 0},
				{true, false, false, true, 1, 1, 1, 1},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e, setEngineOks := newMockEngine()
			m := newMockMonitor(e)
			for _, act := range tc.actions {
				setEngineOks(act.eInitOk, act.eRunOk)
				var err error
				if !act.stop {
					_, err = m.Start()
				} else {
					err = m.Stop()
				}
				if act.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				e.AssertNumberOfCalls(t, "Init", act.initCount)
				e.AssertNumberOfCalls(t, "Start", act.startCount)
				e.AssertNumberOfCalls(t, "Stop", act.stopCount)
				e.AssertNumberOfCalls(t, "Deinit", act.deinitCount)
			}
		})
	}
}

func TestMonitorLog(t *testing.T) {
	t.Parallel()
	m := newMockMonitor(nil)
	logs := []string{}
	mockLogCb := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logs = append(logs, strings.ToLower(msg))
	}
	m.SetLogCb(mockLogCb)
	m.Start()
	m.Stop()
	startLogged := false
	stopLogged := false
	for _, msg := range logs {
		startLogged = startLogged || strings.Contains(msg, "start")
		stopLogged = stopLogged || strings.Contains(msg, "stop")
	}
	assert.Equal(t, true, startLogged, "expeced monitor start log")
	assert.Equal(t, true, stopLogged, "expeced monitor stop log")
}

// This test currently causes a data race in Testify.
// @see https://github.com/stretchr/testify/issues/625
// func TestMonitorConcurrency(t *testing.T) {
// 	t.Parallel()
// 	e, _ := newMockEngine()
// 	m := newMockMonitor(e)
// 	iterCount := 100
// 	var wg sync.WaitGroup
// 	wg.Add(iterCount)
// 	for i := 0; i < iterCount; i++ {
// 		go func() {
// 			defer wg.Done()
// 			m.Start()
// 		}()
// 	}
// 	wg.Wait()
// 	e.AssertNumberOfCalls(t, "Start", 1)
// }

func TestMonitorEvents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		events []monitor.HotkeyEvent
	}{
		{
			"shares no events",
			[]monitor.HotkeyEvent{},
		},
		{
			"shares all events",
			[]monitor.HotkeyEvent{
				{1, true, time.Time{}},
				{1, false, time.Time{}},
				{2, true, time.Time{}},
				{2, false, time.Time{}},
				{1, true, time.Time{}},
				{1, false, time.Time{}},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newMockMonitor(nil)
			eventsCh, err := m.Start()
			if assert.NoError(t, err) {
				gotEvents := []monitor.HotkeyEvent{}
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					for event := range eventsCh {
						gotEvents = append(gotEvents, event)
					}
					wg.Done()
				}()
				for _, event := range tc.events {
					m.Dispatch(event)
				}
				m.Stop()
				wg.Wait()
				wantEvents := tc.events
				assert.Equal(t, gotEvents, wantEvents, "invalid events list")
			}
		})
	}
}
