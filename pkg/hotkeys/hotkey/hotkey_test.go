package hotkey_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/birdkid/mouser/pkg/hotkeys/hotkey/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type hkID = hotkey.ID

func newMockIDProvider(staticID hkID) hotkey.IDProvider {
	idp := new(mocks.IDProvider)
	idp.On("NextID").Return(staticID)
	return idp
}

func newMockEngine(t *testing.T, engineOk bool) hotkey.Engine {
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
	return e
}

func TestRegistry(t *testing.T) {
	t.Run("only requires engine", func(t *testing.T) {
		r := hotkey.NewRegistry(newMockEngine(t, true), nil)
		_, err := r.Add("f1")
		assert.NoError(t, err)
	})

	t.Run("fails with faulty engine", func(t *testing.T) {
		r := hotkey.NewRegistry(newMockEngine(t, false), nil)
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
				r := hotkey.NewRegistry(newMockEngine(t, tc.engineOk), idp)
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
