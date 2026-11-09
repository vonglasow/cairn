package gui

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cairn/internal/build"
	"cairn/internal/engine"
)

// ErrShownElsewhere is returned when a window is opened on a store another window already shows, which is brought forward instead.
var ErrShownElsewhere = errors.New("another window shows this store: it is brought forward instead")

// Windows is where each window open on this machine answers for the store it shows: a socket a store, named after the store, in a folder of this person's.
type Windows struct{ Dir string }

const (
	runtimeVariable = "XDG_RUNTIME_DIR"
	answerWithin    = 2 * time.Second
	raiseSaid       = "raise"
	raisedSaid      = "ok"
)

// WindowsHere is the folder the windows of this person answer in: XDG_RUNTIME_DIR when it is absolute, the system's temporary folder otherwise.
func WindowsHere(lookup func(string) (string, bool)) Windows {
	if dir, ok := lookup(runtimeVariable); ok && filepath.IsAbs(dir) {
		return Windows{Dir: dir}
	}
	return Windows{Dir: os.TempDir()}
}

func (ws Windows) socket(store string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(store)))
	return filepath.Join(ws.Dir, "cairn-"+hex.EncodeToString(sum[:8])+".window")
}

func (ws Windows) dial(store string) (net.Conn, error) {
	return net.DialTimeout("unix", ws.socket(store), answerWithin)
}

// Shows says whether a window answers for the store.
func (ws Windows) Shows(store string) bool {
	conn, err := ws.dial(store)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Raise asks the window showing the store to come forward, and says whether one did.
func (ws Windows) Raise(store string) bool {
	conn, err := ws.dial(store)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(answerWithin))
	if _, err := io.WriteString(conn, raiseSaid+"\n"); err != nil {
		return false
	}
	said, _ := bufio.NewReader(conn).ReadString('\n')
	return strings.TrimSpace(said) == raisedSaid
}

// Claim makes this window the one answering for the store, raise called each time another asks it to come forward; a window already answering for it is brought forward instead, and ErrShownElsewhere returned. What a window gone left behind is claimed over.
func (ws Windows) Claim(store string, raise func()) (io.Closer, error) {
	if ws.Raise(store) {
		return nil, ErrShownElsewhere
	}
	path := ws.socket(store)
	_ = os.Remove(path)
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	go answering(listener, raise)
	return listener, nil
}

func answering(listener net.Listener, raise func()) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go answer(conn, raise)
	}
}

func answer(conn net.Conn, raise func()) {
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(answerWithin))
	said, _ := bufio.NewReader(conn).ReadString('\n')
	if strings.TrimSpace(said) != raiseSaid {
		return
	}
	raise()
	_, _ = io.WriteString(conn, raisedSaid+"\n")
}

// Reach says what a window of its own on another workspace, folder or store would hold, and how a cairn started for it names it.
type Reach func(engine.Elsewhere) (engine.Apart, error)

// Starter starts a cairn with these arguments, apart from this one, and does not wait for it.
type Starter func(args ...string) error

// Went is what came of asking for a window elsewhere: the one showing it brought forward, a cairn started to show it, or why neither.
type Went struct {
	Raised bool
	Err    error
}

// OpeningApart brings forward the window showing what is asked for, or starts a cairn to show it in a window of its own.
func OpeningApart(to engine.Elsewhere, reach Reach, ws Windows, start Starter) Went {
	apart, err := reach(to)
	if err != nil {
		return Went{Err: err}
	}
	if ws.Raise(apart.Store) {
		return Went{Raised: true}
	}
	return Went{Err: start(append([]string{"gui"}, apart.Flags...)...)}
}

// StartCairn starts the cairn running with these arguments, its own process, and does not wait for it.
func StartCairn(args ...string) error {
	running, err := build.Running()
	if err != nil {
		return err
	}
	command := exec.Command(running, args...)
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
