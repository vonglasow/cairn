#!/usr/bin/env bash
# dist.sh — make the source archive a version is handed over as.
#
#   scripts/dist.sh v1.2.0 dist [bin/THIRD_PARTY_NOTICES]
#
# A version goes to colleagues as its sources, built where they are with
# `make build` — on macOS and on Linux alike, with nothing to host anywhere.
# An archive has no git history, so it carries what git would have said: the
# version in VERSION and the commit in COMMIT, which the Makefile reads when
# there is no repository to ask. Without them a build from the archive called
# itself "dev", made from "none".
#
# The tree archived is the one as it stands: on a release that is the tagged
# commit, since make release refuses anything else, and while this is worked
# on it is what is being worked on rather than what was last committed.
set -euo pipefail

version=${1:?usage: scripts/dist.sh VERSION OUT [NOTICES]}
out=${2:?usage: scripts/dist.sh VERSION OUT [NOTICES]}
notices=${3:-}

cd "$(dirname "$0")/.."

# Asked of a clean tree, git stash create can fail outright: in a container
# the repository is mounted into, the index still carries the stat of the
# machine it came from, and git 2.39 gives up on it rather than refresh it.
# git status refreshes it, and says whether there is anything to stash.
tree=HEAD
if [ -n "$(git status --porcelain)" ]; then
  tree=$(git stash create)
fi
commit=$(git rev-parse HEAD)
name="cairn-$version"

extra=()
if [ -n "$notices" ]; then
  extra+=(--add-file="$notices")
fi

mkdir -p "$out"
git archive --format=tar.gz --prefix="$name/" \
  --add-virtual-file="$name/VERSION:$version" \
  --add-virtual-file="$name/COMMIT:$commit" \
  ${extra[@]+"${extra[@]}"} \
  -o "$out/$name-src.tar.gz" "$tree"

cd "$out"
shasum -a 256 "$name-src.tar.gz" > SHA256SUMS
cat SHA256SUMS
