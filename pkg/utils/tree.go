package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

type Node struct {
	Path     string
	Name     string
	Parent   *Node
	Children []*Node
}

func generatePathParts(path string) []string {
	var parts []string

	parts = append(parts, path)

	for path != "/" && path != "." {
		path = filepath.Dir(path)
		parts = append(parts, path)
	}

	parts = append(parts, "")

	return parts
}

func BuildTree(files []string) *Node {
	root := &Node{
		Path:     "",
		Name:     "",
		Parent:   nil,
		Children: make([]*Node, 0),
	}

	var nodeMap = make(map[string]*Node)

	nodeMap[""] = root

	for _, path := range files {
		path := filepath.Clean(path)
		name := filepath.Base(path)

		parts := generatePathParts(path)

		parentNode := root

		for _, part := range parts {
			if foundNode, ok := nodeMap[part]; ok {
				parentNode = foundNode
				break
			}
		}

		node := &Node{
			Path:     path,
			Name:     name,
			Parent:   parentNode,
			Children: make([]*Node, 0),
		}

		parentNode.Children = append(parentNode.Children, node)

		nodeMap[path] = node
	}

	return root
}

func PrintTree(root *Node, depth int) {
	indent := strings.Repeat(" ", depth*2)
	fmt.Println(indent, "|", root.Path)
	fmt.Println(indent, "|", root.Name)

	for _, child := range root.Children {
		PrintTree(child, depth+1)
	}
}

func Traverse(node *Node, processFunc func(*Node)) {
	if node == nil {
		return
	}

	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup

		for _, child := range node.Children {
			wg.Add(1)
			go func(child *Node) {
				defer wg.Done()
				Traverse(child, processFunc)
			}(child)
		}

		wg.Wait()

		processFunc(node)

		done <- true
	}()

	<-done
}
