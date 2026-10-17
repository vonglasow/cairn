#!/usr/bin/env bash
# formula.sh — write the Homebrew formula that installs a version's binaries.
#
#   scripts/formula.sh VERSION DIR URL
#
# DIR holds the binary archives of VERSION and their SHA256SUMS, as the release
# workflow gathers them; URL is where the release serves them. The formula
# fetches the archive made for the system it is installed on and puts the
# binary on the PATH: nothing is built, which is what a binary is handed over
# for — whoever wants to build cairn clones it. It is written twice, as the tap
# keeps them: Formula/cairn.rb, the latest, and Formula/cairn@X.Y.rb, kept for
# every minor. A system whose archive is missing makes no formula at all.
set -euo pipefail

version=${1:?usage: scripts/formula.sh VERSION DIR URL}
dir=${2:?usage: scripts/formula.sh VERSION DIR URL}
url=${3:?usage: scripts/formula.sh VERSION DIR URL}

sum() {
  local archive="cairn-$version-$1.tar.gz" found
  found=$(awk -v a="$archive" '$2 == a { print $1 }' "$dir/SHA256SUMS" 2>/dev/null)
  if [ -z "$found" ]; then
    echo "formula.sh: no sum for $archive in $dir/SHA256SUMS" >&2
    exit 1
  fi
  echo "$found"
}
darwin_arm=$(sum darwin-arm64)
darwin_intel=$(sum darwin-amd64)
linux_arm=$(sum linux-arm64)
linux_intel=$(sum linux-amd64)

formula() {
  cat <<FORMULA
class $1 < Formula
  desc "Personal objective and task manager: a command line, an MCP server and a board"
  homepage "https://github.com/vonglasow/cairn"
  version "${version#v}"

  on_macos do
    on_arm do
      url "$url/cairn-$version-darwin-arm64.tar.gz"
      sha256 "$darwin_arm"
    end
    on_intel do
      url "$url/cairn-$version-darwin-amd64.tar.gz"
      sha256 "$darwin_intel"
    end
  end

  on_linux do
    on_arm do
      url "$url/cairn-$version-linux-arm64.tar.gz"
      sha256 "$linux_arm"
    end
    on_intel do
      url "$url/cairn-$version-linux-amd64.tar.gz"
      sha256 "$linux_intel"
    end
  end

  def install
    bin.install "cairn"
  end

  def caveats
    <<~EOS
      cairn carries its Claude Code plugin. To use it, once:
        cairn plugin write
      then add the folder it names in Claude Code, as it says. cairn keeps
      that plugin at its own version from then on: after an upgrade, the
      next Claude Code session brings it up to date.
    EOS
  end

  test do
    assert_match "${version#v}", shell_output("#{bin}/cairn --version")
  end
end
FORMULA
}

minor=$(echo "${version#v}" | cut -d. -f1-2)
mkdir -p "$dir/Formula"
formula Cairn >"$dir/Formula/cairn.rb"
formula "CairnAT${minor//./}" >"$dir/Formula/cairn@$minor.rb"
