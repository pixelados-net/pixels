package path

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// nodeKey stores a pathfinding node identity.
type nodeKey struct {
	// X stores the node x coordinate.
	X uint16

	// Y stores the node y coordinate.
	Y uint16

	// Z stores the node section height.
	Z grid.Height
}

// nodeFromPosition creates a node key from a position.
func nodeFromPosition(position Position) nodeKey {
	return nodeKey{X: position.Point.X, Y: position.Point.Y, Z: position.Z}
}

// point returns the node tile coordinate.
func (key nodeKey) point() grid.Point {
	return grid.Point{X: key.X, Y: key.Y}
}

// position returns the node position.
func (key nodeKey) position() Position {
	return Position{Point: key.point(), Z: key.Z}
}

// openNode stores one open-set entry.
type openNode struct {
	// key stores the node identity.
	key nodeKey

	// cost stores the accumulated movement cost.
	cost int

	// priority stores the heap priority.
	priority int
}

// direction stores one neighbor direction.
type direction struct {
	// dx stores x movement.
	dx int

	// dy stores y movement.
	dy int

	// diagonal reports whether the direction is diagonal.
	diagonal bool
}

var directions = [...]direction{
	{dx: 0, dy: -1},
	{dx: 1, dy: 0},
	{dx: 0, dy: 1},
	{dx: -1, dy: 0},
	{dx: 1, dy: -1, diagonal: true},
	{dx: 1, dy: 1, diagonal: true},
	{dx: -1, dy: 1, diagonal: true},
	{dx: -1, dy: -1, diagonal: true},
}

// openHeap stores open nodes ordered by priority.
type openHeap struct {
	// nodes stores heap entries.
	nodes []openNode
}

// Len returns the heap length.
func (heap openHeap) Len() int {
	return len(heap.nodes)
}

// Push adds a node to the heap.
func (heap *openHeap) Push(node openNode) {
	heap.nodes = append(heap.nodes, node)
	heap.up(len(heap.nodes) - 1)
}

// Pop removes the lowest-priority node.
func (heap *openHeap) Pop() (openNode, bool) {
	if len(heap.nodes) == 0 {
		return openNode{}, false
	}

	node := heap.nodes[0]
	last := heap.nodes[len(heap.nodes)-1]
	heap.nodes = heap.nodes[:len(heap.nodes)-1]
	if len(heap.nodes) > 0 {
		heap.nodes[0] = last
		heap.down(0)
	}

	return node, true
}

// up moves a node toward the heap root.
func (heap *openHeap) up(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if heap.less(parent, index) {
			return
		}
		heap.swap(parent, index)
		index = parent
	}
}

// down moves a node away from the heap root.
func (heap *openHeap) down(index int) {
	for {
		left := index*2 + 1
		right := left + 1
		smallest := index
		if left < len(heap.nodes) && heap.less(left, smallest) {
			smallest = left
		}
		if right < len(heap.nodes) && heap.less(right, smallest) {
			smallest = right
		}
		if smallest == index {
			return
		}
		heap.swap(index, smallest)
		index = smallest
	}
}

// less reports whether left should sort before right.
func (heap openHeap) less(left int, right int) bool {
	leftNode := heap.nodes[left]
	rightNode := heap.nodes[right]
	if leftNode.priority == rightNode.priority {
		return leftNode.cost < rightNode.cost
	}

	return leftNode.priority < rightNode.priority
}

// swap swaps two heap entries.
func (heap *openHeap) swap(left int, right int) {
	heap.nodes[left], heap.nodes[right] = heap.nodes[right], heap.nodes[left]
}

// neighborPoint returns a neighboring point.
func neighborPoint(current nodeKey, direction direction) (grid.Point, bool) {
	x := int(current.X) + direction.dx
	y := int(current.Y) + direction.dy

	return grid.NewPoint(x, y)
}
