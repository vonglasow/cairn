#!/usr/bin/env bash
# desktop.sh — give cairn's window its logo on a Linux desktop.
#
#   scripts/desktop.sh [BINARY]
#
# A Linux desktop finds a window's icon through a desktop entry named after the
# window: cairn's window is known as "cairn", on Wayland and on X11 alike, so
# the entry is cairn.desktop and the logo is the "cairn" icon of the hicolor
# theme. Both go under the user's own data folder, nothing system-wide. BINARY
# is the cairn the entry starts; the one on the PATH when it is left out.
set -euo pipefail

cd "$(dirname "$0")/.."
binary=${1:-$(command -v cairn || true)}
if [ -z "$binary" ]; then
  echo "desktop.sh: no cairn on the PATH — install it first, or name it" >&2
  exit 1
fi
binary=$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary")
data=${XDG_DATA_HOME:-$HOME/.local/share}

for size in 16 32 48 64 128 256 512; do
  mkdir -p "$data/icons/hicolor/${size}x${size}/apps"
  install -m 644 "internal/gui/logo/cairn-$size.png" "$data/icons/hicolor/${size}x${size}/apps/cairn.png"
done

mkdir -p "$data/applications"
cat >"$data/applications/cairn.desktop" <<ENTRY
[Desktop Entry]
Type=Application
Name=cairn
Comment=Objectives and tasks, and the time the work took
Exec="$binary" gui
Icon=cairn
StartupWMClass=cairn
Terminal=false
Categories=Office;ProjectManagement;
ENTRY

command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t "$data/icons/hicolor" || true
command -v update-desktop-database >/dev/null && update-desktop-database -q "$data/applications" || true
echo "cairn.desktop and its logo are in $data"
