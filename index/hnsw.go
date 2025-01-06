package index

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"sync"
)

type HNSWConfig struct {
	// node degree per layer
	M int

	// the max level in the hierarchy
	MaxLevel int

	// the inverse of the probability of promoting a node to a higher level
	LevelMult float64

	// claude wtf???
	EfConstruction int
}

func DefaultHNSWConfig() HNSWConfig {
	return HNSWConfig{
		M:              16,
		MaxLevel:       16,
		LevelMult:      1 / math.Log(2),
		EfConstruction: 128,
	}
}

type node struct {
	// vector data
	vector []float64

	// unique identifier for the node
	id string

	// maxLevel is the maximum level this node exists in
	maxLevel int

	// neighbors[level] is a slice of neighbor IDs at that level
	neighbors [][]string

	lock sync.RWMutex
}

func newNode(data []float64, id string, maxLevel int, maxConnections int) *node {
	neighbors := make([][]string, maxLevel+1)
	for i := range neighbors {
		neighbors[i] = make([]string, maxConnections)
	}

	return &node{
		vector:    data,
		id:        id,
		maxLevel:  maxLevel,
		neighbors: neighbors,
	}
}

type HNSW struct {
	// map of nodes with nodeID keys
	nodes map[string]*node

	// config is the config params declared in HNSWConfig
	config HNSWConfig

	// the entryPoint (node at the highest level) to start our searches
	entryPoint string

	// global lock
	lock sync.RWMutex
}

func NewHNSW(config HNSWConfig) *HNSW {
	return &HNSW{
		nodes:  make(map[string]*node),
		config: config,
	}
}

func (h *HNSW) randomLevel() int {
	level := 0
	for rand.Float64() < 1/h.config.LevelMult && level < h.config.MaxLevel {
		level++
	}

	return level
}

func (h *HNSW) Insert(id string, vector []float64) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	level := h.randomLevel()
	newnode := newNode(vector, id, level, h.config.M)

	if len(h.nodes) == 0 {
		h.nodes[id] = newnode
		h.entryPoint = id
		return nil
	}
	entryNode := h.nodes[h.entryPoint]

	currNode := entryNode
	currDist := h.distanceToNode(vector, entryNode.vector)

	for lc := entryNode.maxLevel; lc > level; lc-- {
		changed := true
		for changed {
			changed = false

			currNode.lock.RLock()
			neighbors := currNode.neighbors[lc]
			currNode.lock.RUnlock()

			for _, neighborID := range neighbors {
				neighbor := h.nodes[neighborID]
				distance := h.distanceToNode(vector, neighbor.vector)

				if distance < currDist {
					currDist = distance
					currNode = neighbor
					changed = true
					break
				}
			}
		}
	}

	for lc := 0; lc <= level; lc++ {
		candidates := h.searchLayer(vector, currNode, lc, h.config.EfConstruction)

		selectedNeighbors := h.selectNeighbors(candidates, h.config.M, false)

		for _, neighbor := range selectedNeighbors {
			h.connectNodes(newnode, neighbor.node, lc)
		}
	}

	h.nodes[id] = newnode

	if level > h.nodes[h.entryPoint].maxLevel {
		h.entryPoint = id
	}

	return nil
}

func (h *HNSW) searchLayer(queryVector []float64, entryNode *node, level int, ef int) []*queueItem {
	visited := make(map[string]bool)
	visited[entryNode.id] = true

	candidates := make(distQueue, 0)
	heap.Init(&candidates)

	results := make(distQueue, 0)
	heap.Init(&results)

	startDist := h.distanceToNode(queryVector, entryNode.vector)
	item := &queueItem{node: entryNode, distance: startDist, id: entryNode.id}
	heap.Push(&candidates, item)
	heap.Push(&results, item)

	for candidates.Len() > 0 && candidates[0].distance <= results[results.Len()-1].distance {
		current := heap.Pop(&candidates).(*queueItem)

		current.node.lock.RLock()
		neighbors := current.node.neighbors[level]
		current.node.lock.RUnlock()

		for _, neighborID := range neighbors {
			if visited[neighborID] {
				continue
			}

			visited[neighborID] = true
			neighbor := h.nodes[neighborID]
			distance := h.distanceToNode(queryVector, neighbor.vector)

			if results.Len() < ef || distance < results[results.Len()-1].distance {
				item := &queueItem{node: neighbor, distance: distance, id: neighborID}
				heap.Push(&candidates, item)
				heap.Push(&results, item)

				if results.Len() > ef {
					heap.Pop(&results)
				}
			}
		}
	}

	resultSlice := make([]*queueItem, results.Len())
	for i := len(resultSlice) - 1; i >= 0; i-- {
		resultSlice[i] = heap.Pop(&results).(*queueItem)
	}

	return resultSlice
}

// Delete removes a node with the given ID from the index
func (h *HNSW) Delete(id string) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Check if the node exists
	node, exists := h.nodes[id]
	if !exists {
		return fmt.Errorf("node with id %s does not exist", id)
	}

	// If this is the entry point, we need to find a new one
	if h.entryPoint == id {
		h.updateEntryPointForDeletion(id)
	}

	// Remove references to this node from all its neighbors
	for level := 0; level <= node.maxLevel; level++ {
		node.lock.RLock()
		neighbors := node.neighbors[level]
		node.lock.RUnlock()

		for _, neighborID := range neighbors {
			if neighborID == "" {
				continue
			}
			neighbor := h.nodes[neighborID]
			h.removeNeighborConnection(neighbor, id, level)
		}
	}

	// Delete the node from our nodes map
	delete(h.nodes, id)

	return nil
}

// updateEntryPointForDeletion finds a new entry point when the current one is being deleted
func (h *HNSW) updateEntryPointForDeletion(deletingID string) {
	if len(h.nodes) == 1 {
		// If this is the last node, clear the entry point
		h.entryPoint = ""
		return
	}

	// Find the node with the highest level that isn't the one being deleted
	maxLevel := -1
	var newEntryID string

	for nodeID, node := range h.nodes {
		if nodeID != deletingID && node.maxLevel > maxLevel {
			maxLevel = node.maxLevel
			newEntryID = nodeID
		}
	}

	h.entryPoint = newEntryID
}

// removeNeighborConnection removes a connection to the specified node at the given level
func (h *HNSW) removeNeighborConnection(node *node, targetID string, level int) {
	node.lock.Lock()
	defer node.lock.Unlock()

	for i, neighborID := range node.neighbors[level] {
		if neighborID == targetID {
			if i == len(node.neighbors[level])-1 {
				node.neighbors[level][i] = ""
			} else {
				copy(node.neighbors[level][i:], node.neighbors[level][i+1:])
				node.neighbors[level][len(node.neighbors[level])-1] = ""
			}
			break
		}
	}
}
