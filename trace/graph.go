// Copyright (c) 2015 Datacratic. All rights reserved.

package trace

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// Graph represents a trace as interconnected contexts with their exit points.
type Graph struct {
	Nodes map[string]*Node

	root Node
}

// Node represents any node in the graph.
type Node struct {
	ID    int
	Count int64
	Name  string
	Path  string
	Enter map[string]*Node
	Leave map[string]*Node

	ratio float64
}

// HandleTrace updates the graph by keeping track of entering and leaving nodes.
func (graph *Graph) HandleTrace(events []Event) {
	if graph.Nodes == nil {
		graph.Nodes = make(map[string]*Node)
	}

	nodes := make([]*Node, len(events))
	nodes[0] = &graph.root

	for i, n := 1, len(events); i < n; i++ {
		item := &events[i]
		from := nodes[item.From]

		switch item.Kind {
		default:
		case StartEvent:
			node, ok := from.Enter[item.What]
			if !ok {
				node, ok = graph.Nodes[item.What]
				if !ok {
					node = &Node{
						Name: item.What,
						Path: item.What + ".",
					}

					graph.Nodes[item.What] = node
				}

				if from.Enter == nil {
					from.Enter = make(map[string]*Node)
				}

				from.Enter[item.What] = node
				if item.From == 0 {
					graph.Nodes[node.Name] = node
				}
			}

			nodes[i] = node
			node.Count++
		case EnterEvent:
			node, ok := from.Enter[item.What]
			if !ok {
				name := from.Path + item.What
				node = &Node{
					Name: name,
					Path: name + ".",
				}

				if from.Enter == nil {
					from.Enter = make(map[string]*Node)
				}

				from.Enter[item.What] = node
				if item.From == 0 {
					graph.Nodes[node.Name] = node
				}
			}

			nodes[i] = node
			node.Count++
		case LeaveEvent:
			node, ok := from.Leave[item.What]
			if !ok {
				node = &Node{
					Name: from.Path + item.What,
				}

				if from.Leave == nil {
					from.Leave = make(map[string]*Node)
				}

				from.Leave[item.What] = node
			}

			node.Count++
		}
	}
}

func (graph *Graph) Start() error {
	/*
		if graph.Resource != "" {
			http.Handle(graph.Resource, graph)
		}
	*/
	return nil
}

func (graph *Graph) Report(dt time.Duration) {
}

// DrawSVG exports the graph in SVG format.
func (graph *Graph) DrawSVG() (result []byte, err error) {
	dot, err := graph.DrawDOT()
	if err != nil {
		return
	}

	// invoke system's dot binary to get an output in SVG
	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = bytes.NewReader(dot)
	svg := bytes.Buffer{}
	cmd.Stdout = &svg
	if err = cmd.Run(); err != nil {
		return
	}

	result = svg.Bytes()
	return
}

// DrawDOT exports the graph in DOT format.
func (graph *Graph) DrawDOT() (result []byte, err error) {
	nodes := make(map[string]*Node)

	// declare explore to traverse recursively
	var explore func(node *Node)

	// create explorer
	explore = func(node *Node) {
		nodes[node.Name] = node

		process := func(items map[string]*Node) {
			for _, child := range items {
				if node.Count == 0 {
					child.ratio = 1.0
				} else {
					child.ratio = float64(child.Count) / float64(node.Count)
				}

				explore(child)
			}
		}

		process(node.Enter)
		process(node.Leave)
	}

	// traverse graph
	for _, node := range graph.Nodes {
		explore(node)
	}

	lines := bytes.Buffer{}
	label := bytes.Buffer{}

	// create a graphical representation of the graph in DOT format
	dot := bufio.NewWriter(&lines)
	fmt.Fprintf(dot, "digraph Trace {\n")

	// write nodes
	count := 0
	for _, node := range nodes {
		if node.ratio == 0.0 || len(node.Leave) == 0 {
			continue
		}

		node.ID = count
		count++

		label.Reset()
		w := bufio.NewWriter(&label)

		fmt.Fprintf(w, "  n%d [shape=record,label=\"<f>%s\\n%.2f%%|{⏎|", node.ID, node.Name, node.ratio*100.0)
		for _, child := range node.Leave {
			fmt.Fprintf(w, "%s\\n%.2f%%|", child.Name, child.ratio*100.0)
		}

		w.Flush()
		label.Truncate(label.Len() - 1)

		fmt.Fprintf(dot, "%s}\"];\n", label.String())
	}

	// write edges
	for _, node := range nodes {
		if node.ratio == 0.0 || len(node.Leave) == 0 {
			continue
		}

		for _, child := range node.Enter {
			fmt.Fprintf(dot, "  n%d:f -> n%d:f;\n", node.ID, child.ID)
		}
	}

	dot.WriteString("}\n")
	dot.Flush()

	result = lines.Bytes()
	return
}
