package hotkey_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/echocrow/Mouser/pkg/hotkeys/hotkey"
	"github.com/echocrow/Mouser/pkg/hotkeys/hotkey/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type hkID = hotkey.ID

func newMockIDProvider(staticID hkID) hotkey.IDProvider {
	idp := new(mocks.IDProvider)
	idp.On("NextID").Return(staticID)
	return idp
}

func newMockEngine(t *testing.T, engineOk bool, staticID hkID) hotkey.Engine {
	e := new(mocks.Engine)
	e.On("Register", mock.Anything, mock.Anything).Return(
		func(id hkID, key hotkey.KeyName) error {
			if !engineOk {
				return hotkey.ErrRegistrationFailed
			}
			return nil
		},
	)
	e.On("Unregister").Return()
	e.On("IDFromEvent", mock.Anything).Return(
		func(eEvent hotkey.EngineEvent) hkID {
			if !engineOk {
				return hotkey.NoID
			}
			return staticID
		},
		func(eEvent hotkey.EngineEvent) error {
			if !engineOk {
				return errors.New("some error")
			}
			return nil
		},
	)
	return e
}

func TestRegistry(t *testing.T) {
	t.Run("only requires engine", func(t *testing.T) {
		r := hotkey.NewRegistry(newMockEngine(t, true, 0), nil)
		_, err := r.Add("f1")
		assert.NoError(t, err)
	})

	t.Run("fails with faulty engine", func(t *testing.T) {
		r := hotkey.NewRegistry(newMockEngine(t, false, 0), nil)
		_, err := r.Add("f1")
		assert.Error(t, err)
	})

	t.Run("returns next ID on success", func(t *testing.T) {
		staticID := hkID(1)
		idp := newMockIDProvider(staticID)
		tests := []struct {
			keyName  hotkey.KeyName
			engineOk bool
			wantOK   bool
			wantID   hkID
		}{
			{"somekey", true, true, staticID},
			{"invalidkeyname", false, false, hotkey.NoID},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(fmt.Sprint(tc.keyName), func(t *testing.T) {
				t.Parallel()
				r := hotkey.NewRegistry(newMockEngine(t, tc.engineOk, staticID), idp)
				gotID, err := r.Add(tc.keyName)
				assert.Equal(t, tc.wantID, gotID)
				if tc.wantOK {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

func TestIDFromEvent(t *testing.T) {
	staticID := hkID(1)
	mockEngineEvent := hotkey.MockEngineEvent()
	tests := []struct {
		eEvent   hotkey.EngineEvent
		engineOk bool
		wantOK   bool
		wantID   hkID
	}{
		{mockEngineEvent, true, true, staticID},
		{mockEngineEvent, false, false, hotkey.NoID},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			r := hotkey.NewRegistry(newMockEngine(t, tc.engineOk, staticID), nil)
			gotID, err := r.IDFromEvent(tc.eEvent)
			assert.Equal(t, tc.wantID, gotID)
			if tc.wantOK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestIDCounter(t *testing.T) {
	t.Run("starts at 1", func(t *testing.T) {
		idc := hotkey.NewIDCounter()
		assert.Equal(t, hkID(1), idc.NextID())
	})

	t.Run("is concurrency-safe", func(t *testing.T) {
		idc := hotkey.NewIDCounter()
		iterCount := 100
		var wg sync.WaitGroup
		wg.Add(iterCount)
		for i := 1; i <= iterCount; i++ {
			go func() {
				defer wg.Done()
				idc.NextID()
			}()
		}
		wg.Wait()
		assert.Equal(t, hkID(iterCount+1), idc.NextID())
	})
}

func TestIIDSet(t *testing.T) {
	s := hotkey.NewIDSet()
	hkID1 := hkID(1)
	hkID2 := hkID(2)
	tests := []struct {
		name string
		adds []hkID
		dels []hkID
		has  [2]bool
	}{
		{"starts empty", []hkID{}, []hkID{}, [2]bool{false, false}},
		{"adds #1", []hkID{hkID1}, []hkID{}, [2]bool{true, false}},
		{"adds #2", []hkID{hkID2}, []hkID{}, [2]bool{true, true}},
		{"deletes #1", []hkID{}, []hkID{hkID1}, [2]bool{false, true}},
		{"double-adds", []hkID{hkID2}, []hkID{}, [2]bool{false, true}},
		{"double-deletes", []hkID{}, []hkID{hkID1}, [2]bool{false, true}},
		{"re-adds", []hkID{hkID1}, []hkID{}, [2]bool{true, true}},
		{"deletes #2", []hkID{}, []hkID{hkID2}, [2]bool{true, false}},
		{"re-deletes", []hkID{}, []hkID{hkID1}, [2]bool{false, false}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, id := range tc.adds {
				s.Add(id)
			}
			for _, id := range tc.dels {
				s.Del(id)
			}
			assert.Equal(t, tc.has[0], s.Has(hkID1))
			assert.Equal(t, tc.has[1], s.Has(hkID2))
		})
	}
}
