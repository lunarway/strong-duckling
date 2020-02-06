package daemon_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lunarway/strong-duckling/internal/daemon"
	"github.com/stretchr/testify/assert"
)

func TestDaemon_Loop(t *testing.T) {
	tickInterval := 10 * time.Millisecond
	testDuration := 100 * time.Millisecond
	expectedTickCount := int32(10)

	var actualTickCount int32
	d := daemon.New(daemon.Configuration{
		Interval: tickInterval,
		Reporter: &daemon.Reporter{
			Started: func(time.Duration) {
				t.Log("Daemon started")
			},
			Stopped: func() {
				t.Log("Daemon stopped")
			},
			Skipped: func() {
				t.Log("Daemon skipped")
			},
			Ticked: func() {
				t.Log("Daemon ticked")
			},
		},
		Tick: func() {
			atomic.AddInt32(&actualTickCount, 1)
		},
	})

	shutdown := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		d.Loop(shutdown)
	}()

	// let the loop run a couple of cycles
	time.Sleep(testDuration)

	close(shutdown)

	// wait for wait group to be done but limit the time with a timeout in case
	// the loop never stops.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("Loop did not stop on shutdown signal")
	case <-done:
	}

	// assert that we tick multiple times with at most 20% off of the expected
	// count. The exact number is not important as we only want to verify that the
	// loop actually loops.
	assert.InEpsilon(t, expectedTickCount, actualTickCount, 0.2, "tick count %d not as the expected %d", actualTickCount, expectedTickCount)
}
