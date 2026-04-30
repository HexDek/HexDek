import { Panel, KV, Bar, Btn, Tape } from '../components/chrome'

const COSTS = [
  ['DARKSTAR HOSTING', 'Engine box (Ryzen 9, Ubuntu)', 35],
  ['MISTY HOSTING', 'Frontend / Caddy', 15],
  ['DOMAIN', 'hexdek.bluefroganalytics.com', 2],
  ['BANDWIDTH', 'Egress + replays', 8],
  ['BACKUPS', 'Off-site snapshots', 5],
]

const MONTHLY_GOAL = COSTS.reduce((s, r) => s + r[2], 0)
const CURRENT_DONATIONS = 0

export default function Donations() {
  const pct = Math.min(100, Math.round((CURRENT_DONATIONS / MONTHLY_GOAL) * 100))

  return (
    <div style={{ padding: '20px 30px', maxWidth: 800, margin: '0 auto' }}>
      <Panel code="DON.0" title="DONATIONS // KEEP HEXDEK ALIVE">
        <div className="t-md" style={{ lineHeight: 1.7 }}>
          <p>
            HexDek runs on a small home-lab and a couple of cloud services. There are no ads,
            no subscriptions, and no data sales. Every dollar goes directly to keeping the
            engine running, replays archived, and the simulation queue chewing through games.
          </p>
        </div>
      </Panel>

      <Panel code="DON.1" title="MONTHLY GOAL" style={{ marginTop: 16 }}>
        <div style={{ padding: '4px 0 12px' }}>
          <Tape
            left={`$${CURRENT_DONATIONS} RAISED`}
            mid={`${pct}%`}
            right={`$${MONTHLY_GOAL} GOAL`}
          />
          <div style={{ marginTop: 12 }}>
            <Bar value={CURRENT_DONATIONS} max={MONTHLY_GOAL} lg />
          </div>
          <div className="t-xs muted" style={{ marginTop: 10 }}>
            Resets monthly. Surplus rolls into a card-data buffer for Scryfall mirroring and
            tournament prize pools.
          </div>
        </div>
      </Panel>

      <Panel code="DON.2" title="WHERE THE MONEY GOES" style={{ marginTop: 16 }}>
        <KV rows={COSTS.map(([k, v, c]) => [k, `${v} — $${c}/mo`])} />
        <div style={{ marginTop: 12, paddingTop: 10, borderTop: '1px dashed var(--rule-2)' }}>
          <KV rows={[['TOTAL', `$${MONTHLY_GOAL}/mo`]]} />
        </div>
      </Panel>

      <Panel code="DON.3" title="SUPPORT HEXDEK" style={{ marginTop: 16 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div className="t-md" style={{ lineHeight: 1.6 }}>
            One-time tips and recurring sponsorships both work. Pick whichever channel you
            already have an account on.
          </div>
          <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
            <a href="https://ko-fi.com/" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
              <Btn solid>KO-FI / TIP</Btn>
            </a>
            <a href="https://github.com/sponsors" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
              <Btn>GITHUB SPONSORS</Btn>
            </a>
            <a href="https://www.patreon.com/" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
              <Btn ghost>PATREON</Btn>
            </a>
          </div>
          <div className="t-xs muted-2" style={{ marginTop: 4 }}>
            Donation links are placeholders for now — the real accounts ship with the next build.
          </div>
        </div>
      </Panel>

      <Panel code="DON.4" title="PHILOSOPHY" style={{ marginTop: 16 }}>
        <div className="t-md" style={{ lineHeight: 1.7 }}>
          <p><strong>No ads. No paywalls. No tracking.</strong></p>
          <p style={{ marginTop: 8 }}>
            HexDek is community-powered. Engine code is MIT-licensed and the data is yours.
            If you cannot donate, that is fine — contribute card handlers, file bug reports,
            or just play more games. The more replays we ingest, the better the AI gets.
          </p>
        </div>
      </Panel>
    </div>
  )
}
