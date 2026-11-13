// Command site builds cairn's documentation site.
package main

import (
	"os"

	"cairn/internal/site"
)

func main() {
	os.Exit(site.Main(os.Args[1:], os.Stdout, os.Stderr))
}
