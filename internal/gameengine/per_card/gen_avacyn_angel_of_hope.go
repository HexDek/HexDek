package per_card

// registerAvacynAngelOfHope is the auto-generated entry point retained
// so the registry's existing call site keeps compiling. The real
// implementation lives in custom_avacyn_angel_of_hope.go (the
// indestructible anthem grant). Delegating here avoids double
// registration while preserving the registry call.
func registerAvacynAngelOfHope(r *Registry) {
	registerAvacynAngelOfHopeCustom(r)
}
