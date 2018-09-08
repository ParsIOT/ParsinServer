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
	Label string	`json:"Label"`
}

func (n *Node) String() string {
	//return fmt.Sprintf("%v", n.label)
	return n.Label
}

// ItemGraph the Items graph
type Graph struct {
	Nodes []*Node			`json:"Nodes"`
	Edges map[Node][]*Node	`json:"Edges"`
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
	if err!=nil {glb.Error.Println(err)}
	y,err := strconv.ParseFloat(result[1],64)
	if err!=nil {glb.Error.Println(err)}
	return x,y
}

func ConvertStringLocToXY(coords string) (float64, float64){
	//n := g.GetNodeByLabel(coords)
	result := strings.Split(coords,",") // this function is for locations from hadi
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
			//glb.Debug.Printf("found specified node for %s and removed",n2Label)
			//g.lock.Unlock()
			// g.Nodes[i]
		}
	}
	for i := 0; i < len(g.Edges[*n2Node]); i++ {
		//glb.Debug.Println(g.Nodes[i].String())
		if g.Edges[*n2Node][i].String() == n1Label {
			g.Edges[*n2Node] = append(g.Edges[*n2Node][:i], g.Edges[*n2Node][i+1:]...)
			//glb.Debug.Printf("found specified node for %s and removed",n2Label)
			//g.lock.Unlock()
			// g.Nodes[i]
		}
	}
	g.lock.Unlock()
	//glb.Debug.Println("******** removing is done ********")
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
	//glb.Debug.Println("****** curX,curY:  ",curX,curY)
	minimumDist := math.MaxFloat64 // maybe should define a variable like the one hadi made for maxEucleadian distance
	var ownerOfMinimumDist *Node
	var curDist float64

	curDistants := []float64{}
	xys := []float64{}
	flag:=true
	flagenter := true
	//lenlen := len(g.Nodes)
/*	if lenlen != 2 {
		glb.Error.Println(lenlen)
	} else {
		glb.Debug.Println(lenlen)

	}*/
	//if lenlen==0{
	//	glb.Error.Println(" WTF  WTF  WTF  WTF  WTF  WTF  WTF  WTF  WTF  WTF  WTF  WTF  WTF ")
	//}

	for i := 0; i < len(g.Nodes); i++ {
		flagenter = false
		x,y := g.Nodes[i].GetNodeLocation()
		xys = append(xys,x)
		xys = append(xys,y)
		curDist = glb.CalcDist(curX,curY,x,y)
		curDistants = append(curDistants,curDist)
		if curDist<minimumDist{
			minimumDist = curDist
			ownerOfMinimumDist = g.Nodes[i]
			flag = false
		}
	}
	//glb.Debug.Println(graphMap)
	//g.lock.RUnlock()
	if ownerOfMinimumDist==nil {
		glb.Error.Println("**************** curX:",curX," curY:",curY," minDist:",minimumDist," flag:",flag," curDists:",curDistants," xys:",xys, " flagenter:",flagenter,"len(g.Nodes)",len(g.Nodes))
		glb.Error.Println("g.nodes:", g.Nodes)
	}
	//glb.Debug.Println("**************** curX:",curX," curY:",curY," minDist:",minimumDist," flag:",flag," curDists:",curDistants," xys:",xys, " flagenter:",flagenter,"len(g.Nodes)",len(g.Nodes))
	return ownerOfMinimumDist
}


// Traverse implements the BFS traversing algorithm
//func (g *Graph) BFSTraverse(startNode *Node, f func(*Node)) {
func (g *Graph) BFSTraverse(startNode *Node) [][]*Node {
	g.lock.RLock()
	q := NodeQueue{}
	q.New()
	//n := g.Nodes[0]
	n := startNode
	q.Enqueue(*n)
	//visited := make(map[*Node]bool)
	visited := make(map[string]bool)
	k:=1
	flag := true
	//result := make(map[int][]*Node)
	result := [][]*Node{}
	result = append(result,[]*Node{})
	//result = append(result,[]*Node{})
	//result = append(result,[]*Node{})
	result[0] = append(result[0],startNode)
	//glb.Debug.Println("result: ",result)

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
				if k>=len(result){
					//glb.Debug.Println("k is bigger than len")
					result = append(result,[]*Node{})
				}
				//glb.Debug.Println("result: ",result)
				result[k] = append(result[k],j)
				q.Enqueue(*j)
				visited[j.Label] = true
				flag = true
			}else {
				flag = false
			}
		}
		if flag {
			k++
			flag = true
		}
	}
	g.lock.RUnlock()
	//glb.Debug.Println(result)
	return result
}