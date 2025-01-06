package gocfgmodule

type GoCfgModule interface {
	Name() string
	Load(data interface{})
	Close()
}

var (
	modules = map[string]GoCfgModule{}
	regs    = make([]string, 0)
	needs   = make(map[string]bool) // Module必须实现
)

func Register(mod GoCfgModule) {
	if modules[mod.Name()] == nil {
		regs = append(regs, mod.Name())
	}
	modules[mod.Name()] = mod
}

func SetNeeds(m map[string]bool) {
	needs = m
}

func Register2(mod GoCfgModule, need bool) {
	if modules[mod.Name()] == nil {
		regs = append(regs, mod.Name())
	}
	modules[mod.Name()] = mod
	needs[mod.Name()] = need
}

func Resolve(name string, setting interface{}) {
	if modules[name] != nil {
		modules[name].Load(setting)
	}
}

func Destroy() {
	count := len(regs)
	// TODO 后注册先销毁
	for i := count - 1; i >= 0; i-- {
		name := regs[i]
		modules[name].Close()
	}
}

func ResolveAll(settings map[string]interface{}) {
	for _, name := range regs {
		if settings[name] != nil {
			Resolve(name, settings[name])
		} else if needs[name] {
			Resolve(name, nil)
		}
	}
}
