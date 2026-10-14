#!/usr/bin/env bash
# logo.sh — draw cairn's icon at every size a system asks for, from its one source.
#
#   scripts/logo.sh
#
# internal/gui/logo/cairn.svg is the logo; everything else is drawn from it and
# kept beside it so a build needs no drawing tool: the PNGs the window and a
# Linux desktop read, and the Windows resource the .exe carries. Run again after
# the SVG changes. Needs rsvg-convert (librsvg) and Go.
set -euo pipefail

cd "$(dirname "$0")/.."
logo=internal/gui/logo

for size in 16 32 48 64 128 256 512 1024; do
  rsvg-convert -w "$size" -h "$size" "$logo/cairn.svg" -o "$logo/cairn-$size.png"
done

cd cmd/cairn
go run github.com/tc-hib/go-winres@v0.3.3 simply \
  --icon "../../$logo/cairn-256.png" --arch amd64,arm64 --out rsrc
