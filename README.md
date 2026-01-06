# gocfg-load-module

A lightweight, stable, lifecycle-aware module registry for Go.

`gocfg-load-module` 是一个 **显式依赖、顺序稳定、无反射** 的 Go 模块加载与生命周期管理库，适用于：

- 业务模块初始化
- 插件 / 扩展系统
- Service / Agent 启动流程
- 需要**可预测加载顺序**的工程

---

## ✨ 特性

- ✅ **稳定拓扑排序**（保持注册顺序）
- ✅ 显式模块依赖（无反射）
- ✅ 完整生命周期管理
  - `BeforeLoad`
  - `Load`
  - `AfterLoad`
  - `Close`
- ✅ 支持 Required 模块
- ✅ 支持全局模式 & Registry 模式
- ✅ 无全局锁争用、无 magic
- ✅ 易测试、易维护

---

## 📦 安装

```bash
go get github.com/kordar/gocfg-load-module
```



## 🚀 快速上手（全局模式）

### 1️⃣ 定义模块

```go
type DBModule struct{}

func (d *DBModule) Name() string { return "db" }

func (d *DBModule) Load(cfg interface{}) error {
	fmt.Println("db load", cfg)
	return nil
}

func (d *DBModule) Close() error {
	fmt.Println("db close")
	return nil
}
```

------

### 2️⃣ 注册模块 & 依赖

```go
gocfg.Register(&DBModule{})
gocfg.Register(&CacheModule{}, "db")
gocfg.RegisterWithRequired(&HTTPModule{}, true)
```

------

### 3️⃣ 解析依赖 & 加载

```go
gocfg.RefreshDepends(nil)

gocfg.ResolveAll(map[string]interface{}{
	"db": map[string]string{"dsn": "..."},
})

defer gocfg.Destroy()
```

------

## 🔄 生命周期钩子（可选）

模块可按需实现以下接口：

```go
type BeforeLoad interface {
	BeforeLoad() error
}

type AfterLoad interface {
	AfterLoad() error
}
```

### 执行顺序

```
BeforeLoad (正序)
Load       (正序)
AfterLoad  (正序)
Close      (逆序)
```

------

## 🔗 模块依赖

### 显式依赖（推荐）

```go
gocfg.Register(&CacheModule{}, "db")
```

### 接口依赖（可选）

```go
func (c *CacheModule) Depends() []string {
	return []string{"db"}
}
```

------

## ⚠️ Required 模块

Required 模块 **必须被 Load**，且 **不允许缺失配置**：

```go
gocfg.RegisterWithRequired(&HTTPModule{}, true)
```

如果 `ResolveAll` 时未提供配置，将直接返回 error / panic（全局模式）。

------

## 🧩 Registry 模式（推荐用于测试 / 多实例）

```go
reg := gocfg.New()

reg.Register(&DBModule{})
reg.Register(&CacheModule{}, "db")

reg.RefreshDepends(nil)
reg.ResolveAll(nil)
defer reg.Destroy()
```

------

## 🧪 单元测试示例

```go
func TestLifecycleOrder(t *testing.T) {
	cfg := gocfg.New()

	cfg.Register(&DBModule{})
	cfg.Register(&CacheModule{}, "db")

	cfg.RefreshDepends(nil)
	cfg.ResolveAll(nil)
	cfg.Destroy()
}
```

------

## 🆚 与 DI / IoC 框架的区别

| 对比项   | gocfg       | dig / fx / wire |
| -------- | ----------- | --------------- |
| 依赖方式 | 显式声明    | 自动推导        |
| 加载顺序 | 可预测      | 间接            |
| 生命周期 | 模块级      | 构造函数级      |
| 反射     | ❌ 无        | ✅ 有            |
| 适合场景 | 模块 / 插件 | 服务装配        |

------

## ❓ 什么时候该用 gocfg

- ✅ 你关心 **初始化顺序**
- ✅ 你不想引入反射 / 代码生成
- ✅ 你希望模块是「黑盒」
- ❌ 你只想做类型自动注入（推荐 dig / wire）

------

## 🛣 Roadmap

-  Context 支持
-  并行加载（DAG 层）
-  依赖图导出（DOT）
-  可选模块 / Lazy Load
-  Plugin 动态加载

------

## 📄 License

MIT License
