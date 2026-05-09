import { useState } from 'react'
import { Panel, Tag } from './chrome'
import CardLink from './CardLink'

// Shared brutalist row styling used by all three rationale panels.
const rowBase = {
  borderTop: '1px dashed var(--rule-2)',
  padding: '8px 0',
}

const headerBase = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  gap: 8,
  cursor: 'pointer',
  userSelect: 'none',
}

const detailGrid = {
  display: 'grid',
  gridTemplateColumns: '90px 1fr',
  gap: '4px 12px',
  marginTop: 8,
  paddingLeft: 16,
  borderLeft: '2px solid var(--rule-2)',
}

const detailLabel = {
  fontSize: 9,
  fontWeight: 700,
  letterSpacing: '0.08em',
  textTransform: 'uppercase',
  color: 'var(--ink-2)',
  paddingTop: 2,
}

function DetailRow({ label, children }) {
  if (children == null) return null
  if (Array.isArray(children) && children.length === 0) return null
  return (
    <>
      <span style={detailLabel}>{label}</span>
      <span className="t-xs" style={{ lineHeight: 1.5 }}>{children}</span>
    </>
  )
}

function Caret({ open }) {
  return (
    <span style={{
      fontSize: 10,
      color: 'var(--ink-2)',
      width: 14,
      textAlign: 'center',
      transition: 'transform 0.15s',
      transform: open ? 'rotate(90deg)' : 'rotate(0deg)',
    }}>▶</span>
  )
}

// ---------------------------------------------------------------------------
// CONSIDER CUTTING — explains why each cuttable card was flagged.
// ---------------------------------------------------------------------------

export function ConsiderCuttingRationale({ cuts }) {
  const [openIdx, setOpenIdx] = useState(null)
  const items = (cuts || []).filter(c => c && (c.name || typeof c === 'string'))
  if (items.length === 0) return null

  return (
    <Panel
      code="04.QR"
      title={`CONSIDER CUTTING / / ${items.length} CARDS`}
      right={<Tag kind="warn">RATIONALE</Tag>}
    >
      <div className="t-xs muted" style={{ marginBottom: 6 }}>
        Each cut shows what was detected, why it's flagged, and what cutting it frees up.
      </div>
      {items.map((c, i) => {
        // Tolerate both legacy string entries and structured objects so the
        // UI keeps rendering even if an older strategy.json is on disk.
        const name = typeof c === 'string' ? c : c.name
        const reason = typeof c === 'string' ? '' : (c.reason || '')
        const detected = typeof c === 'string' ? '' : (c.detected || '')
        const whyCut = typeof c === 'string' ? '' : (c.why_cut || '')
        const effect = typeof c === 'string' ? '' : (c.effect || '')
        const suggested = typeof c === 'string' ? [] : (c.suggested || [])
        const open = openIdx === i
        const hasDetail = detected || whyCut || effect || (suggested && suggested.length > 0)
        return (
          <div key={i} style={rowBase}>
            <div
              style={headerBase}
              onClick={() => hasDetail && setOpenIdx(open ? null : i)}
            >
              <span style={{ display: 'flex', alignItems: 'center', gap: 8, minWidth: 0 }}>
                {hasDetail && <Caret open={open} />}
                <CardLink name={name} className="t-md" style={{ fontWeight: 700, borderBottom: 'none' }}>
                  {name}
                </CardLink>
                {reason && (
                  <span className="t-xs muted" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    — {reason}
                  </span>
                )}
              </span>
              <Tag kind="warn">CUT</Tag>
            </div>
            {open && hasDetail && (
              <div style={detailGrid}>
                <DetailRow label="Detected">{detected}</DetailRow>
                <DetailRow label="Why">{whyCut}</DetailRow>
                <DetailRow label="If cut">{effect}</DetailRow>
                {suggested && suggested.length > 0 && (
                  <DetailRow label="Swap for">
                    <ul style={{ margin: 0, paddingLeft: 16 }}>
                      {suggested.map((s, si) => <li key={si}>{s}</li>)}
                    </ul>
                  </DetailRow>
                )}
              </div>
            )}
          </div>
        )
      })}
    </Panel>
  )
}

// ---------------------------------------------------------------------------
// VALUE ENGINE — explains why each engine was identified for this deck.
// ---------------------------------------------------------------------------

const REDUNDANCY_KIND = { HIGH: 'ok', MEDIUM: null, LOW: 'warn' }
const RECURSION_KIND = { infinite: 'ok', deep: 'ok', shallow: null, none: null }

export function ValueEngineRationale({ chains }) {
  const [openIdx, setOpenIdx] = useState(0) // first chain expanded by default
  const items = (chains || []).filter(Boolean)
  if (items.length === 0) return null

  return (
    <Panel
      code="04.ER"
      title={`VALUE ENGINES / / ${items.length} DETECTED`}
      right={<Tag kind="ok">RATIONALE</Tag>}
    >
      <div className="t-xs muted" style={{ marginBottom: 6 }}>
        Each engine is a multi-step resource pipeline detected from card flows.
      </div>
      {items.map((chain, i) => {
        const open = openIdx === i
        const rationale = chain.rationale || {}
        const totalPieces = (chain.steps || []).reduce((s, st) => s + (st.cards?.length || 0), 0)
        return (
          <div key={i} style={rowBase}>
            <div style={headerBase} onClick={() => setOpenIdx(open ? -1 : i)}>
              <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Caret open={open} />
                <span className="t-md" style={{ fontWeight: 700 }}>{chain.name}</span>
                <span className="t-xs muted">{chain.depth} steps · {totalPieces} pieces</span>
              </span>
              <span style={{ display: 'flex', gap: 4 }}>
                {chain.redundancy && (
                  <Tag kind={REDUNDANCY_KIND[chain.redundancy] ?? null}>
                    {chain.redundancy} REDUNDANCY
                  </Tag>
                )}
                {chain.recursion_depth && chain.recursion_depth !== 'none' && (
                  <Tag kind={RECURSION_KIND[chain.recursion_depth] ?? null}>
                    {chain.recursion_depth.toUpperCase()} RECURSION
                  </Tag>
                )}
              </span>
            </div>
            {open && (
              <div style={detailGrid}>
                <DetailRow label="Trigger">{rationale.trigger}</DetailRow>
                <DetailRow label="How">{rationale.how_it_works}</DetailRow>
                {rationale.key_pieces && rationale.key_pieces.length > 0 && (
                  <DetailRow label="Key pieces">
                    <span style={{ display: 'inline-flex', flexWrap: 'wrap', gap: 4 }}>
                      {rationale.key_pieces.map((kp, ki) => (
                        <CardLink key={ki} name={kp} className="t-xs" style={{ fontWeight: 700 }}>
                          {kp}
                        </CardLink>
                      ))}
                    </span>
                  </DetailRow>
                )}
                {(chain.steps || []).map((step, si) => (
                  <ChainStepRow key={si} step={step} weakest={si === chain.weakest_link} />
                ))}
              </div>
            )}
          </div>
        )
      })}
    </Panel>
  )
}

function ChainStepRow({ step, weakest }) {
  const cards = step.cards || []
  return (
    <>
      <span style={{ ...detailLabel, color: weakest ? 'var(--warn)' : 'var(--ink-2)' }}>
        [{step.label}]
      </span>
      <span className="t-xs" style={{ lineHeight: 1.5 }}>
        <span className="muted-2" style={{ fontSize: 9 }}>
          {step.from}→{step.to}{step.resource ? ` (${step.resource})` : ''}
        </span>
        <span style={{ marginLeft: 6 }}>
          {cards.length === 0 ? <span className="muted">— none</span> :
            cards.slice(0, 8).map((c, ci) => (
              <span key={ci}>
                {ci > 0 && <span className="muted-2">, </span>}
                <CardLink name={c} className="t-xs" style={{ borderBottom: 'none' }}>{c}</CardLink>
              </span>
            ))}
          {cards.length > 8 && <span className="muted-2"> +{cards.length - 8} more</span>}
        </span>
        {weakest && <Tag kind="warn" style={{ marginLeft: 8 }}>WEAKEST LINK</Tag>}
      </span>
    </>
  )
}

// ---------------------------------------------------------------------------
// WIN CONDITION — explains the detection logic for each win line.
// ---------------------------------------------------------------------------

const WINLINE_KIND = {
  infinite: 'bad',
  determined: 'bad',
  finisher: 'warn',
  combat: 'warn',
  commander_damage: 'ok',
  alt_wincon: 'bad',
}

export function WinConditionRationale({ winLines }) {
  const [openIdx, setOpenIdx] = useState(0)
  const items = (winLines || []).filter(Boolean)
  if (items.length === 0) return null

  return (
    <Panel
      code="04.WR"
      title={`WIN CONDITIONS / / ${items.length} LINES`}
      right={<Tag kind="ok">RATIONALE</Tag>}
    >
      <div className="t-xs muted" style={{ marginBottom: 6 }}>
        Each line shows the cards that form it, the conditions required, and how it resolves to a win.
      </div>
      {items.map((wl, i) => {
        const open = openIdx === i
        const rationale = wl.rationale || {}
        const tutorCount = (wl.tutor_paths || []).length
        return (
          <div key={i} style={rowBase}>
            <div style={headerBase} onClick={() => setOpenIdx(open ? -1 : i)}>
              <span style={{ display: 'flex', alignItems: 'center', gap: 8, minWidth: 0, flex: 1 }}>
                <Caret open={open} />
                <Tag kind={WINLINE_KIND[wl.type] ?? null} solid>{(wl.type || '').toUpperCase()}</Tag>
                <span className="t-md" style={{ fontWeight: 700, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {(wl.pieces || []).join(' + ')}
                </span>
              </span>
              {tutorCount > 0 && <Tag>{tutorCount} TUTOR PATH{tutorCount === 1 ? '' : 'S'}</Tag>}
            </div>
            {open && (
              <div style={detailGrid}>
                {rationale.forms && rationale.forms.length > 0 && (
                  <DetailRow label="Forms">
                    <span style={{ display: 'inline-flex', flexWrap: 'wrap', gap: 4 }}>
                      {rationale.forms.map((f, fi) => (
                        // Forms can be card names OR descriptive phrases (e.g.
                        // "10 creatures CMC 3+ in deck"). Only linkify when
                        // the string looks like a single card name.
                        /^[A-Z][^,]+$/.test(f) && f.length < 50
                          ? <CardLink key={fi} name={f} className="t-xs" style={{ fontWeight: 700 }}>{f}</CardLink>
                          : <span key={fi} className="t-xs">{f}</span>
                      )).reduce((acc, el, fi) => fi === 0 ? [el] : [...acc, <span key={`s${fi}`} className="muted-2">, </span>, el], [])}
                    </span>
                  </DetailRow>
                )}
                {rationale.conditions && rationale.conditions.length > 0 && (
                  <DetailRow label="Conditions">
                    <ul style={{ margin: 0, paddingLeft: 16 }}>
                      {rationale.conditions.map((c, ci) => <li key={ci}>{c}</li>)}
                    </ul>
                  </DetailRow>
                )}
                {rationale.resolves && rationale.resolves.length > 0 && (
                  <DetailRow label="Resolves">
                    <ul style={{ margin: 0, paddingLeft: 16 }}>
                      {rationale.resolves.map((r, ri) => <li key={ri}>{r}</li>)}
                    </ul>
                  </DetailRow>
                )}
                {!rationale.resolves && wl.description && (
                  <DetailRow label="Description">{wl.description}</DetailRow>
                )}
              </div>
            )}
          </div>
        )
      })}
    </Panel>
  )
}
