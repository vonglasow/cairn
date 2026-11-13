package test

import (
	"fmt"
	"os"
	"runtime/metrics"
	"time"
)

const heapInUse = "/memory/classes/heap/objects:bytes"

// WatchMemory calls over, every so often, while the heap holds more than limit bytes, until stop is called.
func WatchMemory(limit uint64, every time.Duration, over func(heap uint64)) (stop func()) {
	done := make(chan struct{})
	ended := make(chan struct{})
	go func() {
		defer close(ended)
		sample := []metrics.Sample{{Name: heapInUse}}
		tick := time.NewTicker(every)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				metrics.Read(sample)
				if heap := sample[0].Value.Uint64(); heap > limit {
					over(heap)
				}
			}
		}
	}()
	return func() {
		close(done)
		<-ended
	}
}

// Guarded runs a package's tests and ends the process the moment its heap passes limit bytes, so a loop that never ends fails its test rather than taking the machine's memory.
func Guarded(m interface{ Run() int }, limit uint64) int {
	stop := WatchMemory(limit, 50*time.Millisecond, func(heap uint64) {
		fmt.Fprintf(os.Stderr, "the heap holds %d bytes, past the %d the tests may use: a loop that never ends\n", heap, limit)
		os.Exit(3)
	})
	defer stop()
	return m.Run()
}
