package per_card

// registerFeatherTheRedeemed is the auto-generated entry point retained
// so the existing registry call keeps compiling. Real implementation
// lives in custom_feather_the_redeemed.go (exile-on-resolve + delayed
// return-to-hand at the next end step).
func registerFeatherTheRedeemed(r *Registry) {
	registerFeatherTheRedeemedCustom(r)
}
