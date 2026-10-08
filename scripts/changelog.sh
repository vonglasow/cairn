#!/usr/bin/env bash
# changelog.sh — say what changed, version by version, newest first.
#
#   scripts/changelog.sh [VERSION]
#
# Read off the history, never kept in a file: every version is a tag, and what
# it brought is the conventional commits between it and the one before —
# commitlint already holds every message to that shape. VERSION names the
# commits after the last tag when they are not one yet, which is how an archive
# made before its tag says what it is.
#
# What a person using cairn reads: what breaks, what is new, what was fixed.
# The rest — docs, tests, chores — is in the history for whoever wants it.
set -euo pipefail

version=${1:-}
cd "$(dirname "$0")/.."

# notes writes one version's part: its heading, then each kind that has
# something to say. $1 is the version, $2 its day, $3 the range of commits.
notes() {
  local breaking="" features="" fixes="" hash subject body type scope said line
  while read -r hash; do
    subject=$(git log -1 --format=%s "$hash")
    body=$(git log -1 --format=%b "$hash")
    [[ $subject =~ ^([a-z]+)(\(([^\)]*)\))?(!)?:\ (.*)$ ]] || continue
    type=${BASH_REMATCH[1]} scope=${BASH_REMATCH[3]} said=${BASH_REMATCH[5]}
    line="- ${scope:+**$scope:** }$said ($hash)"
    if [ -n "${BASH_REMATCH[4]}" ] || [[ $body == *"BREAKING CHANGE"* ]]; then
      breaking+="$line"$'\n'
    fi
    case $type in
      feat) features+="$line"$'\n' ;;
      fix) fixes+="$line"$'\n' ;;
    esac
  done < <(git rev-list --abbrev-commit "$3")
  echo "## $1 — $2"
  echo
  if [ -n "$breaking" ]; then printf '### Breaking changes\n\n%s\n' "$breaking"; fi
  if [ -n "$features" ]; then printf '### Features\n\n%s\n' "$features"; fi
  if [ -n "$fixes" ]; then printf '### Fixes\n\n%s\n' "$fixes"; fi
}

echo "# Changelog"
echo
tags=()
while read -r tag; do
  tags+=("$tag")
done < <(git tag --list 'v*' --merged HEAD --sort=-v:refname)

if [ -n "$version" ] && ! printf '%s\n' ${tags[@]+"${tags[@]}"} | grep -qx -- "$version"; then
  notes "$version" "$(date -u +%F)" "${tags[0]:+${tags[0]}..}HEAD"
fi
for i in ${tags[@]+"${!tags[@]}"}; do
  tag=${tags[$i]}
  before=${tags[$((i + 1))]:-}
  notes "$tag" "$(git log -1 --format=%cs "$tag")" "${before:+$before..}$tag"
done
