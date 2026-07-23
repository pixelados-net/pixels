// Package banzai implements server-authored Battle Banzai tiles.
package banzai

// Tile stores one compact Banzai tile state.
type Tile struct {
	// Team stores color one through four, or zero when unowned.
	Team uint8
	// Progress stores zero through two; two is permanently locked.
	Progress uint8
}

// Locked reports whether the tile can no longer be stolen.
func (tile Tile) Locked() bool { return tile.Team != 0 && tile.Progress == 2 }

// State returns Nitro's multi-state furniture value.
func (tile Tile) State() int {
	if tile.Team == 0 {
		return 0
	}
	return int(tile.Team)*3 + int(tile.Progress)
}

// Step applies one authoritative player step and reports points and a new lock.
func (tile *Tile) Step(team uint8, stealPoints int, lockPoints int) (int, bool) {
	if team < 1 || team > 4 || tile.Locked() {
		return 0, false
	}
	if tile.Team != team {
		stolen := tile.Team != 0
		tile.Team, tile.Progress = team, 0
		if stolen {
			return stealPoints, false
		}
		return 0, false
	}
	if tile.Progress < 2 {
		tile.Progress++
	}
	if tile.Progress == 2 {
		return lockPoints, true
	}
	return 0, false
}

// Board stores one rectangular tile arena and reusable flood-fill scratch.
type Board struct {
	// Width stores the column count.
	Width int
	// Height stores the row count.
	Height int
	// Tiles stores row-major tile state.
	Tiles []Tile
	// visited reuses generation markers across captures.
	visited []uint32
	// generation identifies one traversal without clearing visited.
	generation uint32
	// queue reuses traversal storage.
	queue []int
	// candidate reuses enclosed-region storage.
	candidate []int
	// best reuses the largest enclosed-region storage.
	best []int
}

// NewBoard creates one bounded empty arena.
func NewBoard(width int, height int) *Board {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	size := width * height
	return &Board{Width: width, Height: height, Tiles: make([]Tile, size), visited: make([]uint32, size), queue: make([]int, 0, size), candidate: make([]int, 0, size), best: make([]int, 0, size)}
}

// CaptureLargest locks the largest enemy or empty region fully enclosed by a team.
func (board *Board) CaptureLargest(team uint8) []int {
	board.nextGeneration()
	board.best = board.best[:0]
	for start, tile := range board.Tiles {
		if tile.Team == team || board.visited[start] == board.generation {
			continue
		}
		board.candidate, board.queue = board.candidate[:0], board.queue[:0]
		board.queue = append(board.queue, start)
		board.visited[start] = board.generation
		touchesEdge := false
		for len(board.queue) > 0 {
			index := board.queue[len(board.queue)-1]
			board.queue = board.queue[:len(board.queue)-1]
			board.candidate = append(board.candidate, index)
			x, y := index%board.Width, index/board.Width
			if x == 0 || y == 0 || x == board.Width-1 || y == board.Height-1 {
				touchesEdge = true
			}
			board.visitNeighbor(team, x-1, y)
			board.visitNeighbor(team, x+1, y)
			board.visitNeighbor(team, x, y-1)
			board.visitNeighbor(team, x, y+1)
		}
		if !touchesEdge && len(board.candidate) > len(board.best) {
			board.best = append(board.best[:0], board.candidate...)
		}
	}
	for _, index := range board.best {
		board.Tiles[index] = Tile{Team: team, Progress: 2}
	}
	return board.best
}

// Complete reports whether every arena tile is permanently locked.
func (board *Board) Complete() bool {
	if len(board.Tiles) == 0 {
		return false
	}
	for _, tile := range board.Tiles {
		if !tile.Locked() {
			return false
		}
	}
	return true
}

// visitNeighbor queues one traversable region tile.
func (board *Board) visitNeighbor(team uint8, x int, y int) {
	if x < 0 || y < 0 || x >= board.Width || y >= board.Height {
		return
	}
	index := y*board.Width + x
	if board.visited[index] == board.generation || board.Tiles[index].Team == team {
		return
	}
	board.visited[index] = board.generation
	board.queue = append(board.queue, index)
}

// nextGeneration advances scratch markers and safely handles wraparound.
func (board *Board) nextGeneration() {
	board.generation++
	if board.generation != 0 {
		return
	}
	clear(board.visited)
	board.generation = 1
}
