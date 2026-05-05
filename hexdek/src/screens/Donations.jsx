import { useState, useEffect } from 'react'
import { Panel, KV, Bar, Btn, Tape, Tag } from '../components/chrome'
import { api } from '../services/api'

const RECURRING = [
  ['CLAUDE AI', 200],
  ['DOMAIN', 2],
]

const SUNK = [
  ['DARKSTAR HARDWARE', 2000],
  ['CLAUDE (3 MO)', 600],
  ['DOMAIN REG.', 20],
]

const MONTHLY_GOAL = RECURRING.reduce((s, r) => s + r[2], 0)
const TOTAL_INVESTED = SUNK.reduce((s, r) => s + r[2], 0)

export default function Donations() {
  const [summary, setSummary] = useState({ month_total: 0, all_time_total: 0, month_goal: MONTHLY_GOAL, recent: [] })

  useEffect(() => {
    api.getDonationsSummary().then(setSummary).catch(() => {})
  }, [])

  const raised = summary.month_total
  const goal = summary.month_goal || MONTHLY_GOAL
  const pct = Math.min(100, Math.round((raised / goal) * 100))

  return (
    <div style={{ padding: '20px 30px', maxWidth: 800, margin: '0 auto' }}>
      <Panel code="DON.0" title="DONATIONS // KEEP HEXDEK ALIVE">
        <div className="t-md" style={{ lineHeight: 1.7 }}>
          <p>
            HexDek runs on dedicated hardware and AI-powered development. There are no forced ads,
            no subscriptions, and no data sales. Every dollar goes directly to keeping the
            engine running and the simulation queue chewing through games.
          </p>
        </div>
      </Panel>

      <Panel code="DON.1" title="MONTHLY OPERATING COSTS" style={{ marginTop: 16 }}>
        <div style={{ padding: '4px 0 12px' }}>
          <Tape
            left={`$${raised.toFixed(0)} RAISED`}
            mid={`${pct}%`}
            right={`$${goal}/MO GOAL`}
          />
          <div style={{ marginTop: 12 }}>
            <Bar value={raised} max={goal} lg />
          </div>
        </div>
        <KV rows={RECURRING.map(([k, c]) => [k, `$${c}/mo`])} />
        <div style={{ marginTop: 12, paddingTop: 10, borderTop: '1px dashed var(--rule-2)' }}>
          <KV rows={[['MONTHLY TOTAL', `$${MONTHLY_GOAL}/mo`]]} />
        </div>
        <div className="t-xs muted" style={{ marginTop: 10 }}>
          Surplus rolls into infrastructure upgrades and tournament prize pools.
        </div>
      </Panel>

      {summary.recent.length > 0 && (
        <Panel code="DON.R" title="RECENT SUPPORTERS" style={{ marginTop: 16 }}>
          {summary.recent.map((d, i) => (
            <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', borderBottom: i < summary.recent.length - 1 ? '1px dashed var(--rule-2)' : 'none' }}>
              <span className="t-md" style={{ fontWeight: 700 }}>{d.from_name}</span>
              <span className="t-md" style={{ color: 'var(--ok)' }}>${d.amount}</span>
            </div>
          ))}
        </Panel>
      )}

      <Panel code="DON.2" title="ALREADY INVESTED" style={{ marginTop: 16 }}>
        <KV rows={SUNK.map(([k, c]) => [k, `$${c.toLocaleString()}`])} />
        <div style={{ marginTop: 12, paddingTop: 10, borderTop: '1px dashed var(--rule-2)' }}>
          <KV rows={[['TOTAL TO DATE', `$${(TOTAL_INVESTED + summary.all_time_total).toLocaleString()}`]]} />
        </div>
        <div className="t-xs muted" style={{ marginTop: 10 }}>
          DARKSTAR is committed hardware — amortized, not a recurring cost.
          {summary.all_time_total > 0 && ` Community has contributed $${summary.all_time_total.toFixed(0)} total.`}
        </div>
      </Panel>

      <Panel code="DON.3" title="WAYS TO SUPPORT" style={{ marginTop: 16 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div className="t-md" style={{ lineHeight: 1.6 }}>
            Three ways to keep HexDek running. Pick what works for you.
          </div>

          <div style={{ display: 'flex', gap: 14, flexWrap: 'wrap' }}>
            <div className="panel" style={{ flex: 1, minWidth: 180, padding: '14px 16px' }}>
              <Tag solid>DONATE</Tag>
              <div className="t-md" style={{ marginTop: 8, lineHeight: 1.5 }}>
                One-time tips or recurring sponsorships. Goes directly to infrastructure and AI costs.
              </div>
              <div style={{ display: 'flex', gap: 8, marginTop: 12, flexWrap: 'wrap' }}>
                <a href="https://ko-fi.com/hexdek" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
                  <Btn solid>KO-FI</Btn>
                </a>
              </div>
            </div>

            <div className="panel" style={{ flex: 1, minWidth: 220, padding: '14px 16px', borderColor: 'var(--ink)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
                <Tag solid>BOINC // COMPUTE</Tag>
                <Tag kind="warn" solid style={{ fontSize: 8, padding: '1px 5px' }}>COMING SOON</Tag>
              </div>
              <div className="t-md" style={{ marginTop: 10, lineHeight: 1.55 }}>
                Contribute your CPU cycles to run MTG simulations. Earn credits toward your
                own deck testing.
              </div>
              <ul style={{
                marginTop: 10, marginBottom: 0, paddingLeft: 0, listStyle: 'none',
                fontSize: 10, letterSpacing: '0.04em', lineHeight: 1.7, color: 'var(--ink-2)',
              }}>
                <li>&gt; RUN GAMES IN BACKGROUND VIA BOINC CLIENT</li>
                <li>&gt; CREDITS BUY EXTENDED FREYA ANALYSIS + GAUNTLET QUEUES</li>
                <li>&gt; OPT-OUT ANYTIME — YOUR HARDWARE, YOUR RULES</li>
              </ul>
              <div style={{ marginTop: 12, paddingTop: 10, borderTop: '1px dashed var(--rule-2)' }}>
                <span className="t-xs muted">CLIENT IN DEVELOPMENT — NO ACTION REQUIRED</span>
              </div>
            </div>

            <div className="panel" style={{ flex: 1, minWidth: 220, padding: '14px 16px', borderColor: 'var(--ink)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
                <Tag solid kind="ok">SUPPORT DEV</Tag>
                <span className="t-xs" style={{ color: 'var(--ok)', letterSpacing: '0.1em' }}>● LIVE</span>
              </div>
              <div className="t-md" style={{ marginTop: 10, lineHeight: 1.55 }}>
                Sponsor ongoing development. Funds buy AI dev hours and DARKSTAR upgrades —
                every contribution maps to a public roadmap line.
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 12 }}>
                <a href="https://ko-fi.com/hexdek" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
                  <Btn solid arrow="↗" style={{ width: '100%', justifyContent: 'space-between' }}>KO-FI · ONE-TIME</Btn>
                </a>
                <a href="https://github.com/sponsors/hexdek-labs" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
                  <Btn arrow="↗" style={{ width: '100%', justifyContent: 'space-between' }}>GH SPONSORS · MONTHLY</Btn>
                </a>
              </div>
            </div>
          </div>

          <div className="t-xs muted-2">
            Research tokens unlock extended deck analysis, priority simulation queue, and custom tournament runs.
          </div>
        </div>
      </Panel>

      <Panel code="DON.4" title="PHILOSOPHY" style={{ marginTop: 16 }}>
        <div className="t-md" style={{ lineHeight: 1.7 }}>
          <p><strong>No forced ads. No paywalls. No data sales.</strong></p>
          <p style={{ marginTop: 8 }}>
            HexDek is community-powered. Engine code is MIT-licensed and the data is yours.
            Power users who contribute — whether through compute, donations, or community engagement —
            get rewarded with research tokens and priority access. But the core experience is always free.
          </p>
        </div>
      </Panel>

      <Panel code="DON.5" title="FAQ" style={{ marginTop: 16 }}>
        <FAQItem
          q="WHERE DOES MY MONEY GO?"
          a={
            <>
              Straight to the operating budget at the top of this page. The recurring line is
              dominated by the <strong>Claude AI subscription</strong> that drives engine
              development and the <strong>domain registration</strong>. Surplus is amortized
              against the one-time DARKSTAR hardware investment and stored toward future
              tournament prize pools. Every dollar is itemized — the line items above ARE the
              budget.
            </>
          }
        />
        <FAQItem
          q="IS HEXDEK OPEN SOURCE?"
          a={
            <>
              Yes — <strong>MIT licensed</strong>. The Go engine, Freya analyzer, hat AI,
              and React frontend are all in the public repo. Fork it, run it locally, ship a
              competing fishtank. The deal is simple: code is free, the donations keep the
              public instance running.{' '}
              <a
                href="https://github.com/hexdek-labs/HexDek"
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: 'var(--ok)', textDecoration: 'none', borderBottom: '1px dotted var(--ok)' }}
              >
                github.com/hexdek-labs/HexDek ↗
              </a>
            </>
          }
        />
        <FAQItem
          q="DO I NEED TO PAY TO USE HEXDEK?"
          a={
            <>
              No. Deck import, Freya analysis, gauntlet results, the live fishtank, and the
              full spectator feed are free for everyone forever. Contributions unlock
              priority queue and extended analysis for power users — they don't gate any
              core feature.
            </>
          }
        />
        <FAQItem
          q="CAN I GET A REFUND?"
          a={
            <>
              Donations are non-refundable by Ko-fi's terms, but reach out via{' '}
              <a
                href="https://discord.gg/Mz2ueRFXds"
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: 'var(--ok)', textDecoration: 'none', borderBottom: '1px dotted var(--ok)' }}
              >
                Discord ↗
              </a>{' '}
              if anything's gone wrong and we'll make it right.
            </>
          }
          last
        />
      </Panel>
    </div>
  )
}

function FAQItem({ q, a, last }) {
  return (
    <div style={{
      padding: '12px 0',
      borderBottom: last ? 'none' : '1px dashed var(--rule-2)',
    }}>
      <div className="t-sm" style={{
        color: 'var(--ink)',
        fontWeight: 700,
        letterSpacing: '0.06em',
        marginBottom: 6,
      }}>
        &gt; {q}
      </div>
      <div className="t-md" style={{ lineHeight: 1.6, color: 'var(--ink-2)' }}>
        {a}
      </div>
    </div>
  )
}
