#!/usr/bin/env bash
# logo.sh — draw cairn's icon at every size a system asks for, from its one source.
#
#   scripts/logo.sh
#
# internal/logo/cairn.svg is the logo; everything else is drawn from it and
# kept beside it so a build needs no drawing tool: the PNGs the window and a
# Linux desktop read, the ones a Mac's application wears, and the Windows resource the .exe carries. Run again after
# the SVG changes. Needs rsvg-convert (librsvg) and Go.
set -euo pipefail

cd "$(dirname "$0")/.."
logo=internal/logo

for size in 16 32 48 64 128 256 512 1024; do
  rsvg-convert -w "$size" -h "$size" "$logo/cairn.svg" -o "$logo/cairn-$size.png"
done

# macOS 26 rounds an application's icon itself, and sets one already rounded
# in a grey frame: the Mac's is the same card drawn to its corners, the mark
# kept at the same share of it.
edge=$(mktemp)
trap 'rm -f "$edge"' EXIT
sed -e 's|<path d="M50 12H78[^"]*"|<rect width="128" height="128"|' \
  -e 's|translate(24 24) scale(0.625)|translate(14.77 14.77) scale(0.76923)|' \
  "$logo/cairn.svg" > "$edge"
for size in 16 32 64 128 256 512 1024; do
  rsvg-convert -w "$size" -h "$size" "$edge" -o "$logo/cairn-edge-$size.png"
done

cd cmd/cairn
go tool -modfile=../../tools/go.mod go-winres simply \
  --icon "../../$logo/cairn-256.png" --arch amd64,arm64 --out rsrc
