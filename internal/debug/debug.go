// Package debug writes down what cairn did while debug mode is on, in a file a person can send.
package debug

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Variable is the environment variable that turns debug mode on for one launch.
const Variable = "CAIRN_DEBUG"

// Kept is how many files the debug folder keeps: past it, the oldest go.
const Kept = 20

// Most is how much one file holds before it is cut.
const Most = 10 << 20

const cut = "# cut here: the file reached 10 MB\n"

// Log is one launch's file: what is written to it goes down until it holds Most, and then the cut is said once and nothing more is kept. Every text the person wrote reaches it masked, with a salt of its own.
type Log struct {
	mu      sync.Mutex
	file    *os.File
	written int
	most    int
	cut     bool
	left    bool
	salt    []byte
	now     func() time.Time
}

// Open makes the folder if it is not there, lets the oldest files go so that it holds Kept with the new one, and opens a file of its own for this launch of the surface, named by the moment it opened.
func Open(dir, surface string, at time.Time) (*Log, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	if err := keepingRoom(dir); err != nil {
		return nil, err
	}
	named := at.Format("20060102-150405.000") + "-" + strings.Join(strings.Fields(surface), "-") + "-*.log"
	file, err := os.CreateTemp(dir, named)
	if err != nil {
		return nil, err
	}
	l := &Log{file: file, most: Most, salt: make([]byte, 16), now: time.Now}
	_, _ = rand.Read(l.salt)
	_, _ = fmt.Fprintf(l, "# cairn debug  %s  %s\n", surface, at.Format(time.RFC3339))
	return l, nil
}

func keepingRoom(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var logs []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			logs = append(logs, e.Name())
		}
	}
	slices.Sort(logs)
	for len(logs) >= Kept {
		if err := os.Remove(filepath.Join(dir, logs[0])); err != nil {
			return err
		}
		logs = logs[1:]
	}
	return nil
}

// Write writes down what it is given while the file has room, says the cut once when it has none, and answers that all of it was taken either way: a file full is no reason for cairn to fail.
func (l *Log) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cut || l.left {
		return len(p), nil
	}
	if l.written+len(p) > l.most {
		l.cut = true
		_, err := l.file.WriteString(cut)
		return len(p), err
	}
	n, err := l.file.Write(p)
	l.written += n
	return len(p), err
}

// Say writes one line: the moment, to the millisecond, what said it, and what it said.
func (l *Log) Say(kind, said string) {
	_, _ = fmt.Fprintf(l, "%s %s %s\n", l.now().Format("15:04:05.000"), kind, said)
}

// Fact writes one line of the header: a name, and what it is.
func (l *Log) Fact(name, value string) {
	_, _ = fmt.Fprintf(l, "# %-16s %s\n", name, value)
}

// Text is what the person wrote, as the file shows it: a fingerprint, the same for the same text across this file and no other, and its length — never the text.
func (l *Log) Text(said string) string {
	if said == "" {
		return "‹text 0›"
	}
	sum := sha256.Sum256(append(slices.Clone(l.salt), said...))
	return fmt.Sprintf("‹text %x·%d›", sum[:2], utf8.RuneCountInString(said))
}

// Typed is what was typed, a key or a few at a time, as the file shows it: how many characters, and nothing else. A fingerprint of each would be the same for the same letter all through the file, and the text could be read back from them.
func (l *Log) Typed(said string) string {
	return fmt.Sprintf("‹typed %d›", utf8.RuneCountInString(said))
}

// Args is a command line as the file shows it: the command's own words and every flag's name as they are, a whole number as it is, and anything else the person wrote masked.
func (l *Log) Args(command, args []string) string {
	said := make([]string, 0, len(args))
	for i, arg := range args {
		switch {
		case i < len(command) && arg == command[i]:
			said = append(said, arg)
		case strings.HasPrefix(arg, "-"):
			name, value, valued := strings.Cut(arg, "=")
			if valued {
				name += "=" + l.plainOr(value)
			}
			said = append(said, name)
		default:
			said = append(said, l.plainOr(arg))
		}
	}
	return strings.Join(said, " ")
}

func (l *Log) plainOr(value string) string {
	if _, err := strconv.Atoi(value); err == nil {
		return value
	}
	return l.Text(value)
}

// Panicked writes down what stopped cairn: the value it panicked with, and the stack it was on.
func (l *Log) Panicked(value any, stack []byte) {
	l.Say("panic", fmt.Sprint(value))
	_, _ = l.Write(stack)
}

// Leave closes the file and says where it is, once, however many ways out call it.
func (l *Log) Leave(said io.Writer) {
	l.mu.Lock()
	if l.left {
		l.mu.Unlock()
		return
	}
	l.left = true
	l.mu.Unlock()
	_ = l.file.Close()
	_, _ = fmt.Fprintf(said, "debug: written down in %s\n", l.Path())
}

// Path is where the file is.
func (l *Log) Path() string { return l.file.Name() }

// Close closes the file.
func (l *Log) Close() error { return l.file.Close() }
