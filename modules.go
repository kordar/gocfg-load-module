package gocfgmodule

import "fmt"

func Register(mod GoCfgModule, depends ...string) {
	if mod.Name() == "" {
		fmt.Println("module name is empty")
		return
	}
	if modules[mod.Name()] == nil {
		regs = append(regs, mod.Name())
	}
	modules[mod.Name()] = mod
	// 判断对象是否实现了depends接口
	if v, ok := mod.(GoCfgDepends); ok {
		if v.Depends() != nil && len(v.Depends()) > 0 {
			dependency[mod.Name()] = v.Depends()
		}
	}
	if len(depends) > 0 {
		for _, depend := range depends {
			if mod.Name() != depend {
				dependency[mod.Name()] = append(dependency[mod.Name()], depend)
			}
		}
	}

}

func RegisterWithRequired(mod GoCfgModule, required bool) {
	Register(mod)
	AddRequired(mod.Name())
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

func RefreshDepends(method RefreshDependsMethod) {
	if method == nil {
		regs = defaultRefreshDependsMethod(regs, dependency)
	} else {
		regs = method(regs, dependency)
	}
}

func ResolveAll(settings map[string]interface{}) {
	for _, name := range regs {
		if settings[name] != nil {
			Resolve(name, settings[name])
		} else if IsRequired(name) {
			Resolve(name, nil)
		}
	}
}
