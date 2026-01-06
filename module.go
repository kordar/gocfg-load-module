package gocfgmodule

// 模块基础接口
type GoCfgModule interface {
	Name() string
	Load(data interface{})
	Close()
}

// 依赖接口（可选）
type Depends interface {
	Depends() []string
}

// 生命周期钩子（可选）

type GoCfgBeforeLoad interface {
	BeforeLoad()
}

type GoCfgAfterLoad interface {
	AfterLoad()
}

// 刷新依赖方法
type RefreshDependsMethod func(
	regs []string,
	depends map[string][]string,
) []string
