package gocfgmodule

import "sort"

func defaultRefreshDepends(
	regs []string,
	depends map[string][]string,
) []string {

	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	// 原始顺序索引（保证稳定）
	order := make(map[string]int)
	for i, r := range regs {
		order[r] = i
		inDegree[r] = 0
	}

	// 构建依赖图（严格按 regs 顺序）
	for _, key := range regs {
		for _, dep := range depends[key] {
			graph[dep] = append(graph[dep], key)
			inDegree[key]++
		}
	}

	// 初始入度为 0 的节点
	queue := make([]string, 0)
	for _, r := range regs {
		if inDegree[r] == 0 {
			queue = append(queue, r)
		}
	}

	sorted := make([]string, 0, len(regs))

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		sorted = append(sorted, curr)

		for _, next := range graph[curr] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)

				// 稳定排序
				sort.Slice(queue, func(i, j int) bool {
					return order[queue[i]] < order[queue[j]]
				})
			}
		}
	}

	if len(sorted) != len(regs) {
		panic("circular dependency detected")
	}

	return sorted
}
