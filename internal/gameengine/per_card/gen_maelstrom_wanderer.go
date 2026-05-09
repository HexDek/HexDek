package per_card

// registerMaelstromWanderer is the auto-generated entry point retained
// so the existing registry call keeps compiling. Real implementation
// lives in custom_maelstrom_wanderer.go (second-cascade ETB + creatures-
// have-haste anthem).
func registerMaelstromWanderer(r *Registry) {
	registerMaelstromWandererCustom(r)
}
