#!/usr/bin/env bash
# rewrote.sh — run a formatter that lists what it rewrote, and fail if it did.
#
#   scripts/rewrote.sh gofmt -l -w file.go…
#
# gofmt and goimports exit 0 whether or not they changed a file, and print the
# files they changed. A hook has to fail when a file was not as committed, so
# what they print is the verdict: nothing printed, nothing rewritten.
set -euo pipefail

rewritten=$("$@")
[ -z "$rewritten" ] || { echo "$rewritten"; exit 1; }
