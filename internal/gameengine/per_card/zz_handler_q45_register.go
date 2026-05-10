package per_card

// dev/handler-q45 registration entry point.
//
// Six commanders from the Q45 stub list whose gen_*.go entries were
// pure emitPartial no-ops. Each new register* call below stacks
// alongside the auto-generated stub via the dispatcher's append-on-
// register list, so visible behavior flips from "partial" to whatever
// the custom handler resolves.
//
// The remaining 21 commanders in the Q45 batch are static-only
// abilities (Rakdos, Thrun, Uril, Kudo, Maha, Yorion's flying, etc.)
// where the AST keyword pipeline already handles the static. Their
// gen_*.go stubs correctly emitPartial; shipping parallel handlers
// that also emitPartial would not change behavior, so they are
// intentionally not included here. See the PR body for the breakdown.

func init() {
	r := Global()
	registerGyrudaDoomOfDepthsCustom(r)
	registerMorlunDevourerOfSpidersCustom(r)
	registerUreniTheSongUnendingCustom(r)
	registerEllieVengefulHunterCustom(r)
	registerYorionSkyNomadCustom(r)
	registerMisterNegativeCustom(r)
}

// RegisterHandlerQ45 re-runs the registrations on a given Registry.
// Tests that call Reset() (which wipes the global registry) should
// also call this to put these handlers back, mirroring the pattern in
// zz_handler_coverage_2_register.go.
func RegisterHandlerQ45(r *Registry) {
	if r == nil {
		return
	}
	registerGyrudaDoomOfDepthsCustom(r)
	registerMorlunDevourerOfSpidersCustom(r)
	registerUreniTheSongUnendingCustom(r)
	registerEllieVengefulHunterCustom(r)
	registerYorionSkyNomadCustom(r)
	registerMisterNegativeCustom(r)
}
