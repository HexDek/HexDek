package main

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// TestHasCastConditionalETB exercises the detector against the AST shape
// emitted by the Python parser for "When ~ enters, if you cast it, ..."
// triggered abilities (Modification at effect position → ModificationEffect
// with kind="conditional_effect").
func TestHasCastConditionalETB(t *testing.T) {
	mk := func(rawCond string) *gameast.CardAST {
		return &gameast.CardAST{
			Name: "Test Card",
			Abilities: []gameast.Ability{
				&gameast.Triggered{
					Trigger: gameast.Trigger{Event: "etb"},
					Effect: &gameast.ModificationEffect{
						ModKind: "conditional_effect",
						Args:    []interface{}{rawCond},
					},
				},
			},
		}
	}

	cases := []struct {
		name      string
		raw       string
		wantCast  bool
		wantHand  bool
		wantInv   bool
	}{
		{"the one ring", "if you cast it, you gain protection from everything until your next turn", true, false, false},
		{"tiamat", "if you cast it, search your library for up to five dragon cards", true, false, false},
		{"cyclone summoner from hand", "if you cast it from your hand, return all other nonland permanents to hand", true, true, false},
		{"preston inverse", "if it wasn't cast, create a token that's a copy of that creature", false, false, true},
		{"non-cast conditional", "if you control a swamp, draw a card", false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cast, hand, inv := hasCastConditionalETB(mk(tc.raw))
			if cast != tc.wantCast || hand != tc.wantHand || inv != tc.wantInv {
				t.Fatalf("hasCastConditionalETB(%q) = (cast=%v hand=%v inv=%v); want (%v %v %v)",
					tc.raw, cast, hand, inv, tc.wantCast, tc.wantHand, tc.wantInv)
			}
		})
	}
}
