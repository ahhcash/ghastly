package index

import (
	"container/heap"
	"math"
)

// distanceFunc represents a function that calculates distance between two vectors
// nolint
type distanceFunc func([]float64, []float64) float64

// neighborSet represents a priority queue of potential neighbors
// nolint
type neighborSet struct {
	nodes     []*queueItem
	distances []float64
	maxSize   int
}

// queueItem represents a node in our priority queue
type queueItem struct {
	node     *node
	distance float64
	id       string
}

// distQueue is a priority queue for nearest neighbor candidates
type distQueue []*queueItem

func (pq *distQueue) Len() int {
	return len(*pq)
}

func (pq *distQueue) Less(i, j int) bool {
	return (*pq)[i].distance < (*pq)[j].distance
}

func (pq *distQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
}

func (pq *distQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*queueItem))
}

func (pq *distQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// selectNeighbors implements the Neighborhood Selection algorithm
// It selects the best M neighbors from the candidate set
func (h *HNSW) selectNeighbors(candidates []*queueItem, M int, keepPrunedConnections bool) []*queueItem {
	// If we have fewer candidates than M, return all candidates
	if len(candidates) <= M {
		return candidates
	}

	// Initialize result set and working queue
	result := make([]*queueItem, 0, M)
	workingSet := make(distQueue, len(candidates))
	copy(workingSet, candidates)
	heap.Init(&workingSet)

	// Create a map to track selected nodes for efficient lookup
	selected := make(map[string]bool)

	for len(result) < M && workingSet.Len() > 0 {
		// Get the closest candidate
		candidate := heap.Pop(&workingSet).(*queueItem)

		// Skip if we've already selected this node
		if selected[candidate.id] {
			continue
		}

		// Add to result set
		result = append(result, candidate)
		selected[candidate.id] = true

		if !keepPrunedConnections {
			// Heuristic: If this node would create too many closely connected components,
			// skip it and try the next closest one
			tooClose := false
			for _, existing := range result[:len(result)-1] {
				if h.isDistanceTooClose(candidate.node.vector, existing.node.vector) {
					tooClose = true
					break
				}
			}
			if tooClose {
				result = result[:len(result)-1]
				selected[candidate.id] = false
			}
		}
	}

	return result
}

// isDistanceTooClose checks if two vectors are too close to each other
// This helps maintain diversity in connections
func (h *HNSW) isDistanceTooClose(vec1, vec2 []float64) bool {
	// Calculate cosine similarity between vectors
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	similarity := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))

	// Consider vectors too close if similarity is above threshold
	// This threshold can be tuned based on your needs
	return similarity > 0.95
}

// connectNodes establishes bidirectional connections between nodes
func (h *HNSW) connectNodes(node1 *node, node2 *node, level int) {
	// Acquire locks for both nodes to prevent deadlocks
	if node1.id < node2.id {
		node1.lock.Lock()
		node2.lock.Lock()
	} else {
		node2.lock.Lock()
		node1.lock.Lock()
	}

	defer node1.lock.Unlock()
	defer node2.lock.Unlock()

	// Add bidirectional connections
	node1.neighbors[level] = append(node1.neighbors[level], node2.id)
	node2.neighbors[level] = append(node2.neighbors[level], node1.id)

	// Ensure we don't exceed maximum connections at this level
	if len(node1.neighbors[level]) > h.config.M {
		// Select best M neighbors
		candidates := make([]*queueItem, 0, len(node1.neighbors[level]))
		for _, neighborID := range node1.neighbors[level] {
			neighbor := h.nodes[neighborID]
			dist := h.distanceToNode(node1.vector, neighbor.vector)
			candidates = append(candidates, &queueItem{
				node:     neighbor,
				distance: dist,
				id:       neighborID,
			})
		}

		selected := h.selectNeighbors(candidates, h.config.M, false)

		// Update connections
		newNeighbors := make([]string, len(selected))
		for i, item := range selected {
			newNeighbors[i] = item.id
		}
		node1.neighbors[level] = newNeighbors
	}

	// Do the same for node2
	if len(node2.neighbors[level]) > h.config.M {
		candidates := make([]*queueItem, 0, len(node2.neighbors[level]))
		for _, neighborID := range node2.neighbors[level] {
			neighbor := h.nodes[neighborID]
			dist := h.distanceToNode(node2.vector, neighbor.vector)
			candidates = append(candidates, &queueItem{
				node:     neighbor,
				distance: dist,
				id:       neighborID,
			})
		}

		selected := h.selectNeighbors(candidates, h.config.M, false)

		newNeighbors := make([]string, len(selected))
		for i, item := range selected {
			newNeighbors[i] = item.id
		}
		node2.neighbors[level] = newNeighbors
	}
}

// distanceToNode calculates the distance between two vectors
// Currently using Euclidean distance, but this could be made configurable
func (h *HNSW) distanceToNode(vec1, vec2 []float64) float64 {
	sum := 0.0
	for i := range vec1 {
		diff := vec1[i] - vec2[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}
