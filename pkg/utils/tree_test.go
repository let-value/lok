package utils

import (
	"sync"
	"testing"
)

func TestTraverse(t *testing.T) {
	root := &Node{
		Path: "/",
		Name: "root",
		Children: []*Node{
			{
				Path: "/child1",
				Name: "child1",
				Children: []*Node{
					{
						Path: "/child1/grandchild1",
						Name: "grandchild1",
						Children: []*Node{
							{
								Path: "/child1/grandchild1/greatgrandchild1",
								Name: "greatgrandchild1",
							},
						},
					},
				},
			},
			{
				Path: "/child2",
				Name: "child2",
			},
		},
	}

	processedOrder := make(map[string]int)
	nodes := make(map[string]*Node)
	var mu sync.Mutex
	var counter int

	processFunc := func(node *Node) {
		mu.Lock()
		processedOrder[node.Path] = counter
		nodes[node.Path] = node
		counter++
		mu.Unlock()
	}

	Traverse(root, processFunc)

	for path, order := range processedOrder {
		node := nodes[path]
		for _, child := range node.Children {
			if processedOrder[child.Path] >= order {
				t.Errorf("Child %s was processed after its parent %s", child.Path, path)
			}
		}
	}
}
