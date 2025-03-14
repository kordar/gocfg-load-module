package gocfgmodule_test

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	regs := []string{"A", "B", "C", "D", "E"}
	depends := map[string][]string{
		"D": {"A", "B"}, // D 依赖 A, B
		"B": {},         // E 依赖 B, C
	}

	sorted := topologicalSort(regs, depends)
	fmt.Println(sorted) // 可能的输出: ["A", "B", "C", "D", "E"]
}

// 拓扑排序函数
func topologicalSort(regs []string, depends map[string][]string) []string {
	// 记录入度（每个节点被多少个其他节点依赖）
	inDegree := make(map[string]int)
	// 记录邻接表（从 A 指向 B）
	graph := make(map[string][]string)

	// 初始化入度
	for _, reg := range regs {
		inDegree[reg] = 0 // 确保所有节点都被初始化
	}

	// 构建依赖图和计算入度
	for node, dependencies := range depends {
		for _, dep := range dependencies {
			graph[dep] = append(graph[dep], node) // dep → node
			inDegree[node]++                      // node 的入度 +1
		}
	}

	// 使用队列存储入度为 0 的节点（优先处理）
	queue := []string{}
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// 结果存储
	var result []string

	// Kahn's Algorithm 拓扑排序
	for len(queue) > 0 {
		// 取出队列中的第一个元素
		curr := queue[0]
		queue = queue[1:] // 出队
		result = append(result, curr)

		// 遍历所有被 curr 依赖的节点，并减少它们的入度
		for _, neighbor := range graph[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 { // 如果入度变为 0，加入队列
				queue = append(queue, neighbor)
			}
		}
	}

	// 如果排序后的元素数量不等于原始数量，说明存在循环依赖
	if len(result) != len(regs) {
		return []string{} // 发现环，返回空
	}

	return result
}
