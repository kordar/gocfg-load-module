package gocfgmodule

func IsRequired(name string) bool {
	for _, v := range requirements {
		if v == name {
			return true
		}
	}
	return false
}

func removeDuplicatesRequired() {
	uniqueMap := make(map[string]bool) // 记录元素是否出现
	result := []string{}

	for _, v := range requirements {
		if !uniqueMap[v] { // 如果该元素没出现过
			uniqueMap[v] = true
			result = append(result, v) // 添加到结果切片
		}
	}
	requirements = result
}

func AddRequired(name ...string) {
	requirements = append(requirements, name...)
	removeDuplicatesRequired()
}

var defaultRefreshDependsMethod RefreshDependsMethod = func(regs []string, depends map[string][]string) []string {

	//fmt.Println("---------", regs)
	//fmt.Println("---------", depends)
	inDegree := make(map[string]int) // 记录入度
	graph := make(map[string][]string)

	// 初始化图
	for _, reg := range regs {
		inDegree[reg] = 0 // 初始化入度为 0
	}

	// 构建依赖关系
	for key, deps := range depends {
		for _, dep := range deps {
			graph[dep] = append(graph[dep], key) // dep -> key 表示 key 依赖 dep
			inDegree[key]++                      // 被依赖者入度 +1
		}
	}

	// 使用队列存储入度为 0 的节点
	queue := []string{}
	for _, reg := range regs {
		if inDegree[reg] == 0 {
			queue = append(queue, reg)
		}
	}

	var sorted []string

	// 进行拓扑排序
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		sorted = append(sorted, curr)

		// 遍历当前节点指向的所有依赖项
		for _, next := range graph[curr] {
			inDegree[next]-- // 依赖项入度 -1
			if inDegree[next] == 0 {
				queue = append(queue, next) // 如果入度为 0，加入队列
			}
		}
	}

	// 如果排序后的元素个数和 regs 不符，说明有循环依赖
	if len(sorted) != len(regs) {
		panic("Error: Circular dependency detected!")
		return nil
	}

	//fmt.Println("---------", sorted)

	return sorted
}
