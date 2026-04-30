package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// 1. Mana base grading — taplands, fetches, utility lands, overall grade.
// ---------------------------------------------------------------------------

func computeManaBaseGrade(dp *DeckProfile, report *FreyaReport, oracle *oracleDB) {
	if oracle == nil {
		return
	}

	score := 100 // start at A+, deduct for issues

	for _, p := range report.Profiles {
		if !p.IsLand {
			continue
		}

		entry := oracle.lookup(p.Name)
		if entry == nil {
			continue
		}
		ot := strings.ToLower(entry.OracleText)
		if ot == "" && len(entry.CardFaces) > 0 {
			ot = strings.ToLower(entry.CardFaces[0].OracleText)
		}
		tl := strings.ToLower(p.TypeLine)

		if containsAny(ot, "enters tapped", "enters the battlefield tapped") &&
			!strings.Contains(ot, "you may pay") &&
			!strings.Contains(ot, "unless") {
			dp.TaplandCount++
		}

		if strings.Contains(ot, "search your library") && strings.Contains(tl, "land") &&
			!strings.Contains(tl, "creature") {
			dp.FetchCount++
		}

		if len(p.LandColors) == 0 || (len(p.LandColors) == 1 && p.LandColors[0] == "C") {
			if !containsAny(strings.ToLower(p.Name), "plains", "island", "swamp", "mountain", "forest", "wastes") {
				dp.UtilityLandCount++
			}
		}
	}

	// Deductions
	if dp.TaplandCount > 5 {
		penalty := (dp.TaplandCount - 5) * 5
		score -= penalty
		dp.ManaBaseNotes = append(dp.ManaBaseNotes,
			fmt.Sprintf("%d taplands slowing you down — consider upgrading %d to untapped duals",
				dp.TaplandCount, dp.TaplandCount-5))
	} else if dp.TaplandCount > 3 {
		score -= 5
	}

	if dp.UtilityLandCount > 5 {
		penalty := (dp.UtilityLandCount - 5) * 3
		score -= penalty
		dp.ManaBaseNotes = append(dp.ManaBaseNotes,
			fmt.Sprintf("%d utility lands eating color slots — may cause color screw", dp.UtilityLandCount))
	}

	if dp.FetchCount >= 5 {
		score += 10
		dp.ManaBaseNotes = append(dp.ManaBaseNotes,
			fmt.Sprintf("%d fetchlands providing excellent color fixing", dp.FetchCount))
	}

	if report.Stats != nil && len(report.Stats.ColorGaps) > 0 {
		score -= len(report.Stats.ColorGaps) * 10
	}

	switch {
	case score >= 90:
		dp.ManaBaseGrade = "A"
	case score >= 75:
		dp.ManaBaseGrade = "B"
	case score >= 60:
		dp.ManaBaseGrade = "C"
	case score >= 45:
		dp.ManaBaseGrade = "D"
	default:
		dp.ManaBaseGrade = "F"
	}
}

// ---------------------------------------------------------------------------
// 2. Threat assessment — what specific hosers does this deck fear?
// ---------------------------------------------------------------------------

type hoserMapping struct {
	Condition string
	Hoser     string
	Reason    string
}

var hoserDB = []hoserMapping{
	{"graveyard_heavy", "Rest in Peace", "exiles your graveyard engine"},
	{"graveyard_heavy", "Leyline of the Void", "prevents your graveyard from filling"},
	{"graveyard_heavy", "Dauthi Voidwalker", "exiles your dying creatures"},
	{"artifact_heavy", "Collector Ouphe", "shuts down your artifact mana and combo pieces"},
	{"artifact_heavy", "Stony Silence", "locks your artifact activations"},
	{"artifact_heavy", "Vandalblast", "one-sided artifact wipe"},
	{"creature_heavy", "Cyclonic Rift", "bounces your entire board"},
	{"creature_heavy", "Toxic Deluge", "uncounterable creature wipe"},
	{"creature_heavy", "Elesh Norn, Grand Cenobite", "shrinks your board while pumping theirs"},
	{"combo_heavy", "Rule of Law", "locks you to one spell per turn"},
	{"combo_heavy", "Drannith Magistrate", "prevents casting from non-hand zones"},
	{"combo_heavy", "Stifle", "counters your critical triggers"},
	{"token_heavy", "Massacre Wurm", "kills tokens and drains you"},
	{"token_heavy", "Rakdos Charm", "each creature deals 1 to you"},
	{"enchantment_heavy", "Aura Shards", "destroys enchantments on creature ETB"},
	{"enchantment_heavy", "Back to Nature", "instant-speed enchantment wipe"},
	{"tutor_heavy", "Opposition Agent", "steals your tutored cards"},
	{"tutor_heavy", "Aven Mindcensor", "limits your searches to top 4"},
	{"lifegain", "Erebos, God of the Dead", "prevents your lifegain"},
	{"lifegain", "Sulfuric Vortex", "prevents lifegain and pressures life total"},
	{"etb_heavy", "Torpor Orb", "shuts down all your ETB triggers"},
	{"etb_heavy", "Hushbringer", "prevents ETB and death triggers"},
	{"spellslinger", "Deafening Silence", "limits noncreature spells to one per turn"},
	{"spellslinger", "Thalia, Guardian of Thraben", "taxes your noncreature spells"},
	{"land_ramp", "Blood Moon", "turns your nonbasics into Mountains"},
	{"land_ramp", "Back to Basics", "taps your nonbasics"},
}

func computeThreatAssessment(dp *DeckProfile, report *FreyaReport) {
	conditions := map[string]bool{}

	if report.Roles != nil {
		rolePct := func(r RoleTag) float64 {
			if report.Roles.TotalCards == 0 {
				return 0
			}
			return float64(report.Roles.RoleCounts[r]) / float64(report.Roles.TotalCards)
		}
		if rolePct(RoleCombo) >= 0.08 || len(report.TrueInfinites)+len(report.Determined) >= 3 {
			conditions["combo_heavy"] = true
		}
	}

	graveyardCards := 0
	artifactCards := 0
	creatureCards := 0
	tokenCards := 0
	enchantmentCards := 0
	etbCards := 0
	rampLandCards := 0
	for _, p := range report.Profiles {
		if p.IsLand {
			continue
		}
		tl := strings.ToLower(p.TypeLine)
		if p.IsRecursion || containsAny(strings.Join(p.Effects, ","), "self_mill", "mass_reanimate") {
			graveyardCards++
		}
		if strings.Contains(tl, "artifact") {
			artifactCards++
		}
		if strings.Contains(tl, "creature") {
			creatureCards++
		}
		if strings.Contains(tl, "enchantment") {
			enchantmentCards++
		}
		if p.HasValueETB || profileHasTrigger(p, "etb") {
			etbCards++
		}
		for _, e := range p.Effects {
			if e == "land_fetch" {
				rampLandCards++
				break
			}
		}
		for _, r := range p.Produces {
			if r == ResToken {
				tokenCards++
				break
			}
		}
	}

	if graveyardCards >= 8 {
		conditions["graveyard_heavy"] = true
	}
	if artifactCards >= 12 {
		conditions["artifact_heavy"] = true
	}
	if creatureCards >= 25 {
		conditions["creature_heavy"] = true
	}
	if tokenCards >= 8 {
		conditions["token_heavy"] = true
	}
	if enchantmentCards >= 10 {
		conditions["enchantment_heavy"] = true
	}
	if etbCards >= 10 {
		conditions["etb_heavy"] = true
	}
	if rampLandCards >= 6 {
		conditions["land_ramp"] = true
	}

	if report.TutorCount >= 6 {
		conditions["tutor_heavy"] = true
	}

	arch := strings.ToLower(dp.PrimaryArchetype)
	if containsAny(arch, "lifegain") {
		conditions["lifegain"] = true
	}
	if containsAny(arch, "spellslinger", "storm") {
		conditions["spellslinger"] = true
	}

	seen := map[string]bool{}
	for _, h := range hoserDB {
		if conditions[h.Condition] && !seen[h.Hoser] {
			dp.VulnerableTo = append(dp.VulnerableTo,
				fmt.Sprintf("%s — %s", h.Hoser, h.Reason))
			seen[h.Hoser] = true
		}
	}
}

// ---------------------------------------------------------------------------
// 3. Opening hand simulation — Monte Carlo mulligan analysis.
// ---------------------------------------------------------------------------

func computeOpeningHandSim(dp *DeckProfile, report *FreyaReport) {
	if report.TotalCards < 40 {
		return
	}

	rng := rand.New(rand.NewSource(42))
	trials := 10000
	keepable := 0
	totalTurnsToFour := 0.0
	validTrials := 0

	landIndices := []int{}
	rampIndices := []int{}
	for i, p := range report.Profiles {
		if p.IsLand {
			landIndices = append(landIndices, i)
		} else {
			for _, r := range p.Produces {
				if r == ResMana {
					rampIndices = append(rampIndices, i)
					break
				}
			}
		}
	}

	deck := make([]int, report.TotalCards)
	for i := range deck {
		deck[i] = i
	}

	for t := 0; t < trials; t++ {
		// Shuffle
		for i := len(deck) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			deck[i], deck[j] = deck[j], deck[i]
		}

		hand := deck[:7]
		landsInHand := 0
		rampInHand := 0
		for _, idx := range hand {
			for _, li := range landIndices {
				if idx == li {
					landsInHand++
					break
				}
			}
			for _, ri := range rampIndices {
				if idx == ri {
					rampInHand++
					break
				}
			}
		}

		// Keepable: 2-5 lands with at least 1 non-land action card
		if landsInHand >= 2 && landsInHand <= 5 && (7-landsInHand) >= 1 {
			keepable++
		}

		// Estimate turns to 4 mana: land drops + ramp
		if landsInHand >= 2 {
			validTrials++
			mana := 0
			turn := 0
			landDropsLeft := landsInHand
			rampLeft := rampInHand
			drawIdx := 7

			for mana < 4 && turn < 10 {
				turn++
				if turn > 1 && drawIdx < len(deck) {
					// Draw a card — check if land or ramp
					drawn := deck[drawIdx]
					drawIdx++
					for _, li := range landIndices {
						if drawn == li {
							landDropsLeft++
							break
						}
					}
					for _, ri := range rampIndices {
						if drawn == ri {
							rampLeft++
							break
						}
					}
				}

				if landDropsLeft > 0 {
					mana++
					landDropsLeft--
				}
				// Play ramp if we have mana for it (assume CMC 2 ramp)
				if rampLeft > 0 && mana >= 2 {
					mana++
					rampLeft--
				}
			}
			totalTurnsToFour += float64(turn)
		}
	}

	dp.KeepableHandPct = float64(keepable) / float64(trials) * 100
	if validTrials > 0 {
		dp.AvgTurnToFourMana = totalTurnsToFour / float64(validTrials)
	}
}

// ---------------------------------------------------------------------------
// 4. Synergy clusters — groups of cards that amplify each other.
// ---------------------------------------------------------------------------

func computeSynergyClusters(dp *DeckProfile, report *FreyaReport, oracle *oracleDB) {
	if oracle == nil || len(report.Profiles) < 10 {
		return
	}

	type cardThemes struct {
		name   string
		themes map[string]bool
	}

	var cards []cardThemes
	for _, p := range report.Profiles {
		if p.IsLand {
			continue
		}
		themes := map[string]bool{}

		if p.IsOutlet {
			themes["sacrifice"] = true
		}
		if profileHasTrigger(p, "dies") || profileHasTrigger(p, "sacrifice") {
			themes["death_value"] = true
		}
		if profileHasTrigger(p, "etb") || p.HasValueETB {
			themes["etb_value"] = true
		}
		if p.IsBlinker {
			themes["blink"] = true
			themes["etb_value"] = true
		}
		for _, r := range p.Produces {
			switch r {
			case ResToken:
				themes["tokens"] = true
			case ResCounter:
				themes["counters"] = true
			case ResCard:
				themes["draw"] = true
			case ResMana:
				themes["mana"] = true
			}
		}
		if p.IsRecursion {
			themes["recursion"] = true
		}
		if profileHasTrigger(p, "cast") {
			themes["spellcast"] = true
		}
		if profileHasTrigger(p, "landfall") {
			themes["landfall"] = true
		}
		if profileHasTrigger(p, "lifegain") || profileHasTrigger(p, "lifeloss") {
			themes["lifegain"] = true
		}
		if profileHasTrigger(p, "counter_matters") || profileHasTrigger(p, "counter_placed") {
			themes["counters"] = true
		}

		if len(themes) > 0 {
			cards = append(cards, cardThemes{name: p.Name, themes: themes})
		}
	}

	// Find clusters by theme overlap
	clusterThemes := map[string][]string{
		"death_value": {},
		"etb_value":   {},
		"tokens":      {},
		"counters":    {},
		"landfall":    {},
		"spellcast":   {},
		"lifegain":    {},
	}

	for _, c := range cards {
		for theme := range c.themes {
			if _, ok := clusterThemes[theme]; ok {
				clusterThemes[theme] = append(clusterThemes[theme], c.name)
			}
		}
	}

	// Also merge related themes: sacrifice + death_value + tokens = aristocrats cluster
	aristocratsCards := map[string]bool{}
	for _, c := range cards {
		if c.themes["sacrifice"] || c.themes["death_value"] || c.themes["tokens"] {
			aristocratsCards[c.name] = true
		}
	}

	clusterNames := map[string]string{
		"death_value": "Death Value Package",
		"etb_value":   "ETB Value Package",
		"tokens":      "Token Engine",
		"counters":    "Counters Synergy",
		"landfall":    "Landfall Package",
		"spellcast":   "Spellslinger Package",
		"lifegain":    "Lifegain Engine",
	}

	for theme, cardNames := range clusterThemes {
		if len(cardNames) < 4 {
			continue
		}

		// Deduplicate
		cardNames = uniqueStrings(cardNames)
		if len(cardNames) < 4 {
			continue
		}

		// Cap display at 8 cards
		displayed := cardNames
		if len(displayed) > 8 {
			displayed = displayed[:8]
		}

		pairCount := len(cardNames) * (len(cardNames) - 1) / 2

		dp.SynergyClusters = append(dp.SynergyClusters, SynergyCluster{
			Name:  clusterNames[theme],
			Cards: displayed,
			Theme: theme,
			Score: pairCount,
		})
	}

	// Sort by score descending
	sort.Slice(dp.SynergyClusters, func(i, j int) bool {
		return dp.SynergyClusters[i].Score > dp.SynergyClusters[j].Score
	})

	// Cap at 5 clusters
	if len(dp.SynergyClusters) > 5 {
		dp.SynergyClusters = dp.SynergyClusters[:5]
	}
}

// ---------------------------------------------------------------------------
// 5. Meta positioning — predicted matchup spread by archetype.
// ---------------------------------------------------------------------------

type matchupEntry struct {
	vsArchetype string
	rating      string
	reason      string
}

var metaMatchupDB = map[string][]matchupEntry{
	"aggro": {
		{vsArchetype: "Control", rating: "unfavored", reason: "board wipes and card advantage grind you out"},
		{vsArchetype: "Combo", rating: "favored", reason: "fast clock pressures combo before assembly"},
		{vsArchetype: "Midrange", rating: "neutral", reason: "race depends on draw quality"},
	},
	"combo": {
		{vsArchetype: "Aggro", rating: "unfavored", reason: "fast damage before combo assembles"},
		{vsArchetype: "Control", rating: "neutral", reason: "counterspells vs speed — draw dependent"},
		{vsArchetype: "Stax", rating: "unfavored", reason: "resource denial prevents combo assembly"},
		{vsArchetype: "Midrange", rating: "favored", reason: "goldfish speed outraces midrange value"},
	},
	"control": {
		{vsArchetype: "Aggro", rating: "favored", reason: "board wipes and removal stabilize against creatures"},
		{vsArchetype: "Combo", rating: "neutral", reason: "need to hold counterspells for the right moment"},
		{vsArchetype: "Midrange", rating: "favored", reason: "card advantage wins the long game"},
		{vsArchetype: "Stax", rating: "neutral", reason: "both play long game, stax taxes your answers"},
	},
	"aristocrats": {
		{vsArchetype: "Aggro", rating: "favored", reason: "resilient to removal, drain bypasses combat"},
		{vsArchetype: "Control", rating: "neutral", reason: "recursive threats are hard to answer but slow"},
		{vsArchetype: "Combo", rating: "unfavored", reason: "too slow to race dedicated combo"},
		{vsArchetype: "Graveyard Hate", rating: "unfavored", reason: "Rest in Peace effects shut down the engine"},
	},
	"voltron": {
		{vsArchetype: "Control", rating: "unfavored", reason: "single threat is easy to answer with removal"},
		{vsArchetype: "Token/Go Wide", rating: "unfavored", reason: "chump blockers stall commander damage"},
		{vsArchetype: "Combo", rating: "neutral", reason: "fast commander kills can race some combos"},
	},
	"stax": {
		{vsArchetype: "Combo", rating: "favored", reason: "resource denial prevents combo assembly"},
		{vsArchetype: "Aggro", rating: "favored", reason: "taxes and locks slow aggro to a crawl"},
		{vsArchetype: "Control", rating: "neutral", reason: "both play long game but stax constraints hurt both"},
		{vsArchetype: "Midrange", rating: "favored", reason: "value engines need resources stax denies"},
	},
	"reanimator": {
		{vsArchetype: "Aggro", rating: "favored", reason: "early fatties outclass aggro creatures"},
		{vsArchetype: "Control", rating: "neutral", reason: "counterspells stop reanimation, but recursive"},
		{vsArchetype: "Graveyard Hate", rating: "unfavored", reason: "any graveyard exile effect is devastating"},
	},
	"storm": {
		{vsArchetype: "Stax", rating: "unfavored", reason: "Rule of Law effects are game over"},
		{vsArchetype: "Aggro", rating: "neutral", reason: "race depends on who assembles first"},
		{vsArchetype: "Control", rating: "unfavored", reason: "counterspells disrupt the chain"},
		{vsArchetype: "Midrange", rating: "favored", reason: "combo kill before midrange value matters"},
	},
	"enchantress": {
		{vsArchetype: "Aggro", rating: "neutral", reason: "pillow fort effects can stabilize if drawn early"},
		{vsArchetype: "Enchantment Hate", rating: "unfavored", reason: "mass enchantment removal is devastating"},
		{vsArchetype: "Combo", rating: "unfavored", reason: "engine too slow to race dedicated combo"},
	},
	"midrange": {
		{vsArchetype: "Aggro", rating: "neutral", reason: "bigger creatures but slower start"},
		{vsArchetype: "Control", rating: "unfavored", reason: "outgrinded in long games"},
		{vsArchetype: "Combo", rating: "unfavored", reason: "too fair to race combo"},
	},
}

func computeMetaPositioning(dp *DeckProfile) {
	arch := strings.ToLower(dp.PrimaryArchetype)

	// Normalize some archetype names to match the DB
	switch {
	case containsAny(arch, "aggro", "go wide", "tribal", "extra combats"):
		arch = "aggro"
	case containsAny(arch, "combo", "infinite"):
		arch = "combo"
	case containsAny(arch, "stax"):
		arch = "stax"
	case containsAny(arch, "aristocrats"):
		arch = "aristocrats"
	case containsAny(arch, "voltron"):
		arch = "voltron"
	case containsAny(arch, "reanimator"):
		arch = "reanimator"
	case containsAny(arch, "storm", "spellslinger"):
		arch = "storm"
	case containsAny(arch, "enchantress"):
		arch = "enchantress"
	case containsAny(arch, "control", "mill", "discard"):
		arch = "control"
	default:
		arch = "midrange"
	}

	matchups, ok := metaMatchupDB[arch]
	if !ok {
		return
	}

	for _, m := range matchups {
		dp.MetaMatchups = append(dp.MetaMatchups, MetaMatchup{
			Archetype: m.vsArchetype,
			Rating:    m.rating,
			Reason:    m.reason,
		})
	}
}

// ---------------------------------------------------------------------------
// 6. Card quality tiers — identify star performers and cuttable cards.
// ---------------------------------------------------------------------------

func computeCardQualityTiers(dp *DeckProfile, report *FreyaReport, oracle *oracleDB) {
	if report.Roles == nil {
		return
	}

	type cardScore struct {
		name   string
		score  float64
		roles  []RoleTag
		cmc    int
		reason string
	}

	var scores []cardScore

	// Build role assignment lookup
	roleMap := map[string][]RoleTag{}
	for _, a := range report.Roles.Assignments {
		roleMap[a.Name] = a.Roles
	}

	// Score combo pieces mentioned in win lines
	winLinePieces := map[string]bool{}
	if report.WinLines != nil {
		for _, wl := range report.WinLines.WinLines {
			for _, piece := range wl.Pieces {
				winLinePieces[piece] = true
			}
		}
	}

	// Score value chain pieces
	chainPieces := map[string]bool{}
	bridgePieces := map[string]bool{}
	for _, vc := range report.ValueChains {
		for _, step := range vc.Steps {
			for _, card := range step.Cards {
				chainPieces[card] = true
			}
		}
		for _, b := range vc.BridgeCards {
			bridgePieces[b] = true
		}
	}

	for _, p := range report.Profiles {
		if p.IsLand {
			continue
		}

		s := cardScore{
			name:  p.Name,
			cmc:   p.CMC,
			roles: roleMap[p.Name],
		}

		// Multi-role cards score higher
		s.score += float64(len(s.roles)) * 1.0

		// Win line pieces are stars
		if winLinePieces[p.Name] {
			s.score += 3.0
			s.reason = "win condition piece"
		}

		// Bridge cards in value chains are highly valuable
		if bridgePieces[p.Name] {
			s.score += 2.5
			if s.reason == "" {
				s.reason = "value chain bridge card"
			}
		} else if chainPieces[p.Name] {
			s.score += 1.0
		}

		// CMC efficiency: low CMC with multiple roles is efficient
		if p.CMC <= 2 && len(s.roles) >= 2 {
			s.score += 1.5
			if s.reason == "" {
				s.reason = "efficient multi-role at low CMC"
			}
		}

		// High CMC with only utility role is likely cuttable
		if p.CMC >= 5 && len(s.roles) == 1 && s.roles[0] == RoleUtility {
			s.score -= 2.0
			s.reason = "high CMC with no clear role"
		}

		// Cards with only Utility role and high CMC
		if p.CMC >= 4 && len(s.roles) == 1 && s.roles[0] == RoleUtility {
			s.score -= 1.0
			if s.reason == "" {
				s.reason = "filler — no synergy role at CMC " + fmt.Sprint(p.CMC)
			}
		}

		// Tutors that are worse versions of other tutors in the deck
		if p.IsTutor && p.CMC >= 4 {
			cheaperTutors := 0
			for _, other := range report.Profiles {
				if other.IsTutor && other.CMC < p.CMC && other.Name != p.Name {
					cheaperTutors++
				}
			}
			if cheaperTutors >= 3 {
				s.score -= 1.5
				s.reason = fmt.Sprintf("expensive tutor (CMC %d) with %d cheaper alternatives", p.CMC, cheaperTutors)
			}
		}

		scores = append(scores, s)
	}

	// Sort by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Top 5 = stars
	for i := 0; i < len(scores) && i < 5; i++ {
		if scores[i].score >= 3.0 {
			reason := scores[i].reason
			if reason == "" {
				reason = "high synergy density"
			}
			dp.StarCards = append(dp.StarCards, CardQuality{
				Name:   scores[i].name,
				Tier:   "star",
				Reason: reason,
			})
		}
	}

	// Bottom cards with low scores = cuttable
	for i := len(scores) - 1; i >= 0 && i > len(scores)-6; i-- {
		if scores[i].score <= 0 {
			reason := scores[i].reason
			if reason == "" {
				reason = "low synergy with deck strategy"
			}
			dp.CuttableCards = append(dp.CuttableCards, CardQuality{
				Name:   scores[i].name,
				Tier:   "cuttable",
				Reason: reason,
			})
		}
	}
}

// ---------------------------------------------------------------------------
// 8. Color weight optimization — suggest specific land swaps.
// ---------------------------------------------------------------------------

func computeLandSwapSuggestions(dp *DeckProfile, report *FreyaReport) {
	if report.Stats == nil {
		return
	}

	totalDemand := 0
	totalSupply := 0
	demand := map[string]int{}
	supply := map[string]int{}

	for _, c := range []string{"W", "U", "B", "R", "G"} {
		d := report.ColorDemand[c]
		s := report.ColorSupply[c]
		demand[c] = d
		supply[c] = s
		totalDemand += d
		totalSupply += s
	}

	if totalDemand == 0 || totalSupply == 0 {
		return
	}

	type colorImbalance struct {
		color     string
		demandPct float64
		supplyPct float64
		gap       float64
	}

	var imbalances []colorImbalance
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		if demand[c] == 0 && supply[c] == 0 {
			continue
		}
		dPct := float64(demand[c]) / float64(totalDemand)
		sPct := float64(supply[c]) / float64(totalSupply)
		gap := dPct - sPct
		if math.Abs(gap) > 0.05 {
			imbalances = append(imbalances, colorImbalance{
				color: c, demandPct: dPct, supplyPct: sPct, gap: gap,
			})
		}
	}

	// Sort: biggest undersupply first
	sort.Slice(imbalances, func(i, j int) bool {
		return imbalances[i].gap > imbalances[j].gap
	})

	colorNames := map[string]string{"W": "Plains", "U": "Island", "B": "Swamp", "R": "Mountain", "G": "Forest"}

	for _, ib := range imbalances {
		if ib.gap > 0.05 {
			// Undersupplied — find oversupplied color to swap from
			for _, other := range imbalances {
				if other.gap < -0.05 {
					dp.LandSwapSuggestions = append(dp.LandSwapSuggestions,
						fmt.Sprintf("Replace 1 %s with 1 %s: %s has %.0f%% demand but only %.0f%% sources",
							colorNames[other.color], colorNames[ib.color],
							ib.color, ib.demandPct*100, ib.supplyPct*100))
					break
				}
			}
		}
	}

	// Cap suggestions
	if len(dp.LandSwapSuggestions) > 3 {
		dp.LandSwapSuggestions = dp.LandSwapSuggestions[:3]
	}
}

// ---------------------------------------------------------------------------
// 9. Deck personality blurb — 2-3 sentence flavor description.
// ---------------------------------------------------------------------------

func buildPersonalityBlurb(dp *DeckProfile, report *FreyaReport) string {
	arch := strings.ToLower(dp.PrimaryArchetype)
	speed := "methodical"
	if dp.AvgCMC < 2.5 {
		speed = "lightning-fast"
	} else if dp.AvgCMC < 3.0 {
		speed = "agile"
	} else if dp.AvgCMC > 3.8 {
		speed = "slow but devastating"
	}

	var approach string
	switch {
	case containsAny(arch, "combo", "infinite", "storm"):
		if dp.HasTutorAccess && report.TutorCount >= 5 {
			approach = "It assembles its kill with surgical precision, tutoring combo pieces while holding up protection."
		} else {
			approach = "It digs aggressively for its combo pieces, racing to assemble a kill before opponents can disrupt it."
		}
	case containsAny(arch, "control", "stax"):
		approach = "It grinds opponents into submission, answering threats while slowly accumulating an insurmountable advantage."
	case containsAny(arch, "aggro", "go wide", "tribal"):
		approach = "It floods the board and turns sideways, overwhelming opponents before they can stabilize."
	case containsAny(arch, "voltron"):
		approach = "It suits up its commander and swings for lethal, protecting its investment with shields and counters."
	case containsAny(arch, "aristocrats"):
		approach = "It feeds the death machine — sacrificing and recurring creatures in a loop of incremental drains that bypass combat entirely."
	case containsAny(arch, "reanimator"):
		approach = "It cheats massive threats into play from the graveyard, bypassing mana costs for devastating early haymakers."
	case containsAny(arch, "enchantress"):
		approach = "It weaves a web of enchantments, drawing cards off each one until the value engine becomes unstoppable."
	case containsAny(arch, "lands"):
		approach = "It turns land drops into a resource engine, triggering landfall chains that generate exponential value."
	case containsAny(arch, "blink", "flicker"):
		approach = "It blinks creatures in and out of existence, squeezing maximum value from every ETB trigger."
	case containsAny(arch, "mill"):
		approach = "It attacks libraries instead of life totals, grinding opponents out card by card until they draw from nothing."
	case containsAny(arch, "superfriends"):
		approach = "It deploys an army of planeswalkers, ticking up loyalty counters behind a wall of protection until ultimates end the game."
	default:
		approach = "It plays a flexible game, adapting its strategy based on the table and finding the right moment to strike."
	}

	var closer string
	if dp.Bracket >= 4 {
		closer = fmt.Sprintf("Built at bracket %d (%s), this is a deck that demands respect from turn one.", dp.Bracket, dp.BracketLabel)
	} else if dp.WinLineCount >= 3 {
		closer = fmt.Sprintf("With %d paths to victory, it always has a plan B.", dp.WinLineCount)
	} else if dp.CommanderSynergy >= 0.50 {
		closer = fmt.Sprintf("Tightly built around its commander's strengths, every card pulls its weight.")
	} else {
		closer = fmt.Sprintf("A solid %s build that rewards patient piloting.", dp.BracketLabel)
	}

	return fmt.Sprintf("This is a %s %s deck. %s %s", speed, dp.PrimaryArchetype, approach, closer)
}

// ---------------------------------------------------------------------------
// 10. Power percentile — estimated ranking within archetype.
// ---------------------------------------------------------------------------

func estimatePowerPercentile(dp *DeckProfile, report *FreyaReport) (int, []string) {
	score := 50 // start at median
	var factors []string

	// Tutor density
	if report.TutorCount >= 8 {
		score += 15
		factors = append(factors, fmt.Sprintf("deep tutor package (%d) → +15", report.TutorCount))
	} else if report.TutorCount >= 5 {
		score += 8
		factors = append(factors, fmt.Sprintf("solid tutor access (%d) → +8", report.TutorCount))
	} else if report.TutorCount <= 1 {
		score -= 10
		factors = append(factors, fmt.Sprintf("minimal tutors (%d) → -10", report.TutorCount))
	}

	// Win line count
	if dp.WinLineCount >= 5 {
		score += 10
		factors = append(factors, fmt.Sprintf("diverse win lines (%d) → +10", dp.WinLineCount))
	} else if dp.WinLineCount >= 3 {
		score += 5
	} else if dp.WinLineCount <= 1 {
		score -= 10
		factors = append(factors, "limited win conditions → -10")
	}

	// Mana base quality
	switch dp.ManaBaseGrade {
	case "A":
		score += 10
		factors = append(factors, "A-grade mana base → +10")
	case "B":
		score += 5
	case "D", "F":
		score -= 10
		factors = append(factors, fmt.Sprintf("%s-grade mana base → -10", dp.ManaBaseGrade))
	}

	// Interaction quality
	if dp.InteractionQuality > 0 && dp.InteractionQuality <= 2.0 {
		score += 8
		factors = append(factors, fmt.Sprintf("fast interaction (avg CMC %.1f) → +8", dp.InteractionQuality))
	} else if dp.InteractionQuality > 3.5 {
		score -= 5
		factors = append(factors, fmt.Sprintf("slow interaction (avg CMC %.1f) → -5", dp.InteractionQuality))
	}

	// Draw density
	if dp.DrawCount >= 12 {
		score += 8
		factors = append(factors, fmt.Sprintf("excellent draw (%d sources) → +8", dp.DrawCount))
	} else if dp.DrawCount < 5 {
		score -= 8
		factors = append(factors, fmt.Sprintf("low draw (%d sources) → -8", dp.DrawCount))
	}

	// Commander synergy
	if dp.CommanderSynergy >= 0.60 {
		score += 5
		factors = append(factors, fmt.Sprintf("strong commander synergy (%.0f%%) → +5", dp.CommanderSynergy*100))
	} else if dp.CommanderSynergy > 0 && dp.CommanderSynergy < 0.25 {
		score -= 5
		factors = append(factors, "low commander synergy → -5")
	}

	// Average CMC (lean curves are generally better)
	if dp.AvgCMC < 2.5 {
		score += 5
		factors = append(factors, fmt.Sprintf("lean curve (%.1f avg) → +5", dp.AvgCMC))
	} else if dp.AvgCMC > 3.8 {
		score -= 5
		factors = append(factors, fmt.Sprintf("heavy curve (%.1f avg) → -5", dp.AvgCMC))
	}

	// Keepable hand rate
	if dp.KeepableHandPct >= 85 {
		score += 5
		factors = append(factors, fmt.Sprintf("consistent opening hands (%.0f%% keepable) → +5", dp.KeepableHandPct))
	} else if dp.KeepableHandPct > 0 && dp.KeepableHandPct < 70 {
		score -= 5
		factors = append(factors, fmt.Sprintf("inconsistent opening hands (%.0f%% keepable) → -5", dp.KeepableHandPct))
	}

	// Clamp to 1-99
	if score < 1 {
		score = 1
	}
	if score > 99 {
		score = 99
	}

	return score, factors
}
