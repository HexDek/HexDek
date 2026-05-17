package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRiverSongsDiary wires River Song's Diary (Muninn parser-gap #47,
// 16,900 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{3}
//	Artifact — Book
//	Imprint — Whenever a player casts an instant or sorcery spell from
//	their hand, exile it instead of putting it into a graveyard as it
//	resolves.
//	At the beginning of your upkeep, if there are four or more cards
//	exiled with this artifact, choose one of them at random. You may
//	cast it without paying its mana cost.
//
// Implementation:
//   - "spell_cast": for every instant/sorcery spell cast from hand by any
//     player, tag the spell's StackItem.CostMeta with
//     ["river_song_imprint_ts"] = perm.Timestamp and
//     ["exile_on_resolve"] = true so the engine routes the resolved card
//     to exile instead of graveyard (Smirking Spelljacker precedent).
//     Also stamp Card.ExiledByTimestamp eagerly so upkeep lookup works
//     without depending on CostMeta propagation.
//   - "upkeep_controller": count cards across every seat's Exile pile
//     whose ExiledByTimestamp == perm.Timestamp. If >=4, log the
//     candidate (random pick is partial — same free-cast-arbitrary-spell
//     gap that Smirking Spelljacker and Transcendent Dragon document).
func registerRiverSongsDiary(r *Registry) {
	r.OnTrigger("River Song's Diary", "spell_cast", riverSongDiarySpellCast)
	r.OnTrigger("River Song's Diary", "upkeep_controller", riverSongDiaryUpkeep)
}

func riverSongDiarySpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "river_song_diary_imprint"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if !cardHasType(card, "instant") && !cardHasType(card, "sorcery") {
		return
	}
	fromZone, _ := ctx["from_zone"].(string)
	// The oracle text gates on "from their hand". If the engine doesn't
	// supply from_zone, accept the cast — non-hand cast paths (flashback,
	// foretell) are rare on instants/sorceries and the worst case here is
	// over-imprinting, not under-imprinting.
	if fromZone != "" && fromZone != "hand" {
		return
	}
	// Find the matching stack item to tag.
	var si *gameengine.StackItem
	for i := len(gs.Stack) - 1; i >= 0; i-- {
		if gs.Stack[i] != nil && gs.Stack[i].Card == card {
			si = gs.Stack[i]
			break
		}
	}
	if si == nil {
		return
	}
	if si.CostMeta == nil {
		si.CostMeta = map[string]interface{}{}
	}
	si.CostMeta["exile_on_resolve"] = true
	si.CostMeta["exiled_by_timestamp"] = perm.Timestamp
	card.ExiledByTimestamp = perm.Timestamp
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"caster":  si.Controller,
		"spell":   card.DisplayName(),
		"link_ts": perm.Timestamp,
	})
}

func riverSongDiaryUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "river_song_diary_upkeep_free_cast"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	imprints := []*gameengine.Card{}
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, c := range s.Exile {
			if c == nil {
				continue
			}
			if c.ExiledByTimestamp == perm.Timestamp {
				imprints = append(imprints, c)
			}
		}
	}
	if len(imprints) < 4 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"imprinted": len(imprints),
			"triggered": false,
		})
		return
	}
	// "Choose one of them at random" — deterministic pick via gs.Turn
	// modulo, threaded through perm.Timestamp so two diaries don't
	// collide on the same selection. The free-cast resolution path is
	// the engine gap documented on Smirking Spelljacker.
	pickIdx := (gs.Turn + perm.Timestamp) % len(imprints)
	if pickIdx < 0 {
		pickIdx = 0
	}
	picked := imprints[pickIdx]
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"imprinted": len(imprints),
		"candidate": picked.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"may_cast_random_imprinted_spell_without_paying_mana_no_engine_free_cast_path")
}
