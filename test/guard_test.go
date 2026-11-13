package test

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var kept [][]byte

func TestWatchMemory_saysSoOnceTheHeapPassesTheLimit(t *testing.T) {
	over := make(chan uint64, 1)
	stop := WatchMemory(1<<20, time.Millisecond, func(heap uint64) {
		select {
		case over <- heap:
		default:
		}
	})
	defer stop()
	kept = append(kept, make([]byte, 64<<20))
	select {
	case heap := <-over:
		if heap <= 1<<20 {
			t.Errorf("it said %d bytes, under the limit", heap)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("64 MiB held past a limit of 1 MiB, and it said nothing")
	}
	kept = nil
}

func TestWatchMemory_saysNothingUnderTheLimit(t *testing.T) {
	said := make(chan uint64, 1)
	stop := WatchMemory(1<<40, time.Millisecond, func(heap uint64) { said <- heap })
	time.Sleep(20 * time.Millisecond)
	stop()
	select {
	case heap := <-said:
		t.Errorf("it said %d bytes, under a limit of 1 TiB", heap)
	default:
	}
}

type endless struct{}

func (endless) Run() int {
	for {
		kept = append(kept, make([]byte, 1<<20))
		time.Sleep(time.Millisecond)
	}
}

func TestGuarded_endsTheRunWhoseHeapPassesTheLimit(t *testing.T) {
	if os.Getenv("CAIRN_GUARDED_CHILD") == "1" {
		os.Exit(Guarded(endless{}, 16<<20))
	}
	run := exec.Command(os.Args[0], "-test.run=^TestGuarded_endsTheRunWhoseHeapPassesTheLimit$")
	run.Env = append(os.Environ(), "CAIRN_GUARDED_CHILD=1")
	out, err := run.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("a run past its limit ended with %v, want exit 3\n%s", err, out)
	}
	if !strings.Contains(string(out), "a loop that never ends") {
		t.Errorf("it does not say why it ended:\n%s", out)
	}
}

type passing struct{}

func (passing) Run() int { return 0 }

func TestGuarded_answersWhatTheTestsAnswered(t *testing.T) {
	if got := Guarded(passing{}, 1<<40); got != 0 {
		t.Errorf("Guarded answered %d for tests that passed", got)
	}
}
