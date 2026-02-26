package interpreter

func registerRuntimeBuiltins() {
	registerRuntimeCoreBuiltins()
	registerRuntimeUtilityBuiltins()
	registerRuntimeSystemBuiltins()
	registerProcessBuiltins()
}
