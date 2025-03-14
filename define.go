package gocfgmodule

// GoCfgModule 模块接口
type GoCfgModule interface {
	Name() string          // 名称
	Load(data interface{}) // 加载数据
	Close()                // 关闭模块
}

// GoCfgDepends 关联模块
type GoCfgDepends interface {
	Depends() []string // 关联模块
}

type SingletonLoad func(moduleName string, id string, cfg map[string]string)
type SingletonLoadI func(moduleName string, id string, cfg map[string]interface{})
type RefreshDependsMethod func(regs []string, depends map[string][]string) []string // 通过依赖关系重新计算模块加载顺序

var (
	// 关联模块
	modules      = map[string]GoCfgModule{}
	regs         = make([]string, 0)
	requirements = make([]string, 0) // Module必须实现
	dependency   = map[string][]string{}
)
