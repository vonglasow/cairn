#!/usr/bin/env bash
# dist.sh — make the source archive a version is handed over as.
#
#   scripts/dist.sh v1.2.0 dist [bin/THIRD_PARTY_NOTICES]
#
# A version goes to colleagues as its sources, built where they are with
# `make build` — on macOS and on Linux alike, with nothing to host anywhere.
# An archive has no git history, so it carries what git would have said: the
# version in VERSION, the commit in COMMIT and its date in COMMIT_DATE, which
# the Makefile reads when there is no repository to ask. Without them a build
# from the archive called itself "dev", made from "none", on a date "unknown".
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
committed=$(TZ=UTC git show -s --format=%cd --date=format-local:%Y-%m-%dT%H:%M:%SZ HEAD)
name="cairn-$version"

# What changed, version by version, read off the tags rather than kept in a
# file: the archive carries it, and so does the folder it is made in.
mkdir -p "$out"
scripts/changelog.sh "$version" >"$out/CHANGELOG.md"

extra=(--add-file="$out/CHANGELOG.md")
if [ -n "$notices" ]; then
  extra+=(--add-file="$notices")
fi

git archive --format=tar.gz --prefix="$name/" \
  --add-virtual-file="$name/VERSION:$version" \
  --add-virtual-file="$name/COMMIT:$commit" \
  --add-virtual-file="$name/COMMIT_DATE:$committed" \
  "${extra[@]}" \
  -o "$out/$name-src.tar.gz" "$tree"

# Every archive made here keeps its sum: the line of the one just made is
# written again, the others are left as they were. Rewritten whole, the file
# lost the sums of every earlier version the moment a later one was made.
cd "$out"
{
  grep -v "  $name-src.tar.gz\$" SHA256SUMS 2>/dev/null || true
  shasum -a 256 "$name-src.tar.gz"
} | sort -k2 >SHA256SUMS.new
mv SHA256SUMS.new SHA256SUMS
cat SHA256SUMS

# The Homebrew formula that builds this archive. Homebrew installs only from a
# tap, and a tap holds formulae: this one, made with the archive, can never
# name another archive or another sum. It fetches the archive from where it
# was made, a file on this machine, unless FORMULA_URL says where archives are
# hosted. See docs/release.md.
sum=$(shasum -a 256 "$name-src.tar.gz" | cut -d' ' -f1)
url="${FORMULA_URL:-file://$(pwd)}/$name-src.tar.gz"
mkdir -p Formula
cat >Formula/cairn.rb <<FORMULA
class Cairn < Formula
  desc "Personal objective and task manager: a command line, an MCP server and a board"
  url "$url"
  version "${version#v}"
  sha256 "$sum"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "bin/cairn"
  end

  test do
    assert_match "${version#v}", shell_output("#{bin}/cairn --version")
  end
end
FORMULA
