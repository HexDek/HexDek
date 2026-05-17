# /share/{owner}/{id} Open Graph Meta — Validation Report

**Round:** 17 (dev/share-og-meta, merged as `529ceef` on 2026-05-17)
**Validated:** 2026-05-17
**Branch:** dev/share-og-validation

## Summary

The backend implementation in `internal/hexapi/share_route.go` renders correct
per-deck Open Graph meta for `/share/{owner}/{id}` when hit directly on
DARKSTAR. **Public unfurls via `https://hexdek.dev/share/...` still return the
default site-wide OG block**, because Caddy on MISTY does not forward the
`/share/*` path to the backend. Discord, Twitter, Slack, and Facebook unfurls
will currently show the generic HEXDEK card — not the deck-specific preview —
until the Caddy matcher is updated.

## Test matrix

All probes use the canonical sample URL `/share/belgarathrk/toph_21` plus a
handful of additional decks to exercise the slug/title pipeline.

### 1. Public (`https://hexdek.dev`) — default browser UA

```
$ curl -sI https://hexdek.dev/share/belgarathrk/toph_21
HTTP/2 200
content-type: text/html; charset=utf-8
content-length: 1677
```

Rendered meta:

```
<title>HEXDEK</title>
<meta property="og:title"       content="HEXDEK" />
<meta property="og:description" content="Open-source MTG Commander engine, AI player, and deck-analysis platform." />
<meta property="og:url"         content="https://hexdek.dev" />
<meta property="og:image"       content="https://hexdek.dev/og-default.png" />
<meta property="og:type"        content="website" />
```

**Result:** ❌ Default SPA index.html — Caddy is not proxying `/share/*` to
the backend.

### 2. Public (`https://hexdek.dev`) — crawler User-Agent

Tested with `Discordbot/2.0` and `facebookexternalhit/1.1`. Same default
OG meta as the browser path — bot-gated rewrite never fires.

### 3. Backend direct (`http://192.168.1.207:8090`) — bypasses Caddy

```
$ curl -s http://192.168.1.207:8090/share/belgarathrk/toph_21 \
    | grep -iE 'og:|<title'
<title>Toph 2.1 · Toph, the First Metalbender — HEXDEK</title>
<meta property="og:site_name"   content="HEXDEK" />
<meta property="og:title"       content="Toph 2.1 · Toph, the First Metalbender" />
<meta property="og:description" content="Lands Matter · Bracket B3 · 26% WR · 9505 games" />
<meta property="og:type"        content="article" />
<meta property="og:url"         content="https://hexdek.dev/share/belgarathrk/toph_21" />
<meta property="og:image"       content="https://hexdek.dev/api/card-art/Toph%2C%20the%20First%20Metalbender" />
<meta name="twitter:card"        content="summary_large_image" />
<meta name="twitter:title"       content="Toph 2.1 · Toph, the First Metalbender" />
<meta name="twitter:description" content="Lands Matter · Bracket B3 · 26% WR · 9505 games" />
<meta name="twitter:image"       content="https://hexdek.dev/api/card-art/Toph%2C%20the%20First%20Metalbender" />
```

**Result:** ✅ All OG and Twitter fields populated correctly. `og:type=article`,
`twitter:card=summary_large_image`, absolute URLs, custom deck title from
`deck_meta.custom_name`, summary built from Freya archetype + bracket + win
record.

### 4. Spot-checks — additional decks

| URL | Status | og:title | og:description |
|-----|--------|----------|----------------|
| `/share/7174n1c/varina_tribal_widetall_b3_lich_queen` | 200 | `VARINA TRIBAL WIDETALL B3 LICH QUEEN · Varina, Lich Queen` | `Tribal · Bracket B2 · 26% WR · 210682 games` |
| `/share/hex/hex_ninjas_b3_yuriko` | 200 | `HEX NINJAS B3 YURIKO · Yuriko, the Tiger&#39;s Shadow` | `Tribal · Bracket B3 · 25% WR · 156180 games` |
| `/share/belgarathrk/nonexistent_deck` | 404 | — | — |
| `/share/INVALID..PATH/foo` | 404 | — | — |

Notes:
- Apostrophes are HTML-encoded (`&#39;`) in attribute content and URL-encoded
  (`%27`) in the `og:image` path — safe to parse for every unfurler tested.
- `/api/card-art/Toph%2C%20the%20First%20Metalbender` returns
  `HTTP 200, image/jpeg, ~52KB` — valid image source for the unfurl.
- Path-component validation (`validatePathComponent`) rejects `..` traversal
  before any disk lookup.

### 5. Discord / Slack / Twitter unfurl

Cannot be exercised end-to-end from this session, but the outcome is fully
determined by step 1: any crawler that fetches `https://hexdek.dev/share/...`
today receives the default index.html, so the unfurl card will read **"HEXDEK
— Open-source MTG Commander engine, AI player, and deck-analysis platform."**
with the generic `og-default.png` artwork — *not* the deck-specific preview.

Once the Caddy fix below is deployed, an opengraph.dev or
https://www.opengraph.xyz/ probe of `https://hexdek.dev/share/belgarathrk/toph_21`
should mirror the backend-direct output from step 3.

## Root cause

`internal/hexapi/share_route.go:23` already warned about this:

> Caddy must route `/share/{owner}/{id}` to this backend (mirror of the existing
> `/decks/{owner}/{id}` Caddy rule).

The MISTY `Caddyfile` (lines 297–303) defines a User-Agent-gated proxy that
covers `/decks/*` but not `/share/*`:

```caddyfile
@bot_share {
    header_regexp User-Agent "(Discordbot|Twitterbot|facebookexternalhit|Slackbot|LinkedInBot|WhatsApp|TelegramBot|Googlebot)"
    path /decks/* /cards/* /operator/* /spectate /leaderboard
}
handle @bot_share {
    reverse_proxy 192.168.1.207:8090
}
```

`/decks/belgarathrk/toph_21` with `facebookexternalhit/1.1` returns the
backend-rendered OG block (verified during this run). `/share/...` with the
same UA does not, because the path is not in the matcher.

## Recommended fix (Caddy on MISTY)

Add `/share/*` to the `@bot_share` matcher:

```caddyfile
@bot_share {
    header_regexp User-Agent "(Discordbot|Twitterbot|facebookexternalhit|Slackbot|LinkedInBot|WhatsApp|TelegramBot|Googlebot)"
    path /decks/* /share/* /cards/* /operator/* /spectate /leaderboard
}
```

Then reload Caddy on MISTY (`caddy reload`, not `systemctl reload caddy` per
prior watchdog quirk).

Open question for Josh: do we also want `/share/*` to render the per-deck
preview for *non-bot* visitors (so a human pasting the link into their own
browser sees the rich title), or keep it bot-gated to preserve SPA hydration?
The current `/decks/*` strategy is bot-gated; the simplest fix mirrors that.

## Secondary observations (non-blocking)

1. **Slug→title leaves bracket tokens mid-string.** `slugToDeckName` only
   trims `_b0..b5` from the end. A deck id like
   `varina_tribal_widetall_b3_lich_queen` renders as
   `VARINA TRIBAL WIDETALL B3 LICH QUEEN` — readable but ugly. A custom_name
   override in `deck_meta` masks this; only decks without one are affected.
   Low priority — fold into the next forge polish pass.

2. **`og:title` upper-cases the whole slug.** Title case (`Varina Tribal
   Widetall …`) would unfurl more naturally. Same fix surface as #1.

3. **404 vs 400 on malformed path components.** `INVALID..PATH` returns 404
   instead of 400 (`validatePathComponent` rejects, but the outer handler
   maps the rejection to NotFound). Not user-visible; noted for parity.

## Conclusion

Backend implementation is correct and ready. **Caddy routing is the only
blocker between the merged code and working Discord/Slack/Twitter unfurls.**
Apply the one-line `Caddyfile` change above and re-run the public-URL probe
in section 1 to confirm.
