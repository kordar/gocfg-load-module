package fx

import (
	"sort"
	"sync"

	"go.uber.org/fx"
)

// GoCfgModule 定义 Fx 配置模块，Load 返回要挂载的 fx.Option。
type GoCfgModule interface {
	Name() string
	Load(data any) []fx.Option
}

// Depends 定义模块依赖关系。
type Depends interface {
	Depends() []string
}

// GoCfgBeforeLoad 在 Load 前执行。
type GoCfgBeforeLoad interface {
	BeforeLoad(data any)
}

// GoCfgAfterLoad 在 Load 后执行。
type GoCfgAfterLoad interface {
	AfterLoad(data any)
}

// GoCfgIndex 返回模块执行优先级。
// 值越小越先执行，最终顺序 = 拓扑依赖(优先) + Index(同层排序)。
type GoCfgIndex interface {
	Index() int
}

// RefreshDependsMethod 自定义依赖刷新策略。
type RefreshDependsMethod func(regs []string, depends map[string][]string) []string

// Registry 维护模块注册、依赖顺序和解析结果。
type Registry struct {
	mu sync.Mutex

	modules      map[string]GoCfgModule
	regs         []string
	required     map[string]bool
	dependencies map[string][]string
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

	if d, ok := mod.(Depends); ok {
		r.dependencies[name] = appendUnique(r.dependencies[name], d.Depends()...)
	}
	r.dependencies[name] = appendUnique(r.dependencies[name], dependsWithoutSelf(name, depends)...)
}

func (r *Registry) RegisterRequired(mod GoCfgModule, depends ...string) {
	r.Register(mod, depends...)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.required[mod.Name()] = true
}

func (r *Registry) AddRequired(names ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, name := range names {
		r.required[name] = true
	}
}

func (r *Registry) IsRequired(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.required[name]
}

func (r *Registry) RefreshDepends(method RefreshDependsMethod) {
	r.mu.Lock()
	defer r.mu.Unlock()

	indices := make(map[string]int, len(r.modules))
	for _, name := range r.regs {
		if mod, ok := r.modules[name]; ok {
			if idx, ok := mod.(GoCfgIndex); ok {
				indices[name] = idx.Index()
			}
		}
	}

	if method == nil {
		r.regs = defaultRefreshDepends(r.regs, r.dependencies, indices)
		return
	}
	r.regs = method(r.regs, cloneDependencies(r.dependencies))
}

func (r *Registry) Resolve(name string, setting any) []fx.Option {
	r.mu.Lock()
	mod, ok := r.modules[name]
	r.mu.Unlock()
	if !ok {
		return nil
	}

	if bl, ok := mod.(GoCfgBeforeLoad); ok {
		bl.BeforeLoad(setting)
	}

	options := mod.Load(setting)

	if al, ok := mod.(GoCfgAfterLoad); ok {
		al.AfterLoad(setting)
	}

	return append([]fx.Option(nil), options...)
}

func (r *Registry) ResolveAll(settings map[string]any) []fx.Option {
	regs, modules, required := r.snapshot()

	options := make([]fx.Option, 0)

	for _, name := range regs {
		if mod, ok := modules[name]; ok {
			if bl, ok := mod.(GoCfgBeforeLoad); ok {
				bl.BeforeLoad(settings[name])
			}
		}
	}

	for _, name := range regs {
		mod, ok := modules[name]
		if !ok {
			continue
		}

		setting, exists := settings[name]
		if !exists && !required[name] {
			continue
		}
		options = append(options, mod.Load(setting)...)
	}

	for _, name := range regs {
		if mod, ok := modules[name]; ok {
			if al, ok := mod.(GoCfgAfterLoad); ok {
				al.AfterLoad(settings[name])
			}
		}
	}

	return options
}

func (r *Registry) Destroy() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.modules = make(map[string]GoCfgModule)
	r.regs = make([]string, 0)
	r.required = make(map[string]bool)
	r.dependencies = make(map[string][]string)
}

func (r *Registry) snapshot() ([]string, map[string]GoCfgModule, map[string]bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	regs := append([]string(nil), r.regs...)
	modules := make(map[string]GoCfgModule, len(r.modules))
	for k, v := range r.modules {
		modules[k] = v
	}
	required := make(map[string]bool, len(r.required))
	for k, v := range r.required {
		required[k] = v
	}
	return regs, modules, required
}

func defaultRefreshDepends(regs []string, depends map[string][]string, indices map[string]int) []string {
	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	order := make(map[string]int, len(regs))
	for i, name := range regs {
		order[name] = i
		inDegree[name] = 0
	}

	for _, name := range regs {
		for _, dep := range depends[name] {
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
		}
	}

	queue := make([]string, 0)
	for _, name := range regs {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}
	// 初始队列：Index 小 → 先执行
	sort.Slice(queue, func(i, j int) bool {
		return indexThenOrder(indices, order, queue[i], queue[j])
	})

	sorted := make([]string, 0, len(regs))
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		sorted = append(sorted, curr)

		for _, next := range graph[curr] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
				sort.Slice(queue, func(i, j int) bool {
					return indexThenOrder(indices, order, queue[i], queue[j])
				})
			}
		}
	}

	if len(sorted) != len(regs) {
		panic("circular dependency detected")
	}

	return sorted
}

// indexThenOrder 先按 Index 升序，Index 相同则按注册顺序。
func indexThenOrder(indices map[string]int, order map[string]int, a, b string) bool {
	ia, aok := indices[a]
	ib, bok := indices[b]

	// 只有双方都设置了 Index 时才比较
	if aok && bok {
		if ia != ib {
			return ia < ib
		}
	}
	return order[a] < order[b]
}

func cloneDependencies(src map[string][]string) map[string][]string {
	dst := make(map[string][]string, len(src))
	for k, v := range src {
		dst[k] = append([]string(nil), v...)
	}
	return dst
}

func appendUnique(dst []string, values ...string) []string {
	seen := make(map[string]struct{}, len(dst))
	for _, item := range dst {
		seen[item] = struct{}{}
	}

	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		dst = append(dst, value)
		seen[value] = struct{}{}
	}

	return dst
}

func dependsWithoutSelf(name string, depends []string) []string {
	filtered := make([]string, 0, len(depends))
	for _, dep := range depends {
		if dep != name {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

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

func RegisterRequired(mod GoCfgModule, depends ...string) {
	getDefaultRegistry().RegisterRequired(mod, depends...)
}

func AddRequired(names ...string) {
	getDefaultRegistry().AddRequired(names...)
}

func IsRequired(name string) bool {
	return getDefaultRegistry().IsRequired(name)
}

func RefreshDepends(method RefreshDependsMethod) {
	getDefaultRegistry().RefreshDepends(method)
}

func Resolve(name string, setting any) []fx.Option {
	return getDefaultRegistry().Resolve(name, setting)
}

func ResolveAll(settings map[string]any) []fx.Option {
	return getDefaultRegistry().ResolveAll(settings)
}

func Destroy() {
	getDefaultRegistry().Destroy()
}
