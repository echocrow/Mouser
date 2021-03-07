package hotkey_test

import (
	"fmt"
	"testing"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/stretchr/testify/assert"
)

func TestNameToIndex(t *testing.T) {
	tests := []struct {
		keyName hotkey.KeyName
		want    hotkey.KeyIndex
		wantErr bool
	}{
		{"", 0, true},
		{"invalidkeyname", 0, true},
		{"f1", 1, false},
		{"f24", 24, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("key \"%s\"", tc.keyName), func(t *testing.T) {
			t.Parallel()
			ans, err := hotkey.NameToIndex(tc.keyName)
			assert.Equal(t, tc.want, ans)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
