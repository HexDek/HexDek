package per_card

// dev/activated-stub-handlers registration entry point.
//
// Wires the six custom activated-ability handlers introduced in this
// branch (Phenax, Obeka, Jhoira, Shadowheart, Splinter, Ghen) onto the
// global registry. The four other commanders from the priority batch
// (Mayael, Yawgmoth, Charix, Tazri) already had custom handlers in
// previous branches and are not re-registered here.
//
// The "zz_" prefix runs after registry.go's init() so the global
// registry is fully constructed by the time we add to it. Each
// register* call below appends to the (cardName, hook) handler list;
// the dispatcher iterates every registered handler so this stacks
// alongside the auto-generated stubs rather than replacing them.

func init() {
	RegisterActivatedStubsBatch1(Global())
	AddResetHook(RegisterActivatedStubsBatch1)
}

// RegisterActivatedStubsBatch1 re-runs the registrations on a given
// Registry. Tests that call Reset() (which wipes the global registry)
// should also call this to put these handlers back, mirroring the
// pattern in zz_handler_coverage_2_register.go.
func RegisterActivatedStubsBatch1(r *Registry) {
	if r == nil {
		return
	}
	registerPhenaxGodOfDeceptionCustom(r)
	registerObekaBruteChronologistCustom(r)
	registerJhoiraAgelessInnovatorCustom(r)
	registerShadowheartDarkJusticiarCustom(r)
	registerSplinterRadicalRatCustom(r)
	registerGhenArcanumWeaverCustom(r)
}
