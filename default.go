package gocfgmodule

import "sync"

var (
	defaultOnce     sync.Once
	defaultRegistry *Registry
)

func getDefaultRegistry() *Registry {
	defaultOnce.Do(func() {
		defaultRegistry = New()
	})
	return defaultRegistry
}

func Register(mod GoCfgModule, depends ...string) {
	getDefaultRegistry().Register(mod, depends...)
}

func RegisterWithRequired(mod GoCfgModule, required bool) {
	getDefaultRegistry().Register(mod)
	if required {
		getDefaultRegistry().required[mod.Name()] = true
	}
}

func Resolve(name string, setting interface{}) {
	getDefaultRegistry().Resolve(name, setting)
}

func RefreshDepends(method RefreshDependsMethod) {
	getDefaultRegistry().RefreshDepends(method)
}

func ResolveAll(settings map[string]interface{}) {
	getDefaultRegistry().ResolveAll(settings)
}

func Destroy() {
	getDefaultRegistry().Destroy()
}

func AddRequired(name ...string) {
	reg := getDefaultRegistry()
	for _, n := range name {
		reg.required[n] = true
	}
}

func IsRequired(name string) bool {
	return getDefaultRegistry().required[name]
}
