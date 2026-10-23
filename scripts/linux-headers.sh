#!/usr/bin/env bash
# linux-headers.sh — install what the window is built against on Debian or Ubuntu.
#
#   scripts/linux-headers.sh
#
# The window needs cgo and the system's graphics headers on Linux. The one list
# of them, for the container that runs the suite (scripts/linux.sh) and the
# machine that builds a release (.github/workflows/release.yml) and the one
# that checks a pull request (check.yml); INSTALL.md
# says the same to a person. Run as root, or with sudo on the PATH.
set -euo pipefail

as_root=()
[ "$(id -u)" -eq 0 ] || as_root=(sudo)
"${as_root[@]}" apt-get update -qq >/dev/null
"${as_root[@]}" apt-get install -qq -y --no-install-recommends \
  gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev \
  libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev "$@" >/dev/null
