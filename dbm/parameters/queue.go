package parameters

import "sync"

/*
// Item the type of the queue
//type Item generic.Type

// ItemQueue the queue of Items
type Queue struct {
	items []Node
	lock  sync.RWMutex
}

// New creates a new ItemQueue
func (s *Queue) New() *Queue {
	s.items = []Node{}
	return s
}

// Enqueue adds an Item to the end of the queue
func (s *Queue) Enqueue(t Node) {
	s.lock.Lock()
	s.items = append(s.items, t)
	s.lock.Unlock()
}

// Dequeue removes an Item from the start of the queue
func (s *Queue) Dequeue() Node {
	s.lock.Lock()
	item := s.items[0]
	s.items = s.items[1:len(s.items)]
	s.lock.Unlock()
	return item // it was &item. I changed it to item. check if it works ok...
}

// Front returns the item next in the queue, without removing it
func (s *Queue) Front() Node {
	s.lock.RLock()
	item := s.items[0]
	s.lock.RUnlock()
	return item
}

// IsEmpty returns true if the queue is empty
func (s *Queue) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of Items in the queue
func (s *Queue) Size() int {
	return len(s.items)
}*/

// NodeQueue the queue of Nodes
type NodeQueue struct {
	items []Node
	lock  sync.RWMutex
}

// New creates a new NodeQueue
func (s *NodeQueue) New() *NodeQueue {
	s.lock.Lock()
	s.items = []Node{}
	s.lock.Unlock()
	return s
}

// Enqueue adds an Node to the end of the queue
func (s *NodeQueue) Enqueue(t Node) {
	s.lock.Lock()
	s.items = append(s.items, t)
	s.lock.Unlock()
}

// Dequeue removes an Node from the start of the queue
func (s *NodeQueue) Dequeue() *Node {
	s.lock.Lock()
	item := s.items[0]
	s.items = s.items[1:len(s.items)]
	s.lock.Unlock()
	return &item
}

// Front returns the item next in the queue, without removing it
func (s *NodeQueue) Front() *Node {
	s.lock.RLock()
	item := s.items[0]
	s.lock.RUnlock()
	return &item
}

// IsEmpty returns true if the queue is empty
func (s *NodeQueue) IsEmpty() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.items) == 0
}

// Size returns the number of Nodes in the queue
func (s *NodeQueue) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.items)
}
