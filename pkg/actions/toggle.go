package actions

import (
	"sync"
	"time"
)

// NewToggle creates a toggable action.
func NewToggle(
	action Action,
	initDelay time.Duration,
	repeatDelay time.Duration,
) (on, off Action) {
	mu := sync.Mutex{}
	isRunning := false
	stopCh := make(chan struct{})
	on = func() {
		mu.Lock()
		defer mu.Unlock()
		if isRunning {
			return
		}
		isRunning = true
		go func() {
			go action()
			time.Sleep(initDelay)
			for {
				select {
				case <-stopCh:
					return
				default:
					go action()
					time.Sleep(repeatDelay)
				}
			}
		}()
	}
	off = func() {
		mu.Lock()
		defer mu.Unlock()
		if !isRunning {
			return
		}
		isRunning = false
		stopCh <- struct{}{}
	}
	return
}
