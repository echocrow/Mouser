package hotkey_test

import (
	"fmt"
	"testing"

	"github.com/echocrow/Mouser/pkg/hotkeys/hotkey"
	"github.com/stretchr/testify/assert"
)

func TestNameToCode(t *testing.T) {
	tests := []struct {
		keyName     hotkey.KeyName
		wantOk      bool
		wantKeyCode interface{}
	}{
		{"", false, nil},
		{"invalidkeyname", false, nil},

		{"f1", true, 0},
		{"f20", true, 0},
		{"f21", false, nil},

		{"mouse1", false, nil},
		{"mouse2", false, nil},
		{"mouse3", true, 0},
		{"mouse5", true, 0},
		{"mouse6", false, nil},
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
