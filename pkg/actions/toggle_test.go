package actions_test

import (
	"sync"
	"testing"

	"github.com/echocrow/Mouser/pkg/actions"
	"github.com/stretchr/testify/assert"
)

func TestCreateToggle(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	var on actions.Action
	var off actions.Action

	mu := sync.Mutex{}
	called := uint(0)
	maxCalls := uint(3)
	stopped := false
	action := func() {
		mu.Lock()
		defer mu.Unlock()
		if !stopped {
			called++
			if called >= maxCalls {
				stopped = true
				off()
				done <- struct{}{}
			}
		}
	}

	on, off = actions.NewToggle(action, 0, 0)
	assert.NotNil(t, on, `want "on" action`)
	assert.NotNil(t, off, `want "off" action`)

	assert.Equal(t, uint(0), called, `want no calls before "on"`)
	on()
	<-done
	assert.Equal(t, maxCalls, called, "want right number of action calls")
}
