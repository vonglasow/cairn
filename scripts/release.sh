#!/usr/bin/env bash
# release.sh — make a version: refuse what is not one, check everything, tag
# it, and make the archive it is handed over as.
#
#   make release          (which runs this)
#   make next-version     (which says what it would tag, and does nothing)
#
# The one way a version is made. Its number is not chosen by hand: svu reads
# it off the commits since the last version — a fix moves the patch, a
# feature the minor, a BREAKING CHANGE the major once there is a 1.0 — the
# rules docs/release.md writes down, kept by the commit messages commitlint
# already holds to. A number typed by hand is a number somebody misread the
# history for.
#
# The tag is made only once everything has passed, so a version that failed
# its checks is not a version at all. Run again on a commit already tagged, it
# checks again and makes the archive again, under the tag that is there.
set -euo pipefail

cd "$(dirname "$0")/.."

die() {
  echo "$*" >&2
  exit 1
}

[ "$(git rev-parse --abbrev-ref HEAD)" = main ] || die "a version is made on main"
[ -z "$(git status --porcelain)" ] || die "a version is made from a clean tree"

svu() { go run "github.com/caarlos0/svu/v3@${SVU_VERSION:?SVU_VERSION names the svu to use}" "$@"; }

if tag=$(git describe --tags --exact-match 2>/dev/null); then
  [[ $tag =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "this commit is tagged $tag, which is not a version"
  echo "this commit is $tag already: checking it, and making its archive again"
  fresh=false
else
  tag=$(svu next --v0)
  [ "$tag" != "$(svu current)" ] || die "nothing since $tag calls for a version: no feat, no fix, no BREAKING CHANGE"
  echo "this commit becomes $tag"
  fresh=true
fi

make check
make linux
make build

if $fresh; then
  git tag -a "$tag" -m "$tag"
  echo "tagged $tag"
fi

make dist VERSION="$tag"
