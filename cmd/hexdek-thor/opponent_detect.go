package main

// opponent_detect.go — board-state enrichment for opponent-referencing cards.
//
// Complementary to opponent_autodetect.go (which fires adversarial
// ACTIONS like cast/attack/draw), this file detects what BOARD STATE
// the opponent needs so that targeting effects can resolve. A card like
// "destroy target creature an opponent controls" needs the opponent to
// actually have a creature on the battlefield; "exile target artifact
// an opponent controls" needs an artifact; etc.
//
// The detection scans oracle text for canonical MTG targeting patterns
// and returns an OpponentRequirement describing what types of permanents,
// hand cards, graveyard cards, or library cards the opponent seat needs.
// EnrichOpponentSeat then materializes the required objects.

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// OpponentRequirement describes which board-state elements the opponent
// seat needs for the card under test to have valid targets / interactions.
type OpponentRequirement struct {
	NeedsCreatures    bool
	NeedsArtifacts    bool
	NeedsEnchantments bool
	NeedsLife         bool
	NeedsCast         bool
	NeedsAttack       bool
	NeedsHand         bool
	NeedsGraveyard    bool
	NeedsLibrary      bool
	NeedsPlaneswalker bool
	NeedsLand         bool
}

// HasAny returns true if any board-state enrichment is needed.
func (r OpponentRequirement) HasAny() bool {
	return r.NeedsCreatures || r.NeedsArtifacts || r.NeedsEnchantments ||
		r.NeedsLife || r.NeedsCast || r.NeedsAttack ||
		r.NeedsHand || r.NeedsGraveyard || r.NeedsLibrary ||
		r.NeedsPlaneswalker || r.NeedsLand
}

// DetectOpponentRequirements scans oracle text for patterns that imply
// the opponent needs specific board-state elements. Patterns include
// targeting ("target creature an opponent controls"), broad references
// ("each opponent's creatures"), zone references ("opponent's graveyard",
// "opponent's hand"), and life/cast/attack references.
func DetectOpponentRequirements(oracleText string) OpponentRequirement {
	var req OpponentRequirement
	text := strings.ToLower(oracleText)

	// No opponent reference at all — early exit.
	if !strings.Contains(text, "opponent") {
		return req
	}

	// --- Permanent type targeting ---
	// "target creature an opponent controls"
	// "creature an opponent controls"
	// "creatures your opponents control"
	// "each creature an opponent controls"
	// "all creatures opponents control"
	creaturePatterns := []string{
		"creature an opponent controls",
		"creature your opponents control",
		"creatures an opponent controls",
		"creatures your opponents control",
		"creatures opponents control",
		"opponent controls a creature",
		"each opponent's creature",
		"opponent's creature",
		"nonland permanent an opponent controls", // covers creatures too
		"permanent an opponent controls",
		"permanents your opponents control",
		"nonland permanents your opponents control",
	}
	for _, p := range creaturePatterns {
		if strings.Contains(text, p) {
			req.NeedsCreatures = true
			break
		}
	}

	// Artifact targeting
	artifactPatterns := []string{
		"artifact an opponent controls",
		"artifact your opponents control",
		"artifacts an opponent controls",
		"artifacts your opponents control",
		"artifacts opponents control",
		"opponent controls an artifact",
		"opponent's artifact",
		"nonland permanent an opponent controls",
		"permanent an opponent controls",
	}
	for _, p := range artifactPatterns {
		if strings.Contains(text, p) {
			req.NeedsArtifacts = true
			break
		}
	}

	// Enchantment targeting
	enchantmentPatterns := []string{
		"enchantment an opponent controls",
		"enchantment your opponents control",
		"enchantments an opponent controls",
		"enchantments your opponents control",
		"enchantments opponents control",
		"opponent controls an enchantment",
		"opponent's enchantment",
		"nonland permanent an opponent controls",
		"permanent an opponent controls",
	}
	for _, p := range enchantmentPatterns {
		if strings.Contains(text, p) {
			req.NeedsEnchantments = true
			break
		}
	}

	// Planeswalker targeting
	planeswalkerPatterns := []string{
		"planeswalker an opponent controls",
		"planeswalker your opponents control",
		"opponent controls a planeswalker",
		"opponent's planeswalker",
	}
	for _, p := range planeswalkerPatterns {
		if strings.Contains(text, p) {
			req.NeedsPlaneswalker = true
			break
		}
	}

	// Land targeting
	landPatterns := []string{
		"land an opponent controls",
		"land your opponents control",
		"lands an opponent controls",
		"lands your opponents control",
		"lands opponents control",
		"opponent controls a land",
		"opponent's land",
		"opponent controls more lands",
	}
	for _, p := range landPatterns {
		if strings.Contains(text, p) {
			req.NeedsLand = true
			break
		}
	}

	// --- Life references ---
	lifePatterns := []string{
		"opponent loses life",
		"opponent loses 1 life",
		"opponent loses 2 life",
		"opponent loses 3 life",
		"each opponent loses",
		"opponent's life total",
		"opponent has less life",
		"opponent has more life",
		"opponent would lose life",
		"damage to an opponent",
		"damage to each opponent",
		"deals damage to an opponent",
		"deals combat damage to an opponent",
		"opponent is dealt damage",
	}
	for _, p := range lifePatterns {
		if strings.Contains(text, p) {
			req.NeedsLife = true
			break
		}
	}

	// --- Cast references ---
	castPatterns := []string{
		"opponent casts",
		"opponents cast",
		"whenever an opponent casts",
		"an opponent casts a spell",
	}
	for _, p := range castPatterns {
		if strings.Contains(text, p) {
			req.NeedsCast = true
			break
		}
	}

	// --- Attack references ---
	attackPatterns := []string{
		"opponent attacks",
		"opponents attack",
		"whenever an opponent attacks",
		"an opponent attacks you",
		"opponent declares attackers",
	}
	for _, p := range attackPatterns {
		if strings.Contains(text, p) {
			req.NeedsAttack = true
			break
		}
	}

	// --- Hand zone references ---
	handPatterns := []string{
		"opponent's hand",
		"opponent discards",
		"opponents discard",
		"each opponent discards",
		"opponent reveals",
		"opponent's hand",
		"from an opponent's hand",
		"cards in an opponent's hand",
	}
	for _, p := range handPatterns {
		if strings.Contains(text, p) {
			req.NeedsHand = true
			break
		}
	}

	// --- Graveyard zone references ---
	graveyardPatterns := []string{
		"opponent's graveyard",
		"opponents' graveyards",
		"an opponent's graveyard",
		"from an opponent's graveyard",
		"card from an opponent's graveyard",
		"cards in an opponent's graveyard",
		"exile target card from an opponent's graveyard",
		"opponent mills",
		"each opponent mills",
	}
	for _, p := range graveyardPatterns {
		if strings.Contains(text, p) {
			req.NeedsGraveyard = true
			break
		}
	}

	// --- Library zone references ---
	libraryPatterns := []string{
		"opponent's library",
		"opponents' libraries",
		"an opponent's library",
		"top of an opponent's library",
		"from the top of an opponent's library",
		"opponent searches their library",
		"opponent mills",
		"each opponent mills",
	}
	for _, p := range libraryPatterns {
		if strings.Contains(text, p) {
			req.NeedsLibrary = true
			break
		}
	}

	return req
}

// EnrichOpponentSeat populates the opponent seat (oppSeat) with the
// permanents, hand cards, graveyard cards, and library cards dictated
// by the requirement. Idempotent — skips enrichment categories where
// the seat already has adequate objects.
func EnrichOpponentSeat(gs *gameengine.GameState, oppSeat int, req OpponentRequirement) {
	if gs == nil || oppSeat < 0 || oppSeat >= len(gs.Seats) || !req.HasAny() {
		return
	}
	seat := gs.Seats[oppSeat]
	if seat == nil {
		return
	}

	if req.NeedsCreatures {
		// Add 2-3 vanilla creature tokens if the seat doesn't already
		// have enough creatures.
		existingCreatures := countPermanentsByType(seat, "creature")
		needed := 3 - existingCreatures
		for i := 0; i < needed; i++ {
			card := &gameengine.Card{
				Name:          "Opponent Creature Token",
				Owner:         oppSeat,
				Types:         []string{"creature", "token"},
				BasePower:     2,
				BaseToughness: 2,
			}
			perm := &gameengine.Permanent{
				Card:       card,
				Controller: oppSeat,
				Owner:      oppSeat,
				Flags:      map[string]int{},
			}
			seat.Battlefield = append(seat.Battlefield, perm)
		}
	}

	if req.NeedsArtifacts {
		existingArtifacts := countPermanentsByType(seat, "artifact")
		if existingArtifacts < 1 {
			card := &gameengine.Card{
				Name:  "Opponent Sol Ring",
				Owner: oppSeat,
				Types: []string{"artifact"},
			}
			perm := &gameengine.Permanent{
				Card:       card,
				Controller: oppSeat,
				Owner:      oppSeat,
				Flags:      map[string]int{},
			}
			seat.Battlefield = append(seat.Battlefield, perm)
		}
	}

	if req.NeedsEnchantments {
		existingEnchantments := countPermanentsByType(seat, "enchantment")
		if existingEnchantments < 1 {
			card := &gameengine.Card{
				Name:  "Opponent Enchantment",
				Owner: oppSeat,
				Types: []string{"enchantment"},
			}
			perm := &gameengine.Permanent{
				Card:       card,
				Controller: oppSeat,
				Owner:      oppSeat,
				Flags:      map[string]int{},
			}
			seat.Battlefield = append(seat.Battlefield, perm)
		}
	}

	if req.NeedsPlaneswalker {
		existingPW := countPermanentsByType(seat, "planeswalker")
		if existingPW < 1 {
			card := &gameengine.Card{
				Name:  "Opponent Planeswalker",
				Owner: oppSeat,
				Types: []string{"planeswalker"},
			}
			perm := &gameengine.Permanent{
				Card:       card,
				Controller: oppSeat,
				Owner:      oppSeat,
				Counters:   map[string]int{"loyalty": 3},
				Flags:      map[string]int{},
			}
			seat.Battlefield = append(seat.Battlefield, perm)
		}
	}

	if req.NeedsLand {
		existingLands := countPermanentsByType(seat, "land")
		needed := 3 - existingLands
		for i := 0; i < needed; i++ {
			card := &gameengine.Card{
				Name:  "Opponent Land",
				Owner: oppSeat,
				Types: []string{"land"},
			}
			perm := &gameengine.Permanent{
				Card:       card,
				Controller: oppSeat,
				Owner:      oppSeat,
				Flags:      map[string]int{},
			}
			seat.Battlefield = append(seat.Battlefield, perm)
		}
	}

	if req.NeedsLife {
		// Ensure the opponent has a meaningful life total.
		if seat.Life < 20 {
			seat.Life = 20
		}
	}

	if req.NeedsHand {
		// Ensure the opponent has cards in hand for discard / reveal effects.
		existingHand := len(seat.Hand)
		needed := 5 - existingHand
		for i := 0; i < needed; i++ {
			card := &gameengine.Card{
				Name:  "Opponent Hand Card",
				Owner: oppSeat,
				Types: []string{"creature"},
			}
			seat.Hand = append(seat.Hand, card)
		}
	}

	if req.NeedsGraveyard {
		// Ensure the opponent has cards in graveyard for exile / reanimate effects.
		existingGY := len(seat.Graveyard)
		needed := 4 - existingGY
		for i := 0; i < needed; i++ {
			card := &gameengine.Card{
				Name:          "Opponent Graveyard Card",
				Owner:         oppSeat,
				Types:         []string{"creature"},
				BasePower:     2,
				BaseToughness: 2,
			}
			seat.Graveyard = append(seat.Graveyard, card)
		}
	}

	if req.NeedsLibrary {
		// Ensure the opponent has cards in library for mill / search effects.
		existingLib := len(seat.Library)
		needed := 10 - existingLib
		for i := 0; i < needed; i++ {
			card := &gameengine.Card{
				Name:          "Opponent Library Card",
				Owner:         oppSeat,
				Types:         []string{"creature"},
				BasePower:     1,
				BaseToughness: 1,
			}
			seat.Library = append(seat.Library, card)
		}
	}
}

// countPermanentsByType counts permanents on a seat's battlefield that
// have the given type in their Card.Types slice.
func countPermanentsByType(seat *gameengine.Seat, typeName string) int {
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		for _, t := range p.Card.Types {
			if t == typeName {
				count++
				break
			}
		}
	}
	return count
}
