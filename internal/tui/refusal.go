package tui

import "fmt"

func unknown(line string) error {
	return fmt.Errorf("%q: not a command — a thing and a verb as on the command line "+
		"(task, objective, backup, migration, trace, config), or q", line)
}
