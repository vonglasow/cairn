// Command cairn is the entry point. It hands the process's arguments and
// streams to cli.Main and exits with what it returns — nothing else, so that
// everything worth testing is reachable in-process without spawning a binary.
package main

import (
	"context"
	"os"

	"cairn/internal/cli"
)

func main() {
	os.Exit(cli.Main(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
