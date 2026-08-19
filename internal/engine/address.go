// Package engine holds the operations. A surface parses input, calls one thing
// here, and prints what comes back — it decides nothing. The engine is also the
// only caller of the store, whatever the store is made of.
package engine

import (
	"fmt"
	"path/filepath"
)

// StoreVariable is the environment variable that says where the store lives.
// Named here rather than spelled out at each call site, because a surface has
// to be able to tell a user what to set.
const StoreVariable = "CAIRN_STORE"

// DefaultFolder and DefaultFile make up the place the store goes when nothing
// says otherwise: the same rule on every system, so the answer to "where are my
// tasks" never depends on the machine.
const (
	DefaultFolder = ".cairn"
	DefaultFile   = "cairn.db"
)

// StoreAddress says where the store lives: what the environment names, or the
// default place under the home directory. It takes its two sources as
// arguments so a test can answer for them without touching a real home.
func StoreAddress(lookupEnv func(string) (string, bool), home func() (string, error)) (string, error) {
	if said, ok := lookupEnv(StoreVariable); ok && said != "" {
		return said, nil
	}
	where, err := home()
	if err != nil {
		return "", fmt.Errorf("finding the default place for the store: %w", err)
	}
	return filepath.Join(where, DefaultFolder, DefaultFile), nil
}
