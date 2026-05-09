package per_card

// dev/handler-coverage-2 registration entry point.
//
// We can't edit registry.go's registerDefaults() in this branch, so the
// 13 hand-written commander handlers introduced for handler-coverage-2
// hook themselves into the global registry via this init(). It runs
// after registry.go's init() because Go runs same-package init()s in
// file lexical order and "zz_..." sorts after "registry.go".
//
// Each register call below adds an OnTrigger / OnETB / OnActivated
// alongside the auto-generated stub for the same card; multiple
// handlers per (card, event) coexist (the dispatcher iterates every
// registered handler), so this stacks rather than replaces. The
// generated stubs only emit a parser_gap event, so the visible behavior
// flips from "partial" to whatever the custom handler does.

func init() {
	r := Global()
	registerArcadesCustom(r)
	registerTifaLockhartCustom(r)
	registerVeyranCustom(r)
	registerShadowCustom(r)
	registerErietteCustom(r)
	registerInallaCustom(r)
	registerEddieBrockCustom(r)
	registerTiamatCustom(r)
	registerMayaelCustom(r)
	registerGiadaCustom(r)
	registerGhyrsonStarnCustom(r)
	registerChocoCustom(r)
	registerIsshinCustom(r)
}

// RegisterHandlerCoverage2 re-runs the registrations on a given Registry.
// Tests that call Reset() (which wipes the global registry) should also
// call this to put the dev/handler-coverage-2 handlers back. registry.go
// already calls registerDefaults() in Reset(), but our init() doesn't
// re-fire — so tests invoke this explicitly.
func RegisterHandlerCoverage2(r *Registry) {
	if r == nil {
		return
	}
	registerArcadesCustom(r)
	registerTifaLockhartCustom(r)
	registerVeyranCustom(r)
	registerShadowCustom(r)
	registerErietteCustom(r)
	registerInallaCustom(r)
	registerEddieBrockCustom(r)
	registerTiamatCustom(r)
	registerMayaelCustom(r)
	registerGiadaCustom(r)
	registerGhyrsonStarnCustom(r)
	registerChocoCustom(r)
	registerIsshinCustom(r)
}
