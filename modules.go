package gocfgmodule

type GoCfgModule interface {
	Name() string
	Load(data interface{})
	Close()
}

var (
	modules = map[string]GoCfgModule{}
	regs    = make([]string, 0)
)

func Register(mod GoCfgModule) {
	if modules[mod.Name()] == nil {
		regs = append(regs, mod.Name())
	}
	modules[mod.Name()] = mod
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
