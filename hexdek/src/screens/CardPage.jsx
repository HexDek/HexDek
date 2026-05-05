import { useEffect, useMemo, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { Panel, KV, Tag, Tape, Btn } from '../components/chrome'
import { api, cardArtUrl } from '../services/api'

// CardPage — /cards/:cardName
//
// Three data sources, each independently loaded so a slow Scryfall
// fetch doesn't block the deck-usage section:
//
//   - Scryfall /cards/named?exact=...   → printed art, mana cost,
//     oracle text, type line, set, USD price.
//   - /api/decks?contains=...           → which decks include this
//     card (server-side substring filter).
//   - /api/decks/{owner}/{id}/analysis  → Freya card_roles, fetched
//     for each containing deck and merged into a tag set.
//
// The Freya merge is bounded (first 6 decks, sequential to keep the
// browser well-behaved). Roles surfaced this way are an aggregate
// signal — different decks can label the same card differently.

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

function deckKeyOf(d) {
  if (d?.id && d.id.includes('/')) return d.id
  if (d?.owner && d?.id) return `${d.owner}/${d.id}`
  return d?.id || ''
}

export default function CardPage() {
  const { cardName: rawName } = useParams()
  const navigate = useNavigate()
  const cardName = useMemo(() => decodeURIComponent(rawName || ''), [rawName])

  const [scry, setScry] = useState(null)
  const [scryErr, setScryErr] = useState(false)
  const [scryLoading, setScryLoading] = useState(true)

  const [decks, setDecks] = useState([])
  const [decksLoading, setDecksLoading] = useState(true)

  const [roles, setRoles] = useState([])

  // Scryfall fetch — independent of our backend.
  useEffect(() => {
    if (!cardName) return
    let cancelled = false
    setScryLoading(true)
    setScryErr(false)
    fetch(`https://api.scryfall.com/cards/named?exact=${encodeURIComponent(cardName)}`)
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`scryfall ${r.status}`)))
      .then(data => { if (!cancelled) setScry(data) })
      .catch(() => { if (!cancelled) setScryErr(true) })
      .finally(() => { if (!cancelled) setScryLoading(false) })
    return () => { cancelled = true }
  }, [cardName])

  // Decks containing this card. Backend already supports ?contains=
  // via handleListDecks (substring match against the deck file).
  useEffect(() => {
    if (!cardName) return
    let cancelled = false
    setDecksLoading(true)
    api.getDecks({ contains: cardName })
      .then(d => { if (!cancelled) setDecks(Array.isArray(d) ? d : []) })
      .catch(() => { if (!cancelled) setDecks([]) })
      .finally(() => { if (!cancelled) setDecksLoading(false) })
    return () => { cancelled = true }
  }, [cardName])

  // Freya role tags — sample up to 6 containing decks' analysis
  // sequentially and union the role tags they assign to this card.
  // Sequential rather than parallel keeps the request count bounded
  // when the contains-list returns dozens of decks.
  useEffect(() => {
    if (!cardName || decks.length === 0) { setRoles([]); return }
    let cancelled = false
    const sample = decks.slice(0, 6)
    const lcName = cardName.toLowerCase()
    const collected = new Set()
    ;(async () => {
      for (const d of sample) {
        if (cancelled) return
        const key = deckKeyOf(d)
        if (!key) continue
        try {
          const a = await api.getDeckAnalysis(key)
          const cr = a?.card_roles
          if (!cr) continue
          // card_roles can be either a flat {cardName: [roles]} map or
          // a {role: [cardNames]} grouping. Handle both.
          if (Array.isArray(cr[cardName]) || Array.isArray(cr[lcName])) {
            for (const r of (cr[cardName] || cr[lcName] || [])) collected.add(r)
            continue
          }
          for (const [role, names] of Object.entries(cr)) {
            if (!Array.isArray(names)) continue
            if (names.some(n => (n || '').toLowerCase() === lcName)) collected.add(role)
          }
        } catch {}
      }
      if (!cancelled) setRoles([...collected])
    })()
    return () => { cancelled = true }
  }, [cardName, decks])

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

  const upperName = cardName.toUpperCase()
  const heroArt = scry?.image_uris?.art_crop
    || scry?.card_faces?.[0]?.image_uris?.art_crop
    || cardArtUrl(cardName)

  const manaCost = scry?.mana_cost || scry?.card_faces?.[0]?.mana_cost || ''
  const typeLine = scry?.type_line || scry?.card_faces?.[0]?.type_line || '—'
  const oracle   = scry?.oracle_text || scry?.card_faces?.[0]?.oracle_text || ''
  const setName  = scry?.set_name || '—'
  const setCode  = (scry?.set || '').toUpperCase()
  const price    = priceFromScryfall(scry)

  return (
    <>
      <Tape
        left={`CARD / / ${upperName}`}
        mid={scryLoading ? 'LOADING' : (scryErr ? 'SCRYFALL UNAVAILABLE' : 'CARD RECORD')}
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
              ['CMC', scry?.cmc != null ? `${scry.cmc}` : '—'],
              ['SET', setCode || '—'],
              ['PRICE (USD)', price],
              ['LEGAL COMMANDER',
                scry?.legalities?.commander
                  ? <span style={{ color: scry.legalities.commander === 'legal' ? 'var(--ok)' : 'var(--danger)', fontWeight: 700 }}>
                      {scry.legalities.commander.toUpperCase()}
                    </span>
                  : '—',
              ],
              ['SOURCE', scryErr ? 'SCRYFALL OFFLINE' : 'SCRYFALL'],
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
                AGGREGATED FROM UP TO 6 DECKS THAT INCLUDE THIS CARD.
              </div>
            </Panel>
          )}
        </div>

        <div className="card-page-main">
          {/* Oracle text */}
          <Panel code="07.B" title="ORACLE TEXT" right={
            <span className="t-xs muted">{scryLoading ? 'LOADING' : (scryErr ? 'OFFLINE' : 'PRINTED')}</span>
          }>
            {scryLoading ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; FETCHING SCRYFALL RECORD<span className="blink">_</span>
              </div>
            ) : scryErr ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; SCRYFALL UNAVAILABLE — TRY AGAIN LATER.
              </div>
            ) : (
              <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.55, fontSize: 12 }}>
                {oracle || <span className="muted">&gt; NO ORACLE TEXT.</span>}
              </div>
            )}
          </Panel>

          {/* Used in — decks containing this card */}
          <Panel
            code="07.C"
            title={`USED IN / / ${decksLoading ? '…' : decks.length} DECKS`}
            right={<Tag solid kind={decks.length > 0 ? 'ok' : null}>{decks.length}</Tag>}
          >
            {decksLoading ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; SEARCHING DECK INDEX<span className="blink">_</span>
              </div>
            ) : decks.length === 0 ? (
              <div className="t-md muted" style={{ padding: '14px 0', textAlign: 'center' }}>
                &gt; NO INDEXED DECKS INCLUDE THIS CARD YET.
              </div>
            ) : (
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))',
                gap: 8,
              }}>
                {decks.slice(0, 60).map(d => {
                  const key = deckKeyOf(d)
                  const cmdrName = d.commander_card || d.commander || ''
                  const art = cardArtUrl(cmdrName || d.commander)
                  const owner = d.owner || (key.split('/')[0])
                  const id = d.id?.includes('/') ? d.id.split('/')[1] : d.id
                  return (
                    <Link
                      key={key}
                      to={`/decks/${owner}/${id}`}
                      className="card-page-deck-tile"
                      title={`${d.commander || d.name || id} (${owner})`}
                    >
                      <div
                        className="card-page-deck-art"
                        style={art ? { backgroundImage: `url(${art})` } : undefined}
                      />
                      <div className="card-page-deck-meta">
                        <div className="card-page-deck-name">
                          {(d.commander || d.name || id || '').toUpperCase()}
                        </div>
                        <div className="card-page-deck-owner">{(owner || '').toUpperCase()}</div>
                      </div>
                    </Link>
                  )
                })}
              </div>
            )}
            {decks.length > 60 && (
              <div className="t-xs muted" style={{ textAlign: 'center', marginTop: 8 }}>
                &gt; SHOWING 60 / {decks.length}
              </div>
            )}
          </Panel>
        </div>
      </div>
    </>
  )
}
