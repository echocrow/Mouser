package hotkey_test

import (
	"fmt"
	"testing"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/stretchr/testify/assert"
)

func TestNameToCode(t *testing.T) {
	tests := []struct {
		keyName     hotkey.KeyName
		wantOk      bool
		wantKeyCode hotkey.KeyCode
	}{
		{"", false, hotkey.NotAKeyCode},
		{"invalidkeyname", false, hotkey.NotAKeyCode},
		{"f1", true, 0},
		{"f20", true, 0},
		{"f21", false, hotkey.NotAKeyCode},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("key \"%s\"", tc.keyName), func(t *testing.T) {
			t.Parallel()
			gotKeyCode, err := hotkey.NameToCode(tc.keyName)
			if tc.wantOk {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.wantKeyCode, gotKeyCode)
				assert.Error(t, err)
			}
		})
	}
}
