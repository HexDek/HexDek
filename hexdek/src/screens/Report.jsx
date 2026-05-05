import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { Panel, KV, Bar, Tag, Btn, Tape } from '../components/chrome'
import { api, cardArtUrl } from '../services/api'

// ── API gaps for full report fidelity ──────────────────────────────────
// CompletedGame currently exposes: game_id, commanders[], deck_keys[],
// winner, winner_name, turns, end_reason, finished_at, final_seats[].
// To remove the mocks below we need the backend to also return:
//   * log[] — LogEntry[] retained on CompletedGame (already exists on
//     GameSnapshot during live play; just persist a trimmed copy)
//   * life_history[] — per-seat life total per turn (or per phase)
//   * per_card_impact[] — { name, owner, plays_count, damage_dealt,
//     enabled_combo, etc. } for proper MVP ranking
//   * commander_damage[][] — seat × seat matrix for the 21+ check
// Until those land, the timeline + life sparkline + MVP heuristic are
// labeled "MOCK / HEURISTIC" inline so reviewers know what's real.
// ───────────────────────────────────────────────────────────────────────

const Stat2 = ({ k, v, big }) => (
  <div>
    <div className="t-xs muted">{k}</div>
    <div className={big ? 't-3xl' : 't-2xl'} style={{ fontWeight: 800, marginTop: 4 }}>{v}</div>
  </div>
)

const MockTag = () => (
  <Tag kind="warn" style={{ fontSize: 8, padding: '1px 5px' }}>MOCK · NEEDS API</Tag>
)

const HeuristicTag = () => (
  <Tag kind="warn" style={{ fontSize: 8, padding: '1px 5px' }}>HEURISTIC</Tag>
)

// classifyWinCondition derives a human-readable win-condition label from
// CompletedGame.end_reason. Honest mapping; nothing mocked here.
const WIN_CONDITION_MAP = {
  sba_704_5a: { label: 'COMBAT / BURN DAMAGE', detail: 'OPPONENTS REDUCED TO 0 LIFE', kind: 'damage' },
  sba_704_5b: { label: 'MILL — DECKED OUT', detail: 'OPPONENTS DREW FROM EMPTY LIBRARY', kind: 'mill' },
  sba_704_5c: { label: 'INFECT / POISON', detail: '10+ POISON COUNTERS DELIVERED', kind: 'poison' },
  sba_704_5d: { label: 'COMMANDER DAMAGE', detail: '21+ DAMAGE FROM A SINGLE COMMANDER', kind: 'commander_damage' },
  concession: { label: 'SCOOP / CONCESSION', detail: 'OPPONENT CONCEDED', kind: 'concession' },
  turn_limit: { label: 'TURN LIMIT REACHED', detail: 'GAME ENDED ON SCHEDULED CAP', kind: 'stall' },
}
const classifyWinCondition = (endReason) => {
  if (!endReason) return { label: 'UNKNOWN', detail: 'NO END REASON RECORDED', kind: 'unknown' }
  const key = endReason.toLowerCase()
  return WIN_CONDITION_MAP[key] || {
    label: endReason.replace(/_/g, ' ').toUpperCase(),
    detail: 'STATE-BASED ELIMINATION',
    kind: 'sba',
  }
}

// pickMVP picks a best-guess MVP from the winner's final battlefield.
// Heuristic order: commander → highest power creature → first non-land
// permanent → first permanent. Real MVP scoring requires per-card
// impact telemetry from the engine (see API gaps note above).
const pickMVP = (winnerSeat) => {
  if (!winnerSeat?.battlefield?.length) return null
  const perms = winnerSeat.battlefield
  const cmdr = perms.find(p => p.is_commander)
  if (cmdr) return { perm: cmdr, reason: 'COMMANDER ON THE BATTLEFIELD AT VICTORY' }
  const nonLand = perms.filter(p => !p.is_land)
  if (nonLand.length) {
    const ranked = [...nonLand].sort((a, b) => (b.power || 0) - (a.power || 0))
    if (ranked[0].power > 0) {
      return { perm: ranked[0], reason: `HIGHEST POWER NON-LAND (${ranked[0].power}/${ranked[0].toughness ?? '?'})` }
    }
    return { perm: nonLand[0], reason: 'FIRST NON-LAND PERMANENT' }
  }
  return { perm: perms[0], reason: 'ONLY PERMANENT ON BOARD' }
}

// mockTimeline synthesizes a plausible turn-by-turn key-play list keyed
// off real CompletedGame data (turns, winner, end_reason). Replace once
// CompletedGame.log[] is persisted server-side.
const mockTimeline = (game, commanders) => {
  if (!game) return []
  const total = Math.max(1, game.turns || 1)
  const winner = game.winner ?? 0
  const winName = (commanders[winner] || 'WINNER').split(',')[0].toUpperCase()
  const otherSeats = commanders
    .map((c, i) => ({ c, i }))
    .filter(s => s.i !== winner)
  const t = (frac) => Math.max(1, Math.round(total * frac))
  const cond = classifyWinCondition(game.end_reason)
  const finishLine = {
    'damage':            `${winName} CLOSES OUT WITH LETHAL DAMAGE`,
    'mill':              `${winName} EMPTIES THE LAST OPPONENT'S LIBRARY`,
    'poison':            `${winName} REACHES 10 POISON ON FINAL TARGET`,
    'commander_damage':  `${winName} DEALS 21+ COMMANDER DAMAGE`,
    'concession':        `${winName} ACCEPTS CONCESSION`,
    'stall':             `GAME REACHES TURN ${total} CAP`,
  }[cond.kind] || `${winName} SECURES VICTORY`
  return [
    { turn: 1,         seat: 0,      action: 'OPENING HANDS KEPT — ALL SEATS MULLIGAN-FREE', kind: 'open' },
    { turn: t(0.25),   seat: winner, action: `${winName} CASTS COMMANDER`, kind: 'cmdr' },
    { turn: t(0.4),    seat: otherSeats[0]?.i ?? 0, action: `EARLY THREAT FROM ${(otherSeats[0]?.c || '').split(',')[0].toUpperCase()}`, kind: 'threat' },
    { turn: t(0.6),    seat: winner, action: `${winName} STABILIZES THE BOARD`, kind: 'stabilize' },
    { turn: t(0.85),   seat: otherSeats[1]?.i ?? otherSeats[0]?.i ?? 0, action: `LATE PLAY FROM ${((otherSeats[1] || otherSeats[0])?.c || '').split(',')[0].toUpperCase()}`, kind: 'threat' },
    { turn: total,     seat: winner, action: finishLine, kind: 'win' },
  ]
}

// mockLifeCurve synthesizes a per-turn life sparkline. Real data needs
// a life_history field on CompletedGame.
const mockLifeCurve = (turns, finalLife, isWinner, lost) => {
  const T = Math.max(2, turns || 1)
  const start = 40
  const end = lost ? 0 : (finalLife != null ? finalLife : (isWinner ? Math.max(20, finalLife || 25) : 5))
  const pts = []
  for (let i = 0; i <= T; i++) {
    const t = i / T
    // Slight noise so the line doesn't look perfectly linear.
    const noise = (Math.sin(i * 1.7) + Math.cos(i * 0.9)) * 1.5
    const v = Math.max(0, Math.min(40, start + (end - start) * t + (i > 0 && i < T ? noise : 0)))
    pts.push(v)
  }
  return pts
}

const Sparkline = ({ points, accent = 'var(--ok)', height = 36 }) => {
  if (!points?.length) return null
  const w = 120
  const h = height
  const maxV = 40
  const step = w / (points.length - 1 || 1)
  const path = points.map((v, i) => {
    const x = i * step
    const y = h - (v / maxV) * h
    return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)} ${y.toFixed(1)}`
  }).join(' ')
  return (
    <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} style={{ display: 'block' }}>
      <line x1="0" y1={h - 1} x2={w} y2={h - 1} stroke="var(--rule-2)" strokeWidth="1" strokeDasharray="2 3" />
      <line x1="0" y1={h * 0.5} x2={w} y2={h * 0.5} stroke="var(--rule-2)" strokeWidth="1" strokeDasharray="2 3" opacity="0.5" />
      <path d={path} fill="none" stroke={accent} strokeWidth="1.5" />
    </svg>
  )
}

/* ─── Deck Context Selector (top bar) ────────────────────────── */
const DeckSelector = ({ commanders, selected, onSelect }) => {
  // Deduplicate commander names from all games
  const unique = [...new Set(commanders)].sort()
  if (unique.length === 0) return null

  return (
    <div style={{
      display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap',
      padding: '10px 18px', borderBottom: '1px solid var(--rule-2)',
    }}>
      <span className="t-xs muted" style={{ marginRight: 4 }}>DECK CONTEXT:</span>
      <Tag solid={!selected} onClick={() => onSelect(null)} style={{ cursor: 'pointer' }}>ALL</Tag>
      {unique.map(c => (
        <Tag
          key={c}
          solid={selected === c}
          onClick={() => onSelect(selected === c ? null : c)}
          style={{ cursor: 'pointer' }}
        >
          {c.split(',')[0].toUpperCase()}
        </Tag>
      ))}
    </div>
  )
}

/* ─── Per-Deck Stats Panel ───────────────────────────────────── */
const DeckStatsPanel = ({ games, elo, selectedDeck }) => {
  if (!selectedDeck) return null

  // Find ELO entry for this commander
  const eloEntry = elo.find(e =>
    e.commander?.toLowerCase() === selectedDeck.toLowerCase() ||
    e.commander?.toLowerCase().startsWith(selectedDeck.split(',')[0].trim().toLowerCase())
  )

  // Calculate stats from filtered games
  const wins = games.filter(g => g.winner >= 0 && g.commanders?.[g.winner]?.toLowerCase() === selectedDeck.toLowerCase()).length
  const totalFiltered = games.length
  const losses = totalFiltered - wins
  const avgTurns = totalFiltered > 0
    ? Math.round(games.reduce((sum, g) => sum + (g.turns || 0), 0) / totalFiltered * 10) / 10
    : 0
  const wr = totalFiltered > 0 ? Math.round(wins / totalFiltered * 1000) / 10 : 0

  return (
    <div className="panel" style={{ padding: 0, gridColumn: '1 / -1' }}>
      <div className="panel-hd">
        <span>DECK STATS / / {selectedDeck.split(',')[0].toUpperCase()}</span>
        <span>{totalFiltered} GAMES</span>
      </div>
      <div className="deck-stats-row" style={{ padding: '14px 22px' }}>
        <div>
          <div className="t-xs muted">GAMES</div>
          <div className="t-2xl" style={{ fontWeight: 800, marginTop: 4 }}>{totalFiltered}</div>
        </div>
        <div>
          <div className="t-xs muted">WINS</div>
          <div className="t-2xl" style={{ fontWeight: 800, marginTop: 4, color: 'var(--ok)' }}>{wins}</div>
        </div>
        <div>
          <div className="t-xs muted">LOSSES</div>
          <div className="t-2xl" style={{ fontWeight: 800, marginTop: 4, color: 'var(--danger)' }}>{losses}</div>
        </div>
        <div>
          <div className="t-xs muted">WIN RATE</div>
          <div className="t-2xl" style={{ fontWeight: 800, marginTop: 4 }}>{wr}%</div>
        </div>
        <div>
          <div className="t-xs muted">AVG TURNS</div>
          <div className="t-2xl" style={{ fontWeight: 800, marginTop: 4 }}>{avgTurns}</div>
        </div>
        <div>
          <div className="t-xs muted">ELO</div>
          <div className="t-2xl" style={{ fontWeight: 800, marginTop: 4 }}>
            {eloEntry ? Math.round(eloEntry.rating) : '—'}
            {eloEntry && (
              <span className="t-xs" style={{ marginLeft: 6, color: eloEntry.delta >= 0 ? 'var(--ok)' : 'var(--danger)' }}>
                {eloEntry.delta >= 0 ? '+' : ''}{eloEntry.delta}
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default function Report() {
  const { gameId } = useParams()
  const [game, setGame] = useState(null)
  const [games, setGames] = useState([])
  const [elo, setElo] = useState([])
  const [selectedDeck, setSelectedDeck] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      try {
        // Fetch ELO data
        api.getLiveELO().then(setElo).catch(() => {})

        if (gameId) {
          const g = await api.getGame(gameId)
          setGame(g)
        } else {
          const list = await api.getGames(1)
          if (list?.length > 0) {
            setGame(list[0])
          }
        }
        const full = await api.getGames(50)
        setGames(full || [])
      } catch (err) {
        console.warn('Report load failed:', err.message)
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [gameId])

  if (loading) {
    return (
      <>
        <Tape left="POST-GAME REPORT" mid="LOADING" right="" />
        <div style={{ padding: 36, textAlign: 'center' }}>
          <div className="t-md muted">&gt; LOADING REPORT<span className="blink">_</span></div>
        </div>
      </>
    )
  }

  // Collect all unique commander names across games
  const allCommanders = []
  for (const g of games) {
    if (g.commanders) {
      for (const c of g.commanders) {
        if (c && !allCommanders.includes(c)) allCommanders.push(c)
      }
    }
  }

  // Filter games by selected deck
  const filteredGames = selectedDeck
    ? games.filter(g => g.commanders?.some(c => c?.toLowerCase() === selectedDeck.toLowerCase()))
    : games

  if (!game && filteredGames.length === 0) {
    return (
      <>
        <Tape left="POST-GAME REPORT" mid="NO DATA" right="" />
        <DeckSelector commanders={allCommanders} selected={selectedDeck} onSelect={setSelectedDeck} />
        <div style={{ padding: 36, textAlign: 'center' }}>
          <div className="t-md muted">&gt; NO COMPLETED GAMES YET — WAIT FOR SHOWMATCH TO FINISH<span className="blink">_</span></div>
        </div>
      </>
    )
  }

  // Use the first filtered game as the featured game if no specific gameId
  const featuredGame = gameId ? game : (filteredGames[0] || game)

  const commanders = featuredGame?.commanders || []
  const seats = featuredGame?.final_seats || []
  const winnerIdx = featuredGame?.winner
  const winnerName = featuredGame?.winner_name || 'DRAW'
  const isVictory = winnerIdx >= 0

  return (
    <>
      <Tape
        left={featuredGame ? `POST-GAME / / GAME #${featuredGame.game_id}` : 'POST-GAME REPORT'}
        mid={isVictory ? 'VICTORY' : 'DRAW'}
        right={featuredGame?.finished_at ? new Date(featuredGame.finished_at).toLocaleString().toUpperCase() : ''}
      />

      {/* Deck context selector */}
      <DeckSelector commanders={allCommanders} selected={selectedDeck} onSelect={setSelectedDeck} />

      <div className="report-grid">
        {/* Per-deck stats if a deck is selected */}
        <DeckStatsPanel games={filteredGames} elo={elo} selectedDeck={selectedDeck} />

        {/* Result block */}
        {featuredGame && (
          <div className="panel" style={{ padding: 0, gridColumn: '1 / -1' }}>
            <div className="panel-hd"><span>RESULT BLOCK</span><span>GAME.{featuredGame.game_id}</span></div>
            <div className="result-block-grid" style={{ padding: '18px 22px' }}>
              <div>
                <div className="t-xs muted">OUTCOME</div>
                <div className="t-3xl" style={{ color: isVictory ? 'var(--ok)' : 'var(--warn)', fontWeight: 800 }}>
                  {isVictory ? 'VICTORY' : 'DRAW'}
                </div>
                <div className="t-md muted" style={{ marginTop: 6 }}>
                  TURN {featuredGame.turns} · {(featuredGame.end_reason || '').replace(/_/g, ' ').toUpperCase()}
                </div>
                <div className="t-xs muted-2" style={{ marginTop: 4 }}>
                  WINNER: {winnerName.toUpperCase()}
                </div>
              </div>
              <Stat2 k="TURNS" v={String(featuredGame.turns)} />
              <Stat2 k="PLAYERS" v={String(commanders.length)} big />
              <Stat2 k="END" v={(featuredGame.end_reason || '?').replace(/_/g, ' ').toUpperCase().slice(0, 12)} />
            </div>
          </div>
        )}

        {/* Final board state — all seats */}
        {featuredGame && (
          <div className="panel" style={{ gridColumn: '1 / -1', padding: 0 }}>
            <div className="panel-hd"><span>FINAL BOARD STATE</span><span>{commanders.length} SEATS</span></div>
            <div style={{ padding: '18px 22px' }}>
              <div className="grid col-4 gap-4">
                {seats.map((s, i) => {
                  const isWinner = i === winnerIdx
                  const cmdr = commanders[i] || 'UNKNOWN'
                  const perms = s.battlefield || []
                  const artUrl = cardArtUrl(cmdr)

                  return (
                    <div key={i} className="panel" style={{ padding: 0, borderColor: isWinner ? 'var(--ok)' : s.lost ? 'var(--danger)' : 'var(--rule-2)' }}>
                      {artUrl && (
                        <div style={{
                          height: 80,
                          backgroundImage: `url(${artUrl})`,
                          backgroundSize: 'cover', backgroundPosition: 'center',
                          borderBottom: '1px solid var(--rule-2)',
                          opacity: s.lost && !isWinner ? 0.3 : 0.8,
                        }} />
                      )}
                      <div style={{ padding: '8px 10px' }}>
                        <div className="flex justify-between items-center" style={{ marginBottom: 4 }}>
                          <span className="t-xs muted">SEAT.{String(i + 1).padStart(2, '0')}</span>
                          {isWinner && <Tag kind="ok" solid>WINNER</Tag>}
                          {s.lost && !isWinner && <Tag kind="bad">ELIMINATED</Tag>}
                          {!s.lost && !isWinner && <Tag>ALIVE</Tag>}
                        </div>
                        <div className="t-md" style={{ fontWeight: 700, lineHeight: 1.2 }}>
                          {cmdr.toUpperCase()}
                        </div>
                        <div className="hr" style={{ margin: '8px 0' }} />
                        <KV rows={[
                          ['LIFE', String(s.life)],
                          ['HAND', String(s.hand_size)],
                          ['LIBRARY', String(s.library_size)],
                          ['GRAVEYARD', String(s.gy_size)],
                          ['BATTLEFIELD', String(perms.length)],
                        ]} />
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        )}

        {/* Battlefield breakdown for winner */}
        {featuredGame && isVictory && seats[winnerIdx] && (
          <Panel code="RPT.A" title={`WINNER BATTLEFIELD / / ${winnerName.toUpperCase()}`}>
            <div className="t-xs muted" style={{ marginBottom: 6 }}>PERMANENTS AT GAME END</div>
            {(seats[winnerIdx].battlefield || []).length === 0 ? (
              <div className="t-xs muted-2">— NO PERMANENTS —</div>
            ) : (
              (seats[winnerIdx].battlefield || []).map((p, i) => (
                <div key={i} style={{ display: 'grid', gridTemplateColumns: '1fr 60px 60px', padding: '4px 0', borderBottom: '1px dashed var(--rule-2)', alignItems: 'center', gap: 8 }}>
                  <span className="t-xs" style={{ fontWeight: p.is_commander ? 700 : 400, color: p.is_commander ? 'var(--warn)' : 'var(--ink)' }}>
                    {p.is_commander ? '* ' : ''}{p.name?.toUpperCase()}
                  </span>
                  <span className="t-xs muted">{p.type || '—'}</span>
                  <span className="t-xs muted text-right">{p.tapped ? 'TAPPED' : 'UNTAPPED'}</span>
                </div>
              ))
            )}
          </Panel>
        )}

        {/* Game history list (filtered by deck) */}
        {filteredGames.length > 0 && (
          <Panel code="RPT.B" title={`${selectedDeck ? selectedDeck.split(',')[0].toUpperCase() + ' ' : ''}GAME LOG / / ${filteredGames.length} GAMES`}
            style={featuredGame && isVictory && seats[winnerIdx] ? {} : { gridColumn: '1 / -1' }}
          >
            <div style={{ maxHeight: 400, overflow: 'auto' }}>
              {filteredGames.map((g, i) => {
                const isWin = g.winner >= 0
                const winnerCmdr = isWin && g.commanders?.[g.winner] ? g.commanders[g.winner] : null
                // If filtering by deck, highlight if the selected deck won
                const deckWon = selectedDeck && winnerCmdr?.toLowerCase() === selectedDeck.toLowerCase()

                return (
                  <div key={i} style={{
                    display: 'grid', gridTemplateColumns: '50px 1fr 80px 60px',
                    padding: '8px 0',
                    borderBottom: i < filteredGames.length - 1 ? '1px dashed var(--rule-2)' : 'none',
                    alignItems: 'center', gap: 8,
                    opacity: selectedDeck && !deckWon && isWin ? 0.5 : 1,
                  }}>
                    <span className="t-xs muted-2">#{g.game_id}</span>
                    <div>
                      <div className="t-md" style={{ fontWeight: 600 }}>{g.winner_name?.toUpperCase() || 'DRAW'}</div>
                      <div className="t-xs muted">T{g.turns} · {(g.end_reason || '').replace(/_/g, ' ')}</div>
                    </div>
                    {selectedDeck ? (
                      <Tag kind={deckWon ? 'ok' : isWin ? 'bad' : 'warn'} solid>
                        {deckWon ? 'WIN' : isWin ? 'LOSS' : 'DRAW'}
                      </Tag>
                    ) : (
                      <Tag kind={isWin ? 'ok' : 'warn'} solid>{isWin ? 'WIN' : 'DRAW'}</Tag>
                    )}
                    <span className="t-xs muted text-right">{g.commanders?.length || 0}P</span>
                  </div>
                )
              })}
            </div>
          </Panel>
        )}

        {/* ── ANALYSIS BLOCK ──
           Four sections derived from CompletedGame: win condition (real),
           timeline (mock), MVP card (heuristic), per-seat performance
           (final stats real, life curve mock). Mocked sections wear a
           tag so reviewers know what's synthetic. */}

        {/* Win condition analysis — derived from end_reason. Real data. */}
        {featuredGame && isVictory && (() => {
          const wc = classifyWinCondition(featuredGame.end_reason)
          const eliminated = seats
            .map((s, i) => ({ s, i, name: commanders[i] }))
            .filter(x => x.s.lost && x.i !== winnerIdx)
          return (
            <Panel code="RPT.C" title="WIN CONDITION" style={{ gridColumn: '1 / -1' }}
              right={<Tag solid kind="ok">{wc.label}</Tag>}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1.6fr', gap: 22, padding: '4px 0' }}>
                <div>
                  <div className="t-xs muted">RESOLUTION</div>
                  <div className="t-2xl" style={{ fontWeight: 800, marginTop: 4, color: 'var(--ok)' }}>{wc.label}</div>
                  <div className="t-md muted" style={{ marginTop: 8, lineHeight: 1.55, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                    &gt; {wc.detail}
                    <br />
                    &gt; CR.{(featuredGame.end_reason || '').replace(/_/g, '.')} · TURN {featuredGame.turns}
                    <br />
                    &gt; WINNER: {winnerName.toUpperCase()}
                  </div>
                </div>
                <div>
                  <div className="t-xs muted" style={{ marginBottom: 6 }}>ELIMINATED SEATS</div>
                  {eliminated.length === 0 ? (
                    <div className="t-xs muted-2">— NONE — </div>
                  ) : eliminated.map(x => (
                    <div key={x.i} style={{ display: 'grid', gridTemplateColumns: '1fr auto', padding: '5px 0', borderBottom: '1px dashed var(--rule-2)', alignItems: 'center', gap: 8 }}>
                      <div>
                        <div className="t-md" style={{ fontWeight: 700 }}>{x.name?.toUpperCase()}</div>
                        <div className="t-xs muted">SEAT.{String(x.i + 1).padStart(2, '0')} · {(x.s.loss_reason || x.s.LossReason || 'STATE-BASED').replace(/_/g, ' ').toUpperCase()}</div>
                      </div>
                      <Tag kind="bad">ELIM</Tag>
                    </div>
                  ))}
                </div>
              </div>
            </Panel>
          )
        })()}

        {/* MVP card — heuristic pick from winner's battlefield. */}
        {featuredGame && isVictory && (() => {
          const mvp = pickMVP(seats[winnerIdx])
          if (!mvp) return null
          const art = cardArtUrl(mvp.perm.name)
          return (
            <Panel code="RPT.D" title="MVP CARD" right={<HeuristicTag />}>
              <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
                <div style={{ width: 96, aspectRatio: '5/7', flexShrink: 0, border: '1px solid var(--rule-2)', overflow: 'hidden', background: 'var(--bg-2)' }}
                  className={art ? '' : 'hatch'}>
                  {art && <img src={art} alt={mvp.perm.name} style={{ width: '100%', height: '100%', objectFit: 'cover', filter: 'saturate(0.7) contrast(1.05)' }} onError={e => { e.target.style.display = 'none'; e.target.parentElement.classList.add('hatch') }} />}
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div className="t-xs muted">CARD</div>
                  <div className="t-xl" style={{ fontWeight: 800, marginTop: 2, lineHeight: 1.1 }}>
                    {mvp.perm.is_commander && <span style={{ color: 'var(--warn)' }}>★ </span>}
                    {mvp.perm.name?.toUpperCase()}
                  </div>
                  <div className="t-xs muted" style={{ marginTop: 4 }}>
                    {(mvp.perm.type || 'PERMANENT').toUpperCase()}
                    {mvp.perm.power != null ? ` · ${mvp.perm.power}/${mvp.perm.toughness ?? '?'}` : ''}
                  </div>
                  <div className="hr" style={{ margin: '10px 0' }} />
                  <div className="t-xs muted">SELECTION REASON</div>
                  <div className="t-md" style={{ marginTop: 4, lineHeight: 1.5, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                    &gt; {mvp.reason}
                  </div>
                  <div className="t-xs muted-2" style={{ marginTop: 8, lineHeight: 1.5 }}>
                    HEURISTIC PICK FROM FINAL BOARD STATE. PROPER MVP RANKING NEEDS PER-CARD IMPACT TELEMETRY (DAMAGE DEALT, COMBO ENABLED, ETC.) FROM THE ENGINE.
                  </div>
                </div>
              </div>
            </Panel>
          )
        })()}

        {/* Game timeline — key plays per turn. Mock until log[] persists. */}
        {featuredGame && (() => {
          const timeline = mockTimeline(featuredGame, commanders)
          return (
            <Panel code="RPT.E" title="GAME TIMELINE" right={<MockTag />}
              style={{ gridColumn: featuredGame && isVictory ? 'auto' : '1 / -1' }}>
              <div style={{ position: 'relative', paddingLeft: 18 }}>
                <div style={{ position: 'absolute', left: 6, top: 4, bottom: 4, width: 1, background: 'var(--rule-2)' }} />
                {timeline.map((entry, i) => {
                  const cmdr = commanders[entry.seat] || ''
                  const seatColor = entry.seat === winnerIdx ? 'var(--ok)' : entry.kind === 'win' ? 'var(--ok)' : 'var(--ink-2)'
                  return (
                    <div key={i} style={{ display: 'grid', gridTemplateColumns: '50px 1fr', gap: 10, padding: '8px 0', borderBottom: i < timeline.length - 1 ? '1px dashed var(--rule-2)' : 'none', alignItems: 'flex-start' }}>
                      <span className="t-xs" style={{ fontWeight: 800, color: 'var(--accent)', position: 'relative' }}>
                        <span style={{ position: 'absolute', left: -16, top: 4, width: 9, height: 9, borderRadius: '50%', background: seatColor, border: '2px solid var(--bg)' }} />
                        T{entry.turn}
                      </span>
                      <div>
                        <div className="t-md" style={{ fontWeight: 600, lineHeight: 1.3 }}>{entry.action}</div>
                        <div className="t-xs muted" style={{ marginTop: 2 }}>SEAT.{String(entry.seat + 1).padStart(2, '0')} · {cmdr.split(',')[0].toUpperCase()}</div>
                      </div>
                    </div>
                  )
                })}
                <div className="t-xs muted-2" style={{ marginTop: 10, lineHeight: 1.5 }}>
                  GENERATED FROM RESULT METADATA. REAL TIMELINE NEEDS COMPLETEDGAME.LOG[] RETENTION SERVER-SIDE.
                </div>
              </div>
            </Panel>
          )
        })()}

        {/* Per-seat performance — final stats real, life curve mock. */}
        {featuredGame && (
          <Panel code="RPT.F" title="PER-SEAT PERFORMANCE" style={{ gridColumn: '1 / -1' }}
            right={<MockTag />}>
            <div className="grid col-2 gap-4">
              {seats.map((s, i) => {
                const isWinner = i === winnerIdx
                const cmdr = commanders[i] || 'UNKNOWN'
                const perms = s.battlefield || []
                const lands = perms.filter(p => p.is_land).length
                const nonLand = perms.length - lands
                const lifePct = Math.max(0, Math.min(100, (s.life / 40) * 100))
                const curve = mockLifeCurve(featuredGame.turns, s.life, isWinner, s.lost)
                const accent = isWinner ? 'var(--ok)' : s.lost ? 'var(--danger)' : 'var(--ink-2)'
                return (
                  <div key={i} className="panel" style={{ padding: 0, borderColor: accent }}>
                    <div className="panel-hd">
                      <span>{cmdr.split(',')[0].toUpperCase()}</span>
                      <span className="t-xs">SEAT.{String(i + 1).padStart(2, '0')}</span>
                    </div>
                    <div style={{ padding: 12 }}>
                      <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 10 }}>
                        <div style={{ flex: 1 }}>
                          <div className="t-xs muted" style={{ marginBottom: 2 }}>LIFE</div>
                          <Bar value={lifePct} />
                          <div className="t-xs" style={{ marginTop: 3, fontWeight: 700, color: accent }}>{s.life} / 40</div>
                        </div>
                        <div>
                          <div className="t-xs muted" style={{ marginBottom: 2 }}>LIFE OVER TIME</div>
                          <Sparkline points={curve} accent={accent} />
                        </div>
                      </div>
                      <KV rows={[
                        ['BATTLEFIELD', String(perms.length)],
                        ['NON-LAND', String(nonLand)],
                        ['LANDS', String(lands)],
                        ['HAND', String(s.hand_size)],
                        ['LIBRARY', String(s.library_size)],
                        ['GRAVEYARD', String(s.gy_size)],
                      ]} />
                    </div>
                  </div>
                )
              })}
            </div>
            <div className="t-xs muted-2" style={{ marginTop: 10, lineHeight: 1.5 }}>
              FINAL STATS ARE REAL. LIFE-OVER-TIME SPARKLINES ARE INTERPOLATED ENDPOINTS — REAL CURVES NEED COMPLETEDGAME.LIFE_HISTORY[] FROM THE ENGINE.
            </div>
          </Panel>
        )}
      </div>
    </>
  )
}
