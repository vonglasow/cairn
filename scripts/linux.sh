#!/usr/bin/env bash
# linux.sh — build cairn and run its suite on Linux, in a container.
#
#   scripts/linux.sh
#
# cairn says it runs on macOS and on Linux, and only macOS is where it is
# written: nothing else checked the second half of that sentence. The window
# needs cgo and the system's graphics headers on Linux, which a cross-compile
# from a Mac cannot reach — so the check runs where they are, with the same Go
# go.mod names. Drawing off screen needs a GPU the container does not have, so
# the tests that draw say so and are skipped (see CAIRN_WINDOWLESS). The suite
# runs as somebody other than root, who may write anywhere: a store made
# read-only for a test stays writable to root, and that test then fails for a
# reason no person running cairn would meet.
set -euo pipefail

version=$(awk '/^toolchain go/ {sub("go", "", $2); print $2}' go.mod)
docker run --rm -v "$PWD":/src -w /src \
  -e CAIRN_WINDOWLESS=1 -e GOFLAGS=-buildvcs=false \
  "golang:$version" bash -c '
    apt-get update -qq >/dev/null &&
    apt-get install -qq -y --no-install-recommends \
      libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev \
      libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev >/dev/null &&
    useradd -m tester &&
    runuser -u tester -- env HOME=/home/tester PATH="$PATH" GOPATH=/home/tester/go \
      GOCACHE=/home/tester/.cache/go-build CAIRN_WINDOWLESS=1 GOFLAGS=-buildvcs=false \
      bash -c "go build -o /tmp/cairn ./cmd/cairn && go test ./..."'
