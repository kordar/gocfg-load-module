package gocfgmodule

func (r *Registry) RefreshDepends(method RefreshDependsMethod) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if method == nil {
		r.regs = defaultRefreshDepends(r.regs, r.dependencies)
	} else {
		r.regs = method(r.regs, r.dependencies)
	}
}

func (r *Registry) Resolve(name string, setting interface{}) {
	if mod, ok := r.modules[name]; ok {
		mod.Load(setting)
	}
}

func (r *Registry) ResolveAll(settings map[string]interface{}) {
	// 1. BeforeLoad（顺序）
	for _, name := range r.regs {
		if mod, ok := r.modules[name]; ok {
			if bl, ok := mod.(GoCfgBeforeLoad); ok {
				bl.BeforeLoad()
			}
		}
	}

	// 2. Load（顺序）
	for _, name := range r.regs {
		if mod, ok := r.modules[name]; ok {
			if v, ok := settings[name]; ok {
				mod.Load(v)
			} else if r.required[name] {
				mod.Load(nil)
			}
		}
	}

	// 3. AfterLoad（顺序）
	for _, name := range r.regs {
		if mod, ok := r.modules[name]; ok {
			if al, ok := mod.(GoCfgAfterLoad); ok {
				al.AfterLoad()
			}
		}
	}
}

func (r *Registry) Destroy() {
	for i := len(r.regs) - 1; i >= 0; i-- {
		name := r.regs[i]
		if mod := r.modules[name]; mod != nil {
			mod.Close()
		}
	}
}
