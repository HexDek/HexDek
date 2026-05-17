package per_card

// dev/era2-unification registration entry point.
//
// We can't edit registry.go's registerDefaults() in this branch, so the
// hand-written commander handlers introduced for the Era 2 audit
// (Commander Masters, Lost Caverns of Ixalan, Wilds of Eldraine,
// March of the Machine, Phyrexia All Will Be One — roughly 2023–early
// 2024 sets) hook themselves into the global registry via this init().
//
// File name "zz_era2_register.go" sorts after every gen_* and existing
// custom file, so it runs LAST in the package's init order. Each
// register call below adds an OnTrigger / OnETB / OnActivated alongside
// any auto-generated stub for the same card; the dispatcher iterates
// every registered handler so this stacks rather than replaces.

func init() {
	RegisterEra2(Global())
	AddResetHook(RegisterEra2)
}

// RegisterEra2 re-runs the registrations on a given Registry. Tests
// that call Reset() (which wipes the global registry) should also
// call this to put the era 2 handlers back. registry.go's
// registerDefaults() rebuilds the gen_/existing handlers; init() above
// doesn't re-fire, so we expose a re-registration entry point for
// completeness.
func RegisterEra2(r *Registry) {
	if r == nil {
		return
	}
	registerSliverGravemotherCustom(r)
	registerYennaRedtoothRegentCustom(r)
	registerAmaliaBenavidesCustom(r)
	registerSaheeliRadiantCreatorCustom(r)
	registerKardurDoomscourgeCustom(r)
	registerFelotharSteadfastCustom(r)
	registerMondrakGloryDominus(r)
	registerSolphimMayhemDominus(r)
	registerDrivnodCarnageDominus(r)
	registerZopandrelHungerDominus(r)
	registerIxhelScionOfAtraxa(r)
}
