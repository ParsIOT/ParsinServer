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
	//glb.Debug.Println("####### enetered AddNodeByLabel ########")
	g.lock.Lock()
	n := Node{coords}
	g.nodes = append(g.nodes, &n)
	g.lock.Unlock()
	//glb.Debug.Println("####### exited AddNodeByLabel ########")
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

func (g *Graph) AddEdgeByLabel(n1, n2 string) {
	if g.edges == nil {
		g.edges = make(map[Node][]*Node)
	}
	n1Node,_ := g.GetNodeByLabel(n1)
	n2Node,_ := g.GetNodeByLabel(n2)

	g.lock.Lock()
	g.edges[*n1Node] = append(g.edges[*n1Node], n2Node)
	g.edges[*n2Node] = append(g.edges[*n2Node], n1Node)
	//todo: handle the repetetive entering the same edge
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


func (g *Graph) GetGraphMap() map[string][]string {
	//g.lock.RLock()
	graphMap := make(map[string][]string)
	g.AddNodeByLabel("10#10")
	g.AddNodeByLabel("20#20")
	g.AddNodeByLabel("20#30")
	g.AddNodeByLabel("40#40")
	g.AddNodeByLabel("50#50")
	g.AddEdgeByLabel("10#10", "20#20")
	g.AddEdgeByLabel("10#10", "20#30")
	g.AddEdgeByLabel("20#20", "10#10")
	g.AddEdgeByLabel("20#20", "20#30")
	g.AddEdgeByLabel("20#30", "10#10")
	g.AddEdgeByLabel("20#30", "20#20")
	g.AddEdgeByLabel("20#30", "50#50")
	g.AddEdgeByLabel("50#50", "20#30")
	//glb.Debug.Println("graphMap",graphMap)

	for i := 0; i < len(g.nodes); i++ {
		near := g.edges[*g.nodes[i]]
		for j := 0; j < len(near); j++ {
			graphMap[g.nodes[i].label] = append(graphMap[g.nodes[i].label], near[j].label)
		}
	}
	//glb.Debug.Println(graphMap)
	//g.lock.RUnlock()
	return graphMap
}