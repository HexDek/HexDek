import { useEffect, useMemo, useState } from 'react'
import { Panel, Tag } from './chrome'
import { API_BASE, cardArtUrl } from '../services/api'

// MatchupsPanel — head-to-head record per opposing commander for a
// single deck. Backed by GET /api/decks/{owner}/{id}/matchups (returns
// rows already sorted by games desc, wins desc).
//
// Renders three blocks:
//   1. ALL MATCHUPS table  — every opponent ever faced, tightest WR bar.
//   2. BEST MATCHUPS       — top 3 by win rate, min 3 games (otherwise
//                            small samples dominate the leaderboard).
//   3. WORST MATCHUPS      — bottom 3 by win rate, same min-games gate.
//
// The minimum-games filter is the natural fix for noisy aggregates: a
// 1-0 record looks "best" without the gate. We surface the threshold
// in the section header so it's not a hidden bias.

const MIN_GAMES_FOR_RANKING = 3

function winRateColor(wr) {
  if (wr >= 60) return 'var(--ok)'
  if (wr <= 40) return 'var(--danger)'
  return 'var(--warn)'
}

function WinRateBar({ winRate }) {
  // Bar fills 0–100 left-to-right; a thin marker at 50% gives the
  // viewer an instant "above/below break-even" reference.
  const pct = Math.max(0, Math.min(100, winRate))
  return (
    <div
      title={`${winRate.toFixed(1)}%`}
      style={{
        position: 'relative',
        height: 10,
        background: 'var(--bg-2)',
        border: '1px solid var(--rule-2)',
        overflow: 'hidden',
      }}
    >
      <div
        style={{
          position: 'absolute',
          inset: 0,
          width: `${pct}%`,
          background: winRateColor(winRate),
          opacity: 0.85,
          transition: 'width 200ms ease-out',
        }}
      />
      <div
        style={{
          position: 'absolute',
          left: '50%',
          top: 0,
          bottom: 0,
          width: 1,
          background: 'var(--ink-3)',
          opacity: 0.6,
        }}
      />
    </div>
  )
}

function OpponentArt({ commander }) {
  const url = commander ? cardArtUrl(commander) : null
  return (
    <div
      className={url ? '' : 'hatch'}
      style={{
        width: 28, height: 28,
        flexShrink: 0,
        overflow: 'hidden',
        border: '1px solid var(--rule-2)',
        background: 'var(--bg-2)',
      }}
    >
      {url && (
        <img
          src={url}
          alt=""
          loading="lazy"
          style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
          onError={(e) => { e.target.style.display = 'none'; e.target.parentElement.classList.add('hatch') }}
        />
      )}
    </div>
  )
}

function MatchupRow({ row, dense }) {
  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: dense
        ? '28px 1fr auto'
        : '28px 1fr 36px 36px 36px 70px',
      gap: 8,
      alignItems: 'center',
      padding: '5px 0',
      borderBottom: '1px dashed var(--rule)',
      fontSize: 11,
    }}>
      <OpponentArt commander={row.opponent_commander} />
      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontWeight: 600 }}>
        {row.opponent_commander}
      </span>
      {dense ? (
        <span style={{ display: 'flex', alignItems: 'center', gap: 6, minWidth: 100 }}>
          <span className="t-xs muted" style={{ minWidth: 38, textAlign: 'right' }}>
            {row.wins}-{row.losses}
            {row.draws > 0 && <span className="muted-2"> ({row.draws}D)</span>}
          </span>
          <span style={{ flex: 1, minWidth: 60 }}>
            <WinRateBar winRate={row.win_rate} />
          </span>
          <span style={{
            color: winRateColor(row.win_rate),
            fontWeight: 700,
            minWidth: 38,
            textAlign: 'right',
          }}>
            {row.win_rate.toFixed(0)}%
          </span>
        </span>
      ) : (
        <>
          <span className="t-xs" style={{ textAlign: 'right' }}>{row.games}</span>
          <span className="t-xs" style={{ textAlign: 'right', color: 'var(--ok)' }}>{row.wins}</span>
          <span className="t-xs" style={{ textAlign: 'right', color: 'var(--danger)' }}>{row.losses}</span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{ flex: 1, minWidth: 30 }}>
              <WinRateBar winRate={row.win_rate} />
            </span>
            <span style={{
              color: winRateColor(row.win_rate),
              fontWeight: 700,
              minWidth: 30,
              textAlign: 'right',
              fontSize: 10,
            }}>
              {row.win_rate.toFixed(0)}%
            </span>
          </span>
        </>
      )}
    </div>
  )
}

export default function MatchupsPanel({ owner, id, code = '04.MU' }) {
  const [rows, setRows] = useState(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (!owner || !id) return
    let alive = true
    setRows(null)
    setError(false)
    fetch(`${API_BASE}/api/decks/${encodeURIComponent(owner)}/${encodeURIComponent(id)}/matchups`)
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`${r.status}`)))
      .then(d => { if (alive) setRows(Array.isArray(d?.matchups) ? d.matchups : []) })
      .catch(() => { if (alive) { setError(true); setRows([]) } })
    return () => { alive = false }
  }, [owner, id])

  // Best / worst: rank only opponents with enough games to be meaningful.
  // Backend sorts by games desc; we re-sort by win_rate within the
  // qualified subset for the leaderboards.
  const { best, worst } = useMemo(() => {
    if (!rows || rows.length === 0) return { best: [], worst: [] }
    const qualified = rows.filter(r => r.games >= MIN_GAMES_FOR_RANKING)
    const byWinRate = [...qualified].sort((a, b) => b.win_rate - a.win_rate)
    return {
      best: byWinRate.slice(0, 3),
      worst: byWinRate.slice(-3).reverse(),
    }
  }, [rows])

  if (rows === null) {
    return (
      <Panel code={code} title="MATCHUPS / / LOADING">
        <div className="t-md muted" style={{ textAlign: 'center', padding: 12 }}>
          &gt; FETCHING<span className="blink">_</span>
        </div>
      </Panel>
    )
  }

  if (error) {
    return (
      <Panel code={code} title="MATCHUPS / / OFFLINE">
        <div className="t-xs muted" style={{ textAlign: 'center', padding: 12 }}>
          &gt; MATCHUP HISTORY UNAVAILABLE.
        </div>
      </Panel>
    )
  }

  if (rows.length === 0) {
    return (
      <Panel code={code} title="MATCHUPS / / NO HISTORY">
        <div className="t-xs muted" style={{ textAlign: 'center', padding: 12 }}>
          &gt; NO RECORDED GAMES YET. PLAY A FEW SHOWMATCHES TO SEED THE LEDGER.
        </div>
      </Panel>
    )
  }

  return (
    <Panel
      code={code}
      title={`MATCHUPS / / ${rows.length} OPPONENTS`}
      right={<Tag solid>{rows.reduce((s, r) => s + r.games, 0)} GAMES</Tag>}
    >
      {/* Best */}
      {best.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <div style={{
            fontSize: 9, fontWeight: 700, letterSpacing: '0.14em',
            color: 'var(--ok)', textTransform: 'uppercase',
            marginBottom: 4, paddingBottom: 3,
            borderBottom: '1px dashed color-mix(in srgb, var(--ok) 35%, var(--rule-2))',
          }}>
            BEST MATCHUPS
            <span className="muted-2" style={{ fontWeight: 400, marginLeft: 6 }}>
              MIN {MIN_GAMES_FOR_RANKING} GAMES
            </span>
          </div>
          {best.map((r) => <MatchupRow key={`b-${r.opponent_commander}`} row={r} dense />)}
        </div>
      )}

      {/* Worst */}
      {worst.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <div style={{
            fontSize: 9, fontWeight: 700, letterSpacing: '0.14em',
            color: 'var(--danger)', textTransform: 'uppercase',
            marginBottom: 4, paddingBottom: 3,
            borderBottom: '1px dashed color-mix(in srgb, var(--danger) 35%, var(--rule-2))',
          }}>
            WORST MATCHUPS
            <span className="muted-2" style={{ fontWeight: 400, marginLeft: 6 }}>
              MIN {MIN_GAMES_FOR_RANKING} GAMES
            </span>
          </div>
          {worst.map((r) => <MatchupRow key={`w-${r.opponent_commander}`} row={r} dense />)}
        </div>
      )}

      {/* Full ledger */}
      <div>
        <div style={{
          display: 'grid',
          gridTemplateColumns: '28px 1fr 36px 36px 36px 70px',
          gap: 8,
          padding: '4px 0',
          borderBottom: '1px solid var(--rule-2)',
          fontSize: 9,
          letterSpacing: '0.1em',
          color: 'var(--ink-3)',
          fontWeight: 700,
        }}>
          <span></span>
          <span>OPPONENT</span>
          <span style={{ textAlign: 'right' }}>G</span>
          <span style={{ textAlign: 'right' }}>W</span>
          <span style={{ textAlign: 'right' }}>L</span>
          <span style={{ textAlign: 'right' }}>WIN %</span>
        </div>
        {rows.map((r) => <MatchupRow key={r.opponent_commander} row={r} />)}
      </div>
    </Panel>
  )
}
