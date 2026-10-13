// Command cairn is the entry point.
package main

import (
	"context"
	"os"

	"cairn/internal/cli"
)

func main() {
	os.Exit(cli.Main(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
