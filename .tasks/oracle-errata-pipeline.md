# Task: Oracle Errata Pipeline

## Context
HexDek uses a local `data/oracle-cards.json` (Scryfall bulk data) as the source of truth for card oracle text. When Wizards issues errata or new sets release, oracle text changes. Currently we have no automated way to detect these changes and re-validate affected cards through our parser + Thor.

## Task
Build `cmd/hexdek-oracle-sync/main.go` — a standalone CLI tool:

1. **Scryfall bulk pull** — download the Scryfall "Oracle Cards" bulk data JSON (~80MB):
   - Endpoint: first GET `https://api.scryfall.com/bulk-data` to find the `oracle_cards` entry
   - Then download the `download_uri` from that entry
   - Save to `data/oracle-cards-new.json`
   - Respect Scryfall rate limits (100ms between requests)

2. **Diff engine** — compare `data/oracle-cards.json` (current) vs `data/oracle-cards-new.json` (fresh):
   - Key comparison: match cards by `oracle_id`
   - Detect: oracle_text changes, type_line changes, mana_cost changes, name changes
   - Output: list of changed cards with before/after oracle text
   - Write diff report to `data/oracle-sync-report.md`

3. **Re-parse changed cards** — for each changed card:
   - Run the oracle text through the AST parser (`internal/parser/`)
   - Compare old AST vs new AST
   - Flag cards where parse result changed (new abilities detected, abilities removed, etc)

4. **Thor validation** — run Thor on ONLY the changed cards:
   - Use existing Thor infrastructure to test just the affected subset
   - Append results to the sync report

5. **Promote** — if `--promote` flag is passed, replace `data/oracle-cards.json` with the new version

6. **CLI interface**:
   ```
   hexdek-oracle-sync [flags]
     --dry-run     Download and diff only, no re-parse or Thor
     --promote     Replace oracle-cards.json with new version after validation
     --verbose     Print changed cards to stdout
     --output DIR  Override output directory (default: data/)
   ```

7. **Report format** (`data/oracle-sync-report.md`):
   ```markdown
   # Oracle Sync Report
   **Source:** Scryfall bulk data (oracle_cards)
   **Cards checked:** {total}
   **Changes detected:** {count}
   **Parse changes:** {count where AST differs}
   **Thor failures on changed set:** {count}

   ## Changed Cards
   | Card | Field | Before | After |

   ## Parse Impact

   ## Thor Results on Changed Set
   ```

Look at `data/oracle-cards.json` to understand the current format. Look at `internal/parser/` to understand how to invoke the parser. Look at `internal/tools/thor/` to understand how to run Thor on a subset. This is a standalone CLI tool — keep it self-contained in its own cmd/ directory.
