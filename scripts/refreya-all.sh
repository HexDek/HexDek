#!/usr/bin/env bash
# refreya-all.sh — re-run Freya across every decklist under data/decks/
# so each deck gets a fresh strategy.json (including card_roles, which
# older runs may have left null).
#
# Usage:
#   scripts/refreya-all.sh                  # all top-level deck folders
#   scripts/refreya-all.sh personal meglin  # just the named folders
#
# Oracle corpus loads on every freya invocation (~163MB JSON), so we
# call --all-decks once per top-level folder where possible to amortize
# that cost. The benched/ and test/ folders are special-cased because
# Freya's directory walker SkipDirs anything literally named "benched"
# or "test"; for those we loop over individual files.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

DECKS_DIR="data/decks"
BIN="$(mktemp -t hexdek-freya.XXXXXX)"
trap 'rm -f "$BIN"' EXIT

if [[ ! -f data/rules/oracle-cards.json ]]; then
  echo "missing data/rules/oracle-cards.json — run scripts/fetch-oracle.sh first" >&2
  exit 1
fi

echo "==> building hexdek-freya"
go build -o "$BIN" ./cmd/hexdek-freya/

# Folders Freya's WalkDir refuses to descend into when passed as root.
# These need the per-file fallback.
SKIP_NAMES=(freya benched test)
is_skipped() {
  local name=$1
  for s in "${SKIP_NAMES[@]}"; do [[ "$name" == "$s" ]] && return 0; done
  return 1
}

# Determine which top-level folders to process.
if [[ $# -gt 0 ]]; then
  TARGETS=("$@")
else
  TARGETS=()
  while IFS= read -r -d '' d; do
    TARGETS+=("$(basename "$d")")
  done < <(find "$DECKS_DIR" -mindepth 1 -maxdepth 1 -type d -print0)
fi

processed=0
failed=0

for folder in "${TARGETS[@]}"; do
  path="$DECKS_DIR/$folder"
  if [[ ! -d "$path" ]]; then
    echo "  SKIP $folder (not a directory)"
    continue
  fi

  if [[ "$folder" == "freya" ]]; then
    continue
  fi

  if is_skipped "$folder"; then
    # Per-file fallback: Freya's listDeckFiles skips these names entirely
    # when they appear as the walk root, so --all-decks is a no-op here.
    echo "==> $folder (per-file mode)"
    while IFS= read -r -d '' f; do
      if "$BIN" --deck "$f" --format text >/dev/null 2>&1; then
        processed=$((processed + 1))
        echo "    ok  $(basename "$f")"
      else
        failed=$((failed + 1))
        echo "    FAIL $(basename "$f")" >&2
      fi
    done < <(find "$path" -maxdepth 1 -name "*.txt" -print0)
  else
    echo "==> $folder (--all-decks mode)"
    if "$BIN" --all-decks "$path" --format text >/dev/null; then
      n=$(find "$path" -name "*.txt" -not -path "*/freya/*" | wc -l | tr -d ' ')
      processed=$((processed + n))
      echo "    ok  ($n decks)"
    else
      echo "    FAIL on $folder (--all-decks)" >&2
      failed=$((failed + 1))
    fi
  fi
done

echo
echo "done: $processed deck(s) processed, $failed failure(s)"

# Quick verification: count how many of the produced strategy.json
# files now have a populated card_roles map.
if command -v jq >/dev/null 2>&1; then
  total=0
  populated=0
  while IFS= read -r -d '' f; do
    total=$((total + 1))
    n=$(jq '.card_roles | if . == null then 0 else length end' "$f" 2>/dev/null || echo 0)
    if [[ "${n:-0}" -gt 0 ]]; then populated=$((populated + 1)); fi
  done < <(find "$DECKS_DIR" -name "*.strategy.json" -print0)
  echo "card_roles populated: $populated / $total strategy.json"
fi
