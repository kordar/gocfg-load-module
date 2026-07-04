package fx

import (
	"slices"
	"testing"

	uberfx "go.uber.org/fx"
)

type testModule struct {
	name      string
	events    *[]string
	loads     *[]string
	loadData  *[]any
	depends   []string
	index     int
	optionNum int
}

type testIndexModule struct {
	*testModule
	idx int
}

func (m *testModule) Name() string {
	return m.name
}

func (m *testModule) Depends() []string {
	return m.depends
}

func (m *testIndexModule) Index() int {
	return m.idx
}

func (m *testModule) BeforeLoad() {
	*m.events = append(*m.events, "before:"+m.name)
}

func (m *testModule) Load(data any) []uberfx.Option {
	*m.events = append(*m.events, "load:"+m.name)
	*m.loads = append(*m.loads, m.name)
	*m.loadData = append(*m.loadData, data)

	options := make([]uberfx.Option, 0, m.optionNum)
	for i := 0; i < m.optionNum; i++ {
		options = append(options, uberfx.Supply(m.name))
	}
	return options
}

func (m *testModule) AfterLoad() {
	*m.events = append(*m.events, "after:"+m.name)
}

func TestRegistryResolveAllOrdersAndRequired(t *testing.T) {
	t.Parallel()

	var events []string
	var loads []string
	var loadData []any

	registry := New()
	registry.Register(&testModule{
		name:      "db",
		events:    &events,
		loads:     &loads,
		loadData:  &loadData,
		optionNum: 1,
	})
	registry.Register(&testModule{
		name:      "cache",
		events:    &events,
		loads:     &loads,
		loadData:  &loadData,
		depends:   []string{"db"},
		optionNum: 1,
	})
	registry.RegisterRequired(&testModule{
		name:      "http",
		events:    &events,
		loads:     &loads,
		loadData:  &loadData,
		depends:   []string{"db", "cache"},
		optionNum: 1,
	})

	registry.RefreshDepends(nil)

	options := registry.ResolveAll(map[string]any{
		"db":    map[string]string{"dsn": "..."},
		"cache": map[string]string{"addr": "..."},
	})

	if len(options) != 3 {
		t.Fatalf("unexpected options count: %d", len(options))
	}

	expectedLoads := []string{"db", "cache", "http"}
	if !slices.Equal(loads, expectedLoads) {
		t.Fatalf("unexpected load order: %v", loads)
	}

	if loadData[2] != nil {
		t.Fatalf("required module should receive nil setting when missing, got %#v", loadData[2])
	}

	expectedEvents := []string{
		"before:db", "before:cache", "before:http",
		"load:db", "load:cache", "load:http",
		"after:db", "after:cache", "after:http",
	}
	if !slices.Equal(events, expectedEvents) {
		t.Fatalf("unexpected lifecycle order: %v", events)
	}
}

func TestRegistryResolveSingleModule(t *testing.T) {
	t.Parallel()

	var events []string
	var loads []string
	var loadData []any

	registry := New()
	registry.Register(&testModule{
		name:      "db",
		events:    &events,
		loads:     &loads,
		loadData:  &loadData,
		optionNum: 2,
	})

	options := registry.Resolve("db", "dsn")
	if len(options) != 2 {
		t.Fatalf("unexpected options count: %d", len(options))
	}

	if !slices.Equal(loads, []string{"db"}) {
		t.Fatalf("unexpected loads: %v", loads)
	}

	if len(loadData) != 1 || loadData[0] != "dsn" {
		t.Fatalf("unexpected load data: %#v", loadData)
	}

	expectedEvents := []string{"before:db", "load:db", "after:db"}
	if !slices.Equal(events, expectedEvents) {
		t.Fatalf("unexpected lifecycle order: %v", events)
	}
}

func TestRegistryRefreshDependsWithIndex(t *testing.T) {
	t.Parallel()

	var events []string
	var loads []string
	var loadData []any

	registry := New()

	// D: index=30, no deps
	// C: index=20, no deps → should load before D
	// B: index=10, depends on [E] → blocked until E
	// E: index=15, no deps → loads before B
	// A: index=0,  no deps → lowest index, loads first among root nodes
	//
	// Expected order:
	//   A(index=0) → C(index=20) → E(index=15) → D(index=30) → B(index=10, depends on E)
	//   But wait — E is at index 15, C is at index 20, so: A(0) → E(15) → C(20) → D(30) → B(10)
	//   No, dependencies take priority. B depends on E, so E must come before B regardless of index.
	//   Topological sort: indegree=0 initially: A, C, D, E
	//   Sorted by index: A(0) → E(15) → C(20) → D(30)
	//   After processing E, B becomes indegree=0
	//   Final: A → E → C → D → B
	//   Wait, B has index=10, which is lower than C(20) and D(30). But it was just added to queue.
	//   When B is added, the queue is [C(20), D(30)], then B(10) is appended and sorted.
	//   After sorting: [B(10), C(20), D(30)]
	//   So final order: A → E → B → C → D

	registry.RegisterRequired(&testIndexModule{
		testModule: &testModule{
			name:      "A",
			events:    &events,
			loads:     &loads,
			loadData:  &loadData,
			optionNum: 1,
		},
		idx: 0,
	})
	registry.RegisterRequired(&testIndexModule{
		testModule: &testModule{
			name:      "B",
			events:    &events,
			loads:     &loads,
			loadData:  &loadData,
			depends:   []string{"E"},
			optionNum: 1,
		},
		idx: 10,
	})
	registry.RegisterRequired(&testIndexModule{
		testModule: &testModule{
			name:      "C",
			events:    &events,
			loads:     &loads,
			loadData:  &loadData,
			optionNum: 1,
		},
		idx: 20,
	})
	registry.RegisterRequired(&testIndexModule{
		testModule: &testModule{
			name:      "D",
			events:    &events,
			loads:     &loads,
			loadData:  &loadData,
			optionNum: 1,
		},
		idx: 30,
	})
	registry.RegisterRequired(&testIndexModule{
		testModule: &testModule{
			name:      "E",
			events:    &events,
			loads:     &loads,
			loadData:  &loadData,
			optionNum: 1,
		},
		idx: 15,
	})
	// F has no Index → should fall back to registration order
	registry.RegisterRequired(&testModule{
		name:      "F",
		events:    &events,
		loads:     &loads,
		loadData:  &loadData,
		optionNum: 1,
	})

	registry.RefreshDepends(nil)

	// 触发 Load → 验证 ResolveAll 仍然走拓扑序
	registry.ResolveAll(map[string]any{})

	// Expected: A(0)→E(15)→B(10)→C(20)→D(30)→F(无Index, 注册序)
	expectedLoads := []string{"A", "E", "B", "C", "D", "F"}
	if !slices.Equal(loads, expectedLoads) {
		t.Fatalf("unexpected load order: %v (expected %v)", loads, expectedLoads)
	}
}
