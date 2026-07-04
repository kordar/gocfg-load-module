# gocfg-load-module

配置驱动的模块加载与生命周期管理，提供**显式依赖、稳定排序、无反射**的模块编排能力。适用于业务模块初始化、插件系统、服务启动流程等需要**可预测加载顺序**的场景。

本仓库包含两个包：

| 包 | 路径 | 适用场景 |
|---|------|---------|
| 基础包 | `github.com/kordar/gocfg-load-module` | 通用模块生命周期，不依赖 DI 框架 |
| Fx 包 | `github.com/kordar/gocfg-load-module/fx/v2` | 对接 uber-go/fx，`Load` 返回 `[]fx.Option` |

---

## 核心概念

```
Register → RefreshDepends → ResolveAll → Destroy
   │            │               │           │
  注册模块   拓扑排序 + Index   按序加载    逆序回收
```

### 模块接口

```go
// 基础接口（通用包）
type GoCfgModule interface {
    Name() string
    Load(data interface{})
    Close()
}

// Fx 接口（fx/v2 包）—— Load 返回 fx.Option
type GoCfgModule interface {
    Name() string
    Load(data any) []fx.Option
}
```

### 可选接口

```go
// 声明依赖（被依赖的模块先加载）
type Depends interface {
    Depends() []string
}

// 生命周期钩子
type GoCfgBeforeLoad interface { BeforeLoad() }
type GoCfgAfterLoad interface  { AfterLoad() }

// 优先级控制（fx/v2）—— 拓扑同级时 Index 越小越先执行
type GoCfgIndex interface {
    Index() int
}
```

### 执行顺序

```
BeforeLoad (正序) → Load (正序) → AfterLoad (正序)
关闭时 Close 逆序执行
```

---

## 排序规则

1. **拓扑依赖优先** — 被依赖的模块先加载
2. **Index 同级排序** — 同一批可执行模块中实现 `GoCfgIndex` 的按 `Index()` 升序排列
3. **注册顺序兜底** — 都没有 Index 的按注册顺序

```
注册: A(index=10) → B(depends A, index=5) → C(index=0) → D → E(depends D)
排序: C(0) → A(10) → B(5) → D → E
       ↑        ↑        ↑       ↑跟注册顺序
     同层 Index 小的先  依赖 A 先满足后再执行
```

---

## 基础包用法

### 安装

```bash
go get github.com/kordar/gocfg-load-module
```

### 定义模块

```go
import "github.com/kordar/gocfg-load-module"

type DBModule struct{}

func (d *DBModule) Name() string                 { return "db" }
func (d *DBModule) Load(data interface{})        { fmt.Println("db loaded") }
func (d *DBModule) Close()                       { fmt.Println("db closed") }
```

### 注册并加载

```go
// 全局模式
gocfgmodule.Register(&DBModule{})
gocfgmodule.Register(&CacheModule{}, "db")          // Cache 依赖 DB
gocfgmodule.RegisterRequired(&HTTPModule{}, true)   // HTTP 必须加载

gocfgmodule.RefreshDepends(nil)
gocfgmodule.ResolveAll(map[string]interface{}{
    "db":    map[string]string{"dsn": "root:@tcp(127.0.0.1:3306)/test"},
    "cache": map[string]string{"addr": "127.0.0.1:6379"},
})
defer gocfgmodule.Destroy()
```

### Registry 模式（推荐用于测试 / 多实例）

```go
reg := gocfgmodule.New()
reg.Register(&DBModule{})
reg.Register(&CacheModule{}, "db")
reg.RefreshDepends(nil)
reg.ResolveAll(settings)
reg.Destroy()
```

---

## Fx 包用法（fx/v2）

### 安装

```bash
go get github.com/kordar/gocfg-load-module/fx/v2
```

### 定义模块

```go
import gocfgmodulefx "github.com/kordar/gocfg-load-module/fx/v2"

type cfgModule struct {
    name  string
    index int
}

func (m cfgModule) Name() string             { return m.name }
func (m cfgModule) Index() int                { return m.index }
func (m cfgModule) Load(data any) []fx.Option {
    return []fx.Option{
        fx.Module("my-module",
            fx.Provide(NewService),
        ),
    }
}
```

### 注册到 Fx App

```go
func AdminServerStarter() *fx.App {
    return ServerWithStarter(
        []gocfgmodulefx.GoCfgModule{
            mypkg.StarterModule("gologger", mypkg.WithIndex(-3)),
            mypkg.StarterModule("mysql",    mypkg.WithIndex(-1)),
            mypkg.StarterModule("myapp",    mypkg.WithIndex(2)),
        },
        []string{"myapp"}, // required 模块列表
        []string{"scheduler"}, // 依赖列表
        options...,
    )
}
```

### 自定义排序策略

```go
gocfgmodulefx.RefreshDepends(func(regs []string, depends map[string][]string) []string {
    // 自定义拓扑排序逻辑
    return customSort(regs, depends)
})
```

---

## API 参考

### Registry

| 方法 | 说明 |
|------|------|
| `Register(mod, depends...)` | 注册模块及显式依赖 |
| `RegisterRequired(mod, depends...)` | 注册必须模块（配置缺失也执行） |
| `AddRequired(names...)` | 标记已注册模块为 required |
| `RefreshDepends(method)` | 执行拓扑排序，nil 用默认算法 |
| `Resolve(name, setting)` | 加载单个模块 |
| `ResolveAll(settings)` | 按序加载全部模块 |
| `Destroy()` | 清空注册表 |

### 全局函数（两个包均提供）

`Register` / `RegisterRequired` / `AddRequired` / `RefreshDepends` / `Resolve` / `ResolveAll` / `Destroy`

---

## 与 DI 框架对比

| | gocfg | dig / fx / wire |
|---|---|---|
| 依赖方式 | 显式声明 | 自动推导 |
| 加载顺序 | 可预测（Index + Topo） | 间接 |
| 生命周期 | 模块级 | 构造函数级 |
| 反射 | 无 | 有 |
| 适合场景 | 模块 / 插件编排 | 服务装配 |

---

## License

MIT
