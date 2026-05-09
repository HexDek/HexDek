package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTiamatCustom wires the real ETB tutor for Tiamat. The
// auto-generated registerTiamat in gen_tiamat.go remains in place
// (it now emits a partial breadcrumb only) — this handler runs after
// it via the registry's append-on-register list.
//
// Oracle text (Adventures in the Forgotten Realms, {2}{W}{U}{B}{R}{G}):
//
//	Flying
//	When Tiamat enters, if you cast it, search your library for up to
//	five Dragon cards not named Tiamat that each have different names,
//	reveal them, put them into your hand, then shuffle.
//
// Implementation:
//   - "if you cast it" — gate on perm.Flags["was_cast"] == 1.
//   - Search library for up to five non-Tiamat Dragon creature cards
//     with distinct names. We pick deterministically: scan the library
//     in order, accept the first card matching each unique name until
//     we have five.
//   - Move each found card from library to hand via MoveCard so
//     zone-change triggers fire.
//   - Shuffle the library afterwards.
func registerTiamatCustom(r *Registry) {
	r.OnETB("Tiamat", tiamatCustomETB)
}

func tiamatCustomETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "tiamat_etb_dragon_tutor"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	// "if you cast it" gate.
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}

	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}

	const maxFinds = 5
	pickedNames := map[string]bool{
		"tiamat": true, // can't pick another Tiamat
	}
	var pickedIdxs []int

	for i, c := range seat.Library {
		if c == nil {
			continue
		}
		if !cardHasType(c, "creature") {
			continue
		}
		if !cardHasSubtype(c, "dragon") {
			continue
		}
		nameKey := strings.ToLower(strings.TrimSpace(c.DisplayName()))
		if pickedNames[nameKey] {
			continue
		}
		pickedNames[nameKey] = true
		pickedIdxs = append(pickedIdxs, i)
		if len(pickedIdxs) >= maxFinds {
			break
		}
	}

	foundNames := make([]string, 0, len(pickedIdxs))
	for _, idx := range pickedIdxs {
		card := seat.Library[idx]
		foundNames = append(foundNames, card.DisplayName())
	}
	// Move to hand. Walk indexes high-to-low so library positions stay valid.
	for i := len(pickedIdxs) - 1; i >= 0; i-- {
		idx := pickedIdxs[i]
		card := seat.Library[idx]
		gameengine.MoveCard(gs, card, seatIdx, "library", "hand", "tiamat_dragon_tutor")
	}
	shuffleLibraryPerCard(gs, seatIdx)

	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   seatIdx,
		Source: perm.Card.DisplayName(),
		Amount: len(foundNames),
		Details: map[string]interface{}{
			"slug":        slug,
			"found":       foundNames,
			"destination": "hand",
			"reason":      "tiamat_dragon_tutor",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  seatIdx,
		"found": foundNames,
		"count": len(foundNames),
	})
}
