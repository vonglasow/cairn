#!/usr/bin/env bash
# notices.sh — the licences of everything the binary is built from, in one file.
#
#   scripts/notices.sh > bin/THIRD_PARTY_NOTICES
#
# A binary handed to somebody carries code under other people's licences, and
# most of them ask for their text to travel with it. Read from the module
# cache, for the modules the binary actually links — not every module go.mod
# names, which includes what only the tests use.
set -euo pipefail

cache=$(go env GOMODCACHE)

# escaped is a module path as the cache spells it: an upper-case letter as !
# and its lower case. Written out rather than left to sed, whose \L is GNU's.
escaped() {
  local path=$1 out="" c i
  for ((i = 0; i < ${#path}; i++)); do
    c=${path:i:1}
    if [[ $c == [[:upper:]] ]]; then
      out+="!$(printf '%s' "$c" | tr '[:upper:]' '[:lower:]')"
    else
      out+=$c
    fi
  done
  printf '%s' "$out"
}
go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}@{{.Version}}{{end}}{{end}}' ./cmd/cairn |
  sort -u | while read -r module; do
    [ -n "$module" ] || continue
    dir="$cache/$(escaped "$module")"
    licence=$(find "$dir" -maxdepth 1 -iname 'licen[cs]e*' -o -maxdepth 1 -iname 'copying*' | head -1)
    echo "================================================================"
    echo "$module"
    echo "================================================================"
    if [ -n "$licence" ]; then
      cat "$licence"
    else
      echo "(no licence file found in $dir — look before distributing)"
    fi
    echo
  done
