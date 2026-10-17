#!/usr/bin/env bash
# package.sh — build cairn here and hand it over as an archive named for this system.
#
#   scripts/package.sh VERSION OUT [NOTICES]
#
# The window needs cgo, so a binary is built on the system it runs on and on no
# other: the release workflow runs this once on each (.github/workflows/
# release.yml). The archive holds the binary, stamped with VERSION, the two
# files that say what it is and how to install it, and the licences of what it
# is built from when NOTICES names them; its sum joins SHA256SUMS.
set -euo pipefail

version=${1:?usage: scripts/package.sh VERSION OUT}
out=${2:?usage: scripts/package.sh VERSION OUT [NOTICES]}
notices=${3:-}

cd "$(dirname "$0")/.."
os=$(go env GOOS)
arch=$(go env GOARCH)
name="cairn-$version-$os-$arch"
binary=cairn
[ "$os" = windows ] && binary=cairn.exe

# The binary is put together in a folder of the tree rather than the system's
# temporary one: on Windows, Go is a Windows program, and the /tmp Git Bash
# shows it is not a path it can write to.
mkdir -p "$out"
work=$(mktemp -d ./.package.XXXXXX)
trap 'rm -rf "$work"' EXIT
make -s build VERSION="$version" BIN="$work/$binary"
cp README.md INSTALL.md "$work/"
[ -z "$notices" ] || cp "$notices" "$work/THIRD_PARTY_NOTICES"
out=$(cd "$out" && pwd)

if [ "$os" = windows ]; then
  archive="$name.zip"
  (cd "$work" && 7z a -tzip -bso0 "$out/$archive" .)
else
  archive="$name.tar.gz"
  tar -czf "$out/$archive" -C "$work" .
fi

scripts/sums.sh "$out" "$archive"
echo "$out/$archive"
