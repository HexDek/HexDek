package hat

// GamePlan represents the hat's current strategic intent.
type GamePlan int

const (
	PlanDevelop  GamePlan = iota // default: play lands, cast spells, build board
	PlanAssemble                 // combo pieces tracked, prioritize tutors + protection
	PlanExecute                  // combo ready, go for the win NOW
	PlanDisrupt                  // opponent threatening win, hold interaction
	PlanPivot                    // primary plan failed, switch to plan B
	PlanDefend                   // archenemy, survive first, win later
)

func (p GamePlan) String() string {
	switch p {
	case PlanDevelop:
		return "Develop"
	case PlanAssemble:
		return "Assemble"
	case PlanExecute:
		return "Execute"
	case PlanDisrupt:
		return "Disrupt"
	case PlanPivot:
		return "Pivot"
	case PlanDefend:
		return "Defend"
	default:
		return "Unknown"
	}
}

// PlanState tracks the hat's strategic awareness for plan transitions.
type PlanState struct {
	Current        GamePlan
	ComboReady     int     // count of combo pieces in hand+battlefield
	ComboTotal     int     // total pieces needed (from Freya)
	ThreatLevel    float64 // highest opponent threat score
	TurnsSincePlan int     // turns in current plan (for timeout transitions)
}

// Evaluate updates the plan state based on the current game situation and
// transitions between plans when conditions change. Called once per turn
// (on upkeep) to keep the state machine in sync with the board.
func (ps *PlanState) Evaluate(combo *ComboAssessment, threatLevel float64) {
	ps.TurnsSincePlan++

	// Update combo tracking from assessment.
	if combo != nil {
		ps.ComboReady = combo.PiecesFound
		ps.ComboTotal = combo.PiecesTotal
	}
	ps.ThreatLevel = threatLevel

	// Transition logic — priority order matters.
	switch {
	case combo != nil && combo.Executable:
		// Combo is live — execute now.
		if ps.Current != PlanExecute {
			ps.Current = PlanExecute
			ps.TurnsSincePlan = 0
		}

	case combo != nil && combo.Assembling:
		// Within one piece + tutor — assemble.
		if ps.Current != PlanAssemble && ps.Current != PlanExecute {
			ps.Current = PlanAssemble
			ps.TurnsSincePlan = 0
		}

	case threatLevel > 0.7:
		// High external threat — disrupt.
		if ps.Current != PlanDisrupt && ps.Current != PlanExecute {
			ps.Current = PlanDisrupt
			ps.TurnsSincePlan = 0
		}

	case ps.Current == PlanExecute && (combo == nil || !combo.Executable):
		// Combo fizzled — fall back to develop.
		ps.Current = PlanDevelop
		ps.TurnsSincePlan = 0

	case ps.Current == PlanAssemble && ps.TurnsSincePlan > 5:
		// Been assembling too long — pivot.
		ps.Current = PlanPivot
		ps.TurnsSincePlan = 0

	case (ps.Current == PlanPivot || ps.Current == PlanDefend || ps.Current == PlanDisrupt) && ps.TurnsSincePlan > 3:
		// Reactive plans timeout — return to develop.
		ps.Current = PlanDevelop
		ps.TurnsSincePlan = 0
	}
}

// PlanWeightMultipliers returns per-dimension weight multipliers for the
// current plan. Values > 1 boost, < 1 suppress. The caller applies these
// as temporary multipliers to EvalWeights before scoring and restores
// after. This keeps the plan influence reversible per-decision.
func (ps *PlanState) PlanWeightMultipliers() EvalWeights {
	// Neutral: all 1.0 (no change).
	m := EvalWeights{
		BoardPresence:           1.0,
		CardAdvantage:           1.0,
		ManaAdvantage:           1.0,
		LifeResource:            1.0,
		ComboProximity:          1.0,
		ThreatExposure:          1.0,
		CommanderProgress:       1.0,
		GraveyardValue:          1.0,
		DrainEngine:             1.0,
		ArtifactSynergy:         1.0,
		EnchantmentSynergy:      1.0,
		OpponentGraveyardThreat: 1.0,
		PartnerSynergy:          1.0,
		ActivationTempo:         1.0,
		ToolboxBreadth:          1.0,
		ThreatTrajectory:        1.0,
	}

	switch ps.Current {
	case PlanExecute:
		// All-in on combo. Heavy boost to ComboProximity, suppress
		// everything else — we are going for the win THIS turn.
		m.ComboProximity = 2.5
		m.BoardPresence = 0.3
		m.CardAdvantage = 0.4
		m.ManaAdvantage = 0.5
		m.LifeResource = 0.3
		m.ThreatExposure = 0.3

	case PlanAssemble:
		// Draw + tutors to find pieces. Boost card advantage and toolbox.
		m.CardAdvantage = 1.6
		m.ToolboxBreadth = 1.5
		m.ComboProximity = 1.4
		m.BoardPresence = 0.7

	case PlanDisrupt:
		// Hold interaction, prioritize threat assessment.
		m.ThreatExposure = 1.8
		m.ThreatTrajectory = 1.5
		m.CardAdvantage = 1.2
		m.BoardPresence = 0.6
		m.ComboProximity = 0.5

	case PlanDefend:
		// Survival mode — life and board presence matter most.
		m.LifeResource = 1.8
		m.BoardPresence = 1.4
		m.ThreatExposure = 1.3
		m.ComboProximity = 0.4

	case PlanPivot:
		// Primary plan failed — switch to beatdown.
		m.BoardPresence = 1.6
		m.CommanderProgress = 1.4
		m.ComboProximity = 0.3
		m.ToolboxBreadth = 0.7

	case PlanDevelop:
		// Default — no adjustment, use archetype weights as-is.
	}

	return m
}
