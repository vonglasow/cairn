//go:build !windows

package upgrade

import "os"

func put(from, to string) error { return os.Rename(from, to) }
