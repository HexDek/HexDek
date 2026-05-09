import { useEffect, useState, useCallback } from 'react'
import { Panel, KV, Tag } from './chrome'
import { api } from '../services/api'

// CreditsPanel — displays the operator's current balance, free-tier
// quota state, and recent transaction history. Hosts the
// "insufficient credits" UX messaging. Designed to drop into the
// operator profile but reusable wherever a credit summary makes
// sense (gauntlet pre-flight modal, future spending pages).
//
// Props:
//   compact   — render the body without the history list (sidebar use)
//   refreshKey — bump this to force a re-fetch (e.g. after the user
//                runs a gauntlet; the parent screen flips a counter)
//
// Failure modes:
//   - Unauthenticated → renders "Sign in to view credits"
//   - API error → renders the error message inline; balance reads as —
//
// The component does no spending of its own — the spend flow lives
// alongside the gauntlet button in DeckArchive. This panel is read-
// mostly so the user can audit their balance and recent activity.

const REASON_LABELS = {
  compute_contribution: 'COMPUTE EARNED',
  gauntlet_run:         'GAUNTLET RUN',
  extended_analysis:    'EXTENDED ANALYSIS',
  admin_adjustment:     'ADMIN ADJUSTMENT',
}

function formatReason(reason) {
  return REASON_LABELS[reason] || (reason || 'OTHER').toUpperCase().replace(/_/g, ' ')
}

function formatTimestamp(ts) {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  if (isNaN(d.getTime())) return '—'
  return d.toISOString().slice(0, 16).replace('T', ' ')
}

export default function CreditsPanel({ compact = false, refreshKey = 0 }) {
  const [balance, setBalance] = useState(null)
  const [quota, setQuota] = useState(null)
  const [history, setHistory] = useState([])
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [b, q] = await Promise.all([
        api.getCreditBalance(),
        api.getCreditQuota(),
      ])
      setBalance(b?.credits ?? 0)
      setQuota(q)
      if (!compact) {
        const h = await api.getCreditHistory(20)
        setHistory(h?.transactions || [])
      }
    } catch (err) {
      if (err?.status === 401) {
        setError('Sign in to view credits')
      } else {
        setError(err?.message || 'Failed to load credits')
      }
    } finally {
      setLoading(false)
    }
  }, [compact])

  useEffect(() => { load() }, [load, refreshKey])

  return (
    <Panel
      code="CR.0"
      title="CREDITS"
      right={
        balance != null && quota
          ? <Tag solid kind={quota.can_run_paid || quota.can_run_free ? 'ok' : 'warn'}>
              {balance} CR
            </Tag>
          : null
      }
    >
      {loading && balance == null ? (
        <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
          &gt; LOADING<span className="blink">_</span>
        </div>
      ) : error ? (
        <div className="t-md muted" style={{ textAlign: 'center', padding: 14 }}>
          &gt; {error.toUpperCase()}
        </div>
      ) : (
        <>
          <KV rows={[
            ['BALANCE', `${balance} CR`, 'credit_balance'],
            ['FREE GAUNTLETS TODAY', quota
              ? `${Math.max(0, quota.free_limit - quota.free_remaining)} / ${quota.free_limit}`
              : '—',
              'free_tier'],
            ['COST PER PAID RUN', quota ? `${quota.cost_per_run} CR` : '—'],
            ['STATUS',
              quota?.can_run_free
                ? <span style={{ color: 'var(--ok)' }}>FREE TIER ACTIVE</span>
                : quota?.can_run_paid
                  ? <span style={{ color: 'var(--accent)' }}>CREDITS REQUIRED</span>
                  : <span style={{ color: 'var(--danger)' }}>EXHAUSTED — EARN OR WAIT</span>
            ],
          ]} />

          {!compact && (
            <>
              <div className="hr" style={{ margin: '12px 0' }} />
              <div className="t-xs muted" style={{ marginBottom: 6, letterSpacing: '0.08em' }}>
                RECENT TRANSACTIONS
              </div>
              {history.length === 0 ? (
                <div className="t-md muted" style={{ textAlign: 'center', padding: 12 }}>
                  &gt; NO TRANSACTIONS YET. CONTRIBUTE COMPUTE TO EARN.
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                  {history.map((t) => {
                    const positive = t.amount > 0
                    return (
                      <div
                        key={t.id}
                        style={{
                          display: 'grid',
                          gridTemplateColumns: '1fr auto auto',
                          alignItems: 'center',
                          gap: 8,
                          padding: '4px 6px',
                          borderTop: '1px solid var(--rule-2)',
                          fontSize: 10,
                        }}
                      >
                        <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          <span style={{ color: 'var(--ink)', fontWeight: 600 }}>
                            {formatReason(t.reason)}
                          </span>
                          {t.reference && (
                            <span className="muted" style={{ marginLeft: 6 }}>
                              {t.reference}
                            </span>
                          )}
                        </span>
                        <span style={{
                          color: positive ? 'var(--ok)' : 'var(--danger)',
                          fontWeight: 700,
                          minWidth: 50,
                          textAlign: 'right',
                        }}>
                          {positive ? '+' : ''}{t.amount}
                        </span>
                        <span className="muted-2" style={{ minWidth: 110, textAlign: 'right' }}>
                          {formatTimestamp(t.created_at)}
                        </span>
                      </div>
                    )
                  })}
                </div>
              )}
            </>
          )}
        </>
      )}
    </Panel>
  )
}
