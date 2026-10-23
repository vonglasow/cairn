#!/usr/bin/env bash
# release.sh — make a version: refuse what is not one, check everything, and
# tag it. Pushing the tag hands it over.
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
# its checks is not a version at all. Everything is every check but the
# mutation sweep, which is run on its own — see make mutation. Run again on a
# commit already tagged, it checks it again, under the tag that is there.
set -euo pipefail

cd "$(dirname "$0")/.."

die() {
  echo "$*" >&2
  exit 1
}

[ "$(git rev-parse --abbrev-ref HEAD)" = main ] || die "a version is made on main"
[ -z "$(git status --porcelain)" ] || die "a version is made from a clean tree"

svu() { go tool -modfile=tools/go.mod svu "$@"; }

if tag=$(git describe --tags --exact-match 2>/dev/null); then
  [[ $tag =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "this commit is tagged $tag, which is not a version"
  echo "this commit is $tag already: checking it again"
  fresh=false
else
  tag=$(svu next --v0)
  [ "$tag" != "$(svu current)" ] || die "nothing since $tag calls for a version: no feat, no fix, no BREAKING CHANGE"
  echo "this commit becomes $tag"
  fresh=true
fi

# Every check but the mutation sweep, which is judged on its own with make
# mutation: a release that waited on it took the better part of an hour. What
# the last sweep found is said, so a version is not made in ignorance of it.
if [ -f mutation-report/summary.md ]; then
  echo "the last mutation sweep: $(sed -n 3p mutation-report/summary.md) — $(tail -n 1 mutation-report/summary.md)"
else
  echo "no mutation sweep has been run here: make mutation judges every package"
fi
make release-check
make linux
make build

# The tag carries the version's own notes — what changed since the last one —
# so `git show v1.2.0` answers what a person upgrading wants to know.
# Verbatim: git strips every line starting with # from a tag's message, and
# took every heading of the notes with it.
if $fresh; then
  scripts/changelog.sh "$tag" | awk '/^## / { n++ } n == 1' | git tag -a --cleanup=verbatim "$tag" -F -
  echo "tagged $tag"
fi

# Pushing the tag is what hands the version over: GitHub builds it on every
# system and publishes it (.github/workflows/release.yml). Pushing is left to
# whoever made it, who decides it goes out.
echo "push it to release it: git push origin main && git push origin $tag"
