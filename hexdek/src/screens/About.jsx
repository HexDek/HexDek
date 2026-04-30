import { Panel, KV } from '../components/chrome'

export default function About() {
  return (
    <div style={{ padding: '20px 30px', maxWidth: 800, margin: '0 auto' }}>
      <Panel code="ABT.0" title="ABOUT HEXDEK">
        <div className="t-md" style={{ lineHeight: 1.7 }}>
          <p>
            HexDek is an open-source Magic: The Gathering Commander engine and analytics platform.
            It simulates full 4-player Commander games with AI-driven decision making, tracks deck
            performance via ELO and TrueSkill ratings, and surfaces deep analytics on card synergies,
            matchups, and metagame trends.
          </p>
          <p style={{ marginTop: 12 }}>
            The engine parses real MTG card text into an AST, implements the comprehensive rules
            (priority, stack, combat, state-based actions, replacement effects, layers), and plays
            complete games autonomously. Every game produces rich data that feeds back into the system.
          </p>
        </div>
      </Panel>

      <Panel code="ABT.1" title="PHILOSOPHY" style={{ marginTop: 16 }}>
        <div className="t-md" style={{ lineHeight: 1.7 }}>
          <p>
            <strong>No ads. No subscriptions. No paywalls.</strong>
          </p>
          <p style={{ marginTop: 8 }}>
            HexDek is built by players, for players. The project runs entirely on donations and
            community contributions. We believe competitive MTG analytics should be accessible to
            everyone — not locked behind a paywall.
          </p>
          <p style={{ marginTop: 8 }}>
            All engine code is open source under the MIT license. Run your own tournaments,
            contribute card handlers, or just enjoy the data.
          </p>
        </div>
      </Panel>

      <Panel code="ABT.2" title="WHAT WE DO" style={{ marginTop: 16 }}>
        <KV rows={[
          ['ENGINE', 'Full MTG rules engine — priority, stack, combat, SBAs, layers, zones'],
          ['AI', 'YggdrasilHat — 8-dimensional evaluator with political awareness'],
          ['ANALYTICS', 'Heimdall post-game reports, Freya strategy analysis, Huginn emergent synergies'],
          ['RATINGS', 'ELO + TrueSkill Bayesian ratings across thousands of simulated games'],
          ['INTEGRITY', 'Muninn persistent telemetry — parser gaps, crashes, dead triggers tracked'],
          ['DECKS', 'Import from Moxfield or paste a list — auto-parsed, auto-rated'],
        ]} />
      </Panel>

      <Panel code="ABT.3" title="THE STACK" style={{ marginTop: 16 }}>
        <KV rows={[
          ['ENGINE', 'Go (100K+ lines)'],
          ['FRONTEND', 'React + Vite'],
          ['AI HATS', 'Greedy, Poker, YggdrasilHat (unified political AI)'],
          ['TOOLS', 'Thor, Odin, Loki, Heimdall, Freya, Valkyrie, Huginn, Muninn'],
          ['DATA', 'JSON flat-file persistence (no external DB dependency)'],
          ['LICENSE', 'MIT'],
        ]} />
      </Panel>

      <Panel code="ABT.4" title="CONTACT" style={{ marginTop: 16 }}>
        <div className="t-md" style={{ lineHeight: 1.7 }}>
          <p>
            Found a bug or have a suggestion? Use the <strong style={{ color: 'var(--danger)' }}>Bug Report</strong> button
            in the footer below.
          </p>
          <p style={{ marginTop: 8 }}>
            Want to contribute? The project is on GitHub. Pull requests welcome.
          </p>
        </div>
      </Panel>
    </div>
  )
}
