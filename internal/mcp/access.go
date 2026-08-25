package mcp

import (
	"errors"
	"fmt"
)

// ErrNotAnAccess is the reason an access level is refused: what was given
// spelled none of "read", "write" or "full".
var ErrNotAnAccess = errors.New("not one of the access levels")

// Access is how far an agent may reach, settled once when the server starts.
// The three levels are ordered — read, then write, then full — so a level
// that was never read must not silently be one of them: the zero value names
// none of the three.
type Access int

const (
	_ Access = iota // the zero value: never read, so never one of the three.
	// Read looks and never writes.
	Read
	// Write also adds and changes.
	Write
	// Full also does what loses work.
	Full
)

// ParseAccess reads an access level by name. Anything other than "read",
// "write" or "full" is refused, wrapping ErrNotAnAccess.
func ParseAccess(name string) (Access, error) {
	switch name {
	case "read":
		return Read, nil
	case "write":
		return Write, nil
	case "full":
		return Full, nil
	}
	return 0, fmt.Errorf("%q: expected \"read\", \"write\" or \"full\": %w", name, ErrNotAnAccess)
}
