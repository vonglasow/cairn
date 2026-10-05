#!/usr/bin/env bash
# packages.sh — the package directories of this module, as ./paths.
#
#   packages.sh                 every package
#   packages.sh --tested  [dir…]   those holding a test file
#   packages.sh --untested [dir…]  those holding none
#
# Given directories, it judges those; given none, every package. Written once
# because two gates need it — `make tested-check` asks which have no test file,
# and scripts/mutation.sh which have one to judge — and each used to list the
# packages and look for a test file its own way.

set -euo pipefail

mode=${1:-}
case $mode in
  --tested|--untested) shift ;;
  *) mode= ;;
esac

if [ $# -gt 0 ]; then
  dirs=("$@")
else
  # shellcheck disable=SC2207
  dirs=($(go list ./... | sed "s|^cairn|.|"))
fi

for dir in "${dirs[@]}"; do
  if ls "$dir"/*_test.go >/dev/null 2>&1; then
    tested=true
  else
    tested=false
  fi
  case $mode in
    --tested) $tested && echo "$dir" ;;
    --untested) $tested || echo "$dir" ;;
    *) echo "$dir" ;;
  esac
done
exit 0
