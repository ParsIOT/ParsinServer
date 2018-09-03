package parameters


import (
	"fmt"
	"sync"
	"errors"
	"strings"
	"strconv"
	"ParsinServer/glb"
	"math"
)


// Node a single node that composes the tree
type Node struct {
	Label string
}

func (n *Node) String() string {
	//return fmt.Sprintf("%v", n.label)
	return n.Label
}

// ItemGraph the Items graph
type Graph struct {
	Nodes []*Node
	Edges map[Node][]*Node
	lock  sync.RWMutex
}
func NewGraph() Graph {
	return Graph{
		Nodes: 		[]*Node{},
		Edges: 		make(map[Node][]*Node),
		lock:		sync.RWMutex{},
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(n *Node) {
	g.lock.Lock()
	g.Nodes = append(g.Nodes, n)
	g.lock.Unlock()
}

func (g *Graph) GetNodeByLabel (label string) (*Node,error){
	g.lock.Lock()
	for i := 0; i < len(g.Nodes); i++ {
		if g.Nodes[i].String() == label {
			g.lock.Unlock()
			return g.Nodes[i],nil
		}
	}
	g.lock.Unlock()
	fmt.Printf("couldn't find specified node for %s",label)
	return nil, errors.New("not found")
}

func (g *Graph) RemoveNodeByLabel (label string){
	g.lock.Lock()
	glb.Debug.Println("you see ",label)
	for i := 0; i < len(g.Nodes); i++ {
		glb.Debug.Println(g.Nodes[i].String())
		if g.Nodes[i].String() == label {
			g.Nodes = append(g.Nodes[:i], g.Nodes[i+1:]...)
			glb.Debug.Printf("found specified node for %s and removed",label)
			//g.lock.Unlock()
			// g.Nodes[i]
		}
	}
	g.lock.Unlock()
	//glb.Debug.Printf("couldn't find specified node for %s",label)

}

// AddNode adds a node to the graph by getting string of coords
func (g *Graph) AddNodeByLabel(coords string) {
	//glb.Debug.Println("####### enetered AddNodeByLabel ########")
	g.lock.Lock()
	//glb.Debug.Println("******** it is about to add to the nodes ********",coords)
	n := Node{coords}
	g.Nodes = append(g.Nodes, &n)
	//glb.Debug.Println("********  added to the nodes ********",g.Nodes)
	g.lock.Unlock()
	//glb.Debug.Println("####### exited AddNodeByLabel ########")
}

func (node *Node) GetNodeLocation() (float64, float64){
	//n := g.GetNodeByLabel(coords)
	coords := node.Label
	result := strings.Split(coords,"#")
	x,err := strconv.ParseFloat(result[0],64)
	if err!=nil {fmt.Println(err)}
	y,err := strconv.ParseFloat(result[1],64)
	return x,y
}

func ConvertStringLocToXY(coords string) (float64, float64){
	//n := g.GetNodeByLabel(coords)
	result := strings.Split(coords,"#") // this function is for locations from hadi
	x,err := strconv.ParseFloat(result[0],64)
	if err!=nil {fmt.Println(err)}
	y,err := strconv.ParseFloat(result[1],64)
	return x,y
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(n1, n2 *Node) {
	g.lock.Lock()
	if g.Edges == nil {
		g.Edges = make(map[Node][]*Node)
	}
	g.Edges[*n1] = append(g.Edges[*n1], n2)
	g.Edges[*n2] = append(g.Edges[*n2], n1)
	g.lock.Unlock()
}

func (g *Graph) AddEdgeByLabel(n1, n2 string) {
	if g.Edges == nil {
		g.Edges = make(map[Node][]*Node)
	}
	n1Node,_ := g.GetNodeByLabel(n1)
	n2Node,_ := g.GetNodeByLabel(n2)
	flag := true
	for _, b := range g.Edges[*n1Node] {
		if b == n2Node {
			flag = false
		}
	}
	if flag==true {
		g.lock.Lock()
		//glb.Debug.Println("******** it is about to add to the edeges ********")
		g.Edges[*n1Node] = append(g.Edges[*n1Node], n2Node)
		g.Edges[*n2Node] = append(g.Edges[*n2Node], n1Node)
		g.lock.Unlock()
	}
}

func (g *Graph) RemoveEdgeByLabel(n string) {
	result := strings.Split(n,"&")
	n1Label := result[0]
	n2Label := result[1]
	n1Node,_ := g.GetNodeByLabel(n1Label)
	n2Node,_ := g.GetNodeByLabel(n2Label)
	g.lock.Lock()
	for i := 0; i < len(g.Edges[*n1Node]); i++ {
		//glb.Debug.Println(g.Nodes[i].String())
		if g.Edges[*n1Node][i].String() == n2Label {
			g.Edges[*n1Node] = append(g.Edges[*n1Node][:i], g.Edges[*n1Node][i+1:]...)
			glb.Debug.Printf("found specified node for %s and removed",n2Label)
			//g.lock.Unlock()
			// g.Nodes[i]
		}
	}
	for i := 0; i < len(g.Edges[*n2Node]); i++ {
		//glb.Debug.Println(g.Nodes[i].String())
		if g.Edges[*n2Node][i].String() == n1Label {
			g.Edges[*n2Node] = append(g.Edges[*n2Node][:i], g.Edges[*n2Node][i+1:]...)
			glb.Debug.Printf("found specified node for %s and removed",n2Label)
			//g.lock.Unlock()
			// g.Nodes[i]
		}
	}
	g.lock.Unlock()
	glb.Debug.Println("******** removing is done ********")
}

func (g *Graph) String() {
	g.lock.RLock()
	s := ""
	for i := 0; i < len(g.Nodes); i++ {
		s += g.Nodes[i].String() + " -> "
		near := g.Edges[*g.Nodes[i]]
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
	//for k := range graphMap {
	//	delete(graphMap, k)
	//}
	//g.AddNodeByLabel("10#10")
	//g.AddNodeByLabel("20#20")
	//g.AddNodeByLabel("20#30")
	//g.AddNodeByLabel("40#40")
	//g.AddNodeByLabel("50#50")
	//g.AddEdgeByLabel("10#10", "20#20")
	//g.AddEdgeByLabel("10#10", "20#30")
	//g.AddEdgeByLabel("20#20", "10#10")
	//g.AddEdgeByLabel("20#20", "20#30")
	//g.AddEdgeByLabel("20#30", "10#10")
	//g.AddEdgeByLabel("20#30", "20#20")
	//g.AddEdgeByLabel("20#30", "50#50")
	//g.AddEdgeByLabel("50#50", "20#30")
	//glb.Debug.Println("graphMap",graphMap)

	for i := 0; i < len(g.Nodes); i++ {
		near := g.Edges[*g.Nodes[i]]
		graphMap[g.Nodes[i].Label] = []string{}
		for j := 0; j < len(near); j++ {
			graphMap[g.Nodes[i].Label] = append(graphMap[g.Nodes[i].Label], near[j].Label)
		}
	}
	//glb.Debug.Println(graphMap)
	//g.lock.RUnlock()
	return graphMap
}

func (g *Graph) DeleteGraph() {
	g.lock.Lock()
	for k := range g.Edges {
		delete(g.Edges, k)
	}
	//g.Nodes = []*Node{}
	g.Nodes = g.Nodes[0:0]
	g.lock.Unlock()
	glb.Debug.Println("exiting from deleteGraph")
	glb.Debug.Println(g.Nodes)
}



func (g *Graph) GetNearestNode(location string) *Node {
	//g.lock.RLock()
	curX, curY := ConvertStringLocToXY(location)
	minimumDist := math.MaxFloat64 // maybe should define a variable like the one hadi made for maxEucleadian distance
	var ownerOfMinimumDist *Node
	var curDist float64
	for i := 0; i < len(g.Nodes); i++ {
		x,y := g.Nodes[i].GetNodeLocation()
		curDist = glb.CalcDist(curX,curY,x,y)
		if curDist<minimumDist{
			minimumDist = curDist
			ownerOfMinimumDist = g.Nodes[i]
		}
	}
	//glb.Debug.Println(graphMap)
	//g.lock.RUnlock()
	return ownerOfMinimumDist
}


// Traverse implements the BFS traversing algorithm
func (g *Graph) BFSTraverse(startNode *Node, f func(*Node)) {
	g.lock.RLock()
	q := NodeQueue{}
	q.New()
	//n := g.Nodes[0]
	n := startNode
	q.Enqueue(*n)
	//visited := make(map[*Node]bool)
	visited := make(map[string]bool)
	//i:=0
	for {
		//glb.Debug.Println("visited is ",visited,"in ",i,"th step")
		if q.IsEmpty() {
			break
		}
		node := q.Dequeue()
		visited[node.Label] = true
		near := g.Edges[*node]

		for i := 0; i < len(near); i++ {
			j := near[i]
			if !visited[j.Label] {
				q.Enqueue(*j)
				visited[j.Label] = true
			}
		}
		if f != nil {
			f(node)
		}
	}
	g.lock.RUnlock()
}