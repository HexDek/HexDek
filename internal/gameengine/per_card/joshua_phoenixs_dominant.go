package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJoshuaPhoenixsDominant wires Joshua, Phoenix's Dominant //
// Phoenix, Warden of Fire.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	  Joshua (front, Human Noble Wizard):
//		When Joshua enters, discard up to two cards, then draw that many
//		  cards.
//		{3}{R}{W}, {T}: Exile Joshua, then return it to the battlefield
//		  transformed under its owner's control. Activate only as a
//		  sorcery.
//
//	  Phoenix, Warden of Fire (back, Saga Phoenix):
//		(As this Saga enters and after your draw step, add a lore counter.)
//		I, II — Rising Flames — Phoenix deals 2 damage to each opponent.
//		III — Flames of Rebirth — Return any number of target creature
//		  cards with total mana value 6 or less from your graveyard to the
//		  battlefield. Exile Phoenix, then return it to the battlefield
//		  (front face up).
//		Flying, lifelink
//
// Implementation:
//   - OnETB (Joshua front): discard up to 2 highest-CMC cards in hand,
//     draw that many. Skip discard if hand is empty.
//   - Activated transform and Saga back-face mechanics are continuous
//     mechanics outside per-card scope — emitPartial.
func registerJoshuaPhoenixsDominant(r *Registry) {
	r.OnETB("Joshua, Phoenix's Dominant", joshuaPhoenixsETB)
	r.OnETB("Joshua, Phoenix's Dominant // Phoenix, Warden of Fire", joshuaPhoenixsETB)
	r.OnActivated("Joshua, Phoenix's Dominant", joshuaPhoenixsActivate)
	r.OnActivated("Joshua, Phoenix's Dominant // Phoenix, Warden of Fire", joshuaPhoenixsActivate)
}

func joshuaPhoenixsETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "joshua_phoenixs_etb_loot"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	maxDiscard := 2
	if len(s.Hand) < maxDiscard {
		maxDiscard = len(s.Hand)
	}
	discarded := []string{}
	for i := 0; i < maxDiscard; i++ {
		// Highest-CMC remaining card.
		idx := -1
		bestCMC := -1
		for j, c := range s.Hand {
			if c == nil {
				continue
			}
			if cm := gameengine.ManaCostOf(c); cm > bestCMC {
				bestCMC = cm
				idx = j
			}
		}
		if idx < 0 {
			break
		}
		card := s.Hand[idx]
		discarded = append(discarded, card.DisplayName())
		gameengine.MoveCard(gs, card, seat, "hand", "graveyard", "joshua_etb")
		gs.LogEvent(gameengine.Event{
			Kind:   "discard",
			Seat:   seat,
			Source: perm.Card.DisplayName(),
			Details: map[string]interface{}{
				"slug": slug,
				"card": card.DisplayName(),
			},
		})
	}
	drawn := 0
	for i := 0; i < len(discarded); i++ {
		if drawOne(gs, seat, perm.Card.DisplayName()) != nil {
			drawn++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      seat,
		"discarded": discarded,
		"drawn":     drawn,
	})
}

func joshuaPhoenixsActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, "joshua_phoenixs_transform", src.Card.DisplayName(),
		"transform_to_phoenix_warden_of_fire_back_face_not_implemented")
}
