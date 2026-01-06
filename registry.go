package gocfgmodule

import "sync"

type Registry struct {
	mu sync.Mutex

	modules      map[string]GoCfgModule
	regs         []string
	required     map[string]bool
	dependencies map[string][]string

	refreshMethod RefreshDependsMethod
}

func New() *Registry {
	return &Registry{
		modules:      make(map[string]GoCfgModule),
		regs:         make([]string, 0),
		required:     make(map[string]bool),
		dependencies: make(map[string][]string),
	}
}

func (r *Registry) Register(mod GoCfgModule, depends ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := mod.Name()
	if name == "" {
		panic("module name is empty")
	}

	if _, exists := r.modules[name]; !exists {
		r.regs = append(r.regs, name)
	}

	r.modules[name] = mod

	// 接口依赖
	if d, ok := mod.(Depends); ok {
		r.dependencies[name] = append(r.dependencies[name], d.Depends()...)
	}

	// 显式依赖
	for _, dep := range depends {
		if dep != name {
			r.dependencies[name] = append(r.dependencies[name], dep)
		}
	}
}

func (r *Registry) RegisterRequired(mod GoCfgModule, depends ...string) {
	r.Register(mod, depends...)
	r.required[mod.Name()] = true
}
