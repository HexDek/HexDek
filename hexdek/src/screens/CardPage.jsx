import { useEffect, useMemo, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { Panel, KV, Tag, Tape, Btn } from '../components/chrome'
import { API_BASE, cardArtUrl } from '../services/api'

// CardPage — /cards/:cardName
//
// Two-tier data flow:
//
//   1. Try /api/cards/{name} first. The local oracle corpus knows the
//      card's mana cost, type line, oracle text, CMC, set, and — via a
//      DecksDir scan — every indexed deck that contains it (with its
//      Freya role tag when strategy.json exists). One round trip, no
//      per-deck fan-out, no Scryfall dependency for the page to render.
//
//   2. If the local API returns 404 (card not in the loaded oracle
//      bulk), fall back to Scryfall /cards/named?exact=... for type/
//      oracle/cost. The "USED IN" panel then comes back empty since
//      the card isn't indexed locally.
//
// Card art always uses the existing /api/card-art/{name} proxy
// (cardArtUrl()), which redirects to Scryfall's image_uris.art_crop
// server-side. That keeps a single Scryfall dependency and avoids
// CORS surprises.

const ROLE_KIND = {
  // Map Freya role identifiers to Tag kinds for color.
  removal: 'bad',
  counterspell: 'bad',
  ramp: 'ok',
  draw: 'info',
  tutor: 'warn',
  finisher: 'warn',
  combo_piece: 'warn',
  protection: 'ok',
  recursion: 'info',
  threat: 'bad',
  utility: null,
  value: 'ok',
}

function readableRole(r) {
  return (r || '').replace(/_/g, ' ').toUpperCase()
}

function priceFromScryfall(s) {
  const usd = s?.prices?.usd
  if (usd != null && usd !== '') return `$${usd}`
  const usdFoil = s?.prices?.usd_foil
  if (usdFoil != null && usdFoil !== '') return `$${usdFoil} (FOIL)`
  return '—'
}

export default function CardPage() {
  const { cardName: rawName } = useParams()
  const navigate = useNavigate()
  const cardName = useMemo(() => decodeURIComponent(rawName || ''), [rawName])

  // Local API response (CardDetail). null = no data yet or 404.
  const [detail, setDetail] = useState(null)
  // Scryfall response — only fetched when the local API 404s.
  const [scry, setScry] = useState(null)
  // Loading / error state covers whichever path is in flight.
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [source, setSource] = useState(null) // 'local' | 'scryfall' | null

  useEffect(() => {
    if (!cardName) return
    let cancelled = false
    setLoading(true)
    setError(false)
    setDetail(null)
    setScry(null)
    setSource(null)

    const tryLocal = fetch(`${API_BASE}/api/cards/${encodeURIComponent(cardName)}`)
      .then(async r => {
        if (r.status === 404) return { kind: '404' }
        if (!r.ok) throw new Error(`local ${r.status}`)
        return { kind: 'ok', data: await r.json() }
      })

    const tryScryfall = () =>
      fetch(`https://api.scryfall.com/cards/named?exact=${encodeURIComponent(cardName)}`)
        .then(r => r.ok ? r.json() : Promise.reject(new Error(`scryfall ${r.status}`)))

    tryLocal
      .then(res => {
        if (cancelled) return
        if (res.kind === 'ok') {
          setDetail(res.data)
          setSource('local')
          setLoading(false)
          return
        }
        // 404 from local API → fall back to Scryfall entirely.
        return tryScryfall().then(data => {
          if (cancelled) return
          setScry(data)
          setSource('scryfall')
          setLoading(false)
        })
      })
      .catch(() => {
        if (!cancelled) {
          setError(true)
          setLoading(false)
        }
      })

    return () => { cancelled = true }
  }, [cardName])

  if (!cardName) {
    return (
      <>
        <Tape left="CARD" mid="MISSING NAME" right="DOC HX-700" />
        <div style={{ padding: 36, textAlign: 'center' }}>
          <div className="t-md muted">&gt; NO CARD NAME IN URL.</div>
        </div>
      </>
    )
  }

  // Unified view model. The local CardDetail is preferred; Scryfall's
  // payload is the fallback shape (mana_cost / type_line / oracle_text /
  // set / cmc all live at the top level too, with card_faces for DFCs).
  const upperName = cardName.toUpperCase()

  const localDecks = detail?.decks_using || []
  // Freya roles: union of every non-empty role across the decks_using
  // list. The backend resolves card_roles from each deck's strategy.json,
  // so this needs no per-deck fan-out from the frontend.
  const roles = useMemo(() => {
    const set = new Set()
    for (const d of localDecks) {
      if (d?.role) set.add(d.role)
    }
    return [...set]
  }, [localDecks])

  const manaCost = detail?.mana_cost
    || scry?.mana_cost
    || scry?.card_faces?.[0]?.mana_cost
    || ''
  const typeLine = detail?.type_line
    || scry?.type_line
    || scry?.card_faces?.[0]?.type_line
    || '—'
  const oracle = detail?.oracle_text
    || scry?.oracle_text
    || scry?.card_faces?.[0]?.oracle_text
    || ''
  const setCode = (detail?.set || scry?.set || '').toUpperCase()
  const setName = scry?.set_name || setCode || '—'
  const cmc = detail?.cmc != null ? detail.cmc : (scry?.cmc != null ? scry.cmc : null)
  const price = priceFromScryfall(scry) // local API doesn't carry price; only present on Scryfall fallback.

  const heroArt = cardArtUrl(cardName)
  const sourceLabel = source === 'local'
    ? 'LOCAL CORPUS'
    : source === 'scryfall'
      ? 'SCRYFALL'
      : error ? 'UNAVAILABLE' : 'LOADING'

  return (
    <>
      <Tape
        left={`CARD / / ${upperName}`}
        mid={loading ? 'LOADING' : (error ? 'UNAVAILABLE' : sourceLabel)}
        right="DOC HX-700"
      />

      {/* Full-bleed art hero */}
      <div
        className="card-page-hero"
        style={{
          backgroundImage: heroArt ? `url(${heroArt})` : undefined,
        }}
      >
        <div className="card-page-hero-scrim" />
        <div className="card-page-hero-corner card-page-hero-corner--tl">04.HERO / / {setCode || 'UNKNOWN'}</div>
        <div className="card-page-hero-corner card-page-hero-corner--tr">{upperName}</div>
        <div className="card-page-hero-body">
          <div className="card-page-hero-meta">
            {typeLine !== '—' && <Tag solid>{typeLine.toUpperCase()}</Tag>}
            {manaCost && <Tag>{manaCost}</Tag>}
            {price !== '—' && <Tag kind="ok">{price}</Tag>}
          </div>
          <h1 className="card-page-hero-title">{upperName}</h1>
          {setName !== '—' && (
            <div className="card-page-hero-sub">{setName.toUpperCase()}{setCode ? ` · ${setCode}` : ''}</div>
          )}
        </div>
      </div>

      <div className="card-page-layout">
        <div className="card-page-sidebar">
          <Panel code="07.A" title="CARD SPECS" solid>
            <KV rows={[
              ['NAME', upperName],
              ['TYPE', typeLine],
              ['MANA', manaCost || '—'],
              ['CMC', cmc != null ? `${cmc}` : '—'],
              ['SET', setCode || '—'],
              ...(source === 'scryfall' ? [['PRICE (USD)', price]] : []),
              ...(scry?.legalities?.commander ? [[
                'LEGAL COMMANDER',
                <span style={{ color: scry.legalities.commander === 'legal' ? 'var(--ok)' : 'var(--danger)', fontWeight: 700 }}>
                  {scry.legalities.commander.toUpperCase()}
                </span>,
              ]] : []),
              ['SOURCE', sourceLabel],
            ]} />
            {scry?.scryfall_uri && (
              <>
                <div className="hr" style={{ margin: '10px 0' }} />
                <a href={scry.scryfall_uri} target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
                  <Btn ghost arrow="↗">VIEW ON SCRYFALL</Btn>
                </a>
              </>
            )}
          </Panel>

          {roles.length > 0 && (
            <Panel code="07.R" title={`FREYA ROLES / / ${roles.length}`}>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                {roles.map(r => (
                  <Tag key={r} kind={ROLE_KIND[r]} solid>{readableRole(r)}</Tag>
                ))}
              </div>
              <div className="t-xs muted" style={{ marginTop: 6 }}>
                ROLES SOURCED FROM EACH USING DECK'S STRATEGY.JSON.
              </div>
            </Panel>
          )}
        </div>

        <div className="card-page-main">
          {/* Oracle text */}
          <Panel code="07.B" title="ORACLE TEXT" right={
            <span className="t-xs muted">{loading ? 'LOADING' : (error ? 'OFFLINE' : sourceLabel)}</span>
          }>
            {loading ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; FETCHING CARD RECORD<span className="blink">_</span>
              </div>
            ) : error ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; CARD DATA UNAVAILABLE — TRY AGAIN LATER.
              </div>
            ) : (
              <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.55, fontSize: 12 }}>
                {oracle || <span className="muted">&gt; NO ORACLE TEXT.</span>}
              </div>
            )}
          </Panel>

          {/* Used in — decks containing this card. Sourced from
              detail.decks_using when the local API responded; the
              Scryfall fallback path leaves this panel empty since the
              card isn't in our oracle corpus. */}
          <Panel
            code="07.C"
            title={`USED IN / / ${loading ? '…' : localDecks.length} DECKS`}
            right={<Tag solid kind={localDecks.length > 0 ? 'ok' : null}>{localDecks.length}</Tag>}
          >
            {loading ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; SEARCHING DECK INDEX<span className="blink">_</span>
              </div>
            ) : localDecks.length === 0 ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; NO INDEXED DECKS INCLUDE THIS CARD YET.
              </div>
            ) : (
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))',
                gap: 8,
              }}>
                {localDecks.slice(0, 60).map(d => {
                  const cmdrName = d.commander || ''
                  const art = cardArtUrl(cmdrName)
                  return (
                    <Link
                      key={`${d.owner}/${d.id}`}
                      to={`/decks/${d.owner}/${d.id}`}
                      className="card-page-deck-tile"
                      title={`${d.commander || d.name || d.id} (${d.owner})${d.role ? ' · ' + readableRole(d.role) : ''}`}
                    >
                      <div
                        className="card-page-deck-art"
                        style={art ? { backgroundImage: `url(${art})` } : undefined}
                      />
                      <div className="card-page-deck-meta">
                        <div className="card-page-deck-name">
                          {(d.commander || d.name || d.id || '').toUpperCase()}
                        </div>
                        <div className="card-page-deck-owner">
                          {(d.owner || '').toUpperCase()}
                          {d.role && (
                            <span style={{ marginLeft: 4, color: 'var(--ink-3)' }}>· {readableRole(d.role)}</span>
                          )}
                        </div>
                      </div>
                    </Link>
                  )
                })}
              </div>
            )}
            {localDecks.length > 60 && (
              <div className="t-xs muted" style={{ textAlign: 'center', marginTop: 8 }}>
                &gt; SHOWING 60 / {localDecks.length}
              </div>
            )}
          </Panel>
        </div>
      </div>
    </>
  )
}
