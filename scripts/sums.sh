#!/usr/bin/env bash
# sums.sh — keep the checksum of a file handed over, beside every other one.
#
#   scripts/sums.sh DIR FILE
#
# SHA256SUMS in DIR keeps a line for every file handed over from it: FILE's
# line is written again, the others are left as they were. Rewritten whole,
# it lost the sums of every earlier file the moment a later one was made.
set -euo pipefail

dir=${1:?usage: scripts/sums.sh DIR FILE}
file=${2:?usage: scripts/sums.sh DIR FILE}

cd "$dir"
{
  grep -v "  $file\$" SHA256SUMS 2>/dev/null || true
  shasum -a 256 "$file"
} | sort -k2 >SHA256SUMS.new
mv SHA256SUMS.new SHA256SUMS
