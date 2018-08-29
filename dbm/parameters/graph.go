package parameters


import (
	"fmt"
	"sync"
	"errors"
	"strings"
	"strconv"
)


// Node a single node that composes the tree
type Node struct {
	label string
}

func (n *Node) String() string {
	//return fmt.Sprintf("%v", n.label)
	return n.label
}

// ItemGraph the Items graph
type Graph struct {
	nodes []*Node
	edges map[Node][]*Node
	lock  sync.RWMutex
}
func NewGraph() Graph {
	return Graph{
		nodes: 		[]*Node{},
		edges: 		make(map[Node][]*Node),
		lock:		sync.RWMutex{},
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(n *Node) {
	g.lock.Lock()
	g.nodes = append(g.nodes, n)
	g.lock.Unlock()
}

func (g *Graph) GetNodeByLabel (label string) (*Node,error){
	g.lock.Lock()
	for i := 0; i < len(g.nodes); i++ {
		if g.nodes[i].String() == label {
			g.lock.Unlock()
			return g.nodes[i],nil
		}
	}
	g.lock.Unlock()
	fmt.Printf("couldn't find specified node for %s",label)
	return nil, errors.New("not found")
}

// AddNode adds a node to the graph by getting string of coords
func (g *Graph) AddNodeByLabel(coords string) {
	g.lock.Lock()
	n := Node{coords}
	g.nodes = append(g.nodes, &n)
	g.lock.Unlock()
}

func (g *Graph) GetNodeLocation(coords string) (float64, float64){
	//n := g.GetNodeByLabel(coords)
	result := strings.Split(coords,"#")
	x,err := strconv.ParseFloat(result[0],64)
	if err!=nil {fmt.Println(err)}
	y,err := strconv.ParseFloat(result[1],64)
	return x,y
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(n1, n2 *Node) {
	g.lock.Lock()
	if g.edges == nil {
		g.edges = make(map[Node][]*Node)
	}
	g.edges[*n1] = append(g.edges[*n1], n2)
	g.edges[*n2] = append(g.edges[*n2], n1)
	g.lock.Unlock()
}


func (g *Graph) String() {
	g.lock.RLock()
	s := ""
	for i := 0; i < len(g.nodes); i++ {
		s += g.nodes[i].String() + " -> "
		near := g.edges[*g.nodes[i]]
		for j := 0; j < len(near); j++ {
			s += near[j].String() + " "
		}
		s += "\n"
	}
	fmt.Println(s)
	g.lock.RUnlock()
}