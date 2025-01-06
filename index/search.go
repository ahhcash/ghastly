package index

import (
	"sort"
	"sync"
)

// SearchResult represents a single search result
type SearchResult struct {
	ID       string
	Distance float64
	Vector   []float64
}

// SearchResults is a slice of search results that can be sorted
type SearchResults []SearchResult

func (s SearchResults) Len() int           { return len(s) }
func (s SearchResults) Less(i, j int) bool { return s[i].Distance < s[j].Distance }
func (s SearchResults) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Search finds k nearest neighbors for the given query vector
func (h *HNSW) Search(queryVector []float64, k int) (SearchResults, error) {
	return h.SearchWithAccuracy(queryVector, k, k*2)
}

// SearchWithAccuracy allows control over the search accuracy via ef parameter
func (h *HNSW) SearchWithAccuracy(queryVector []float64, k, ef int) (SearchResults, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	// Handle empty index
	if len(h.nodes) == 0 {
		return SearchResults{}, nil
	}

	// Start from entry point
	entryNode := h.nodes[h.entryPoint]
	currNode := entryNode
	currDist := h.distanceToNode(queryVector, entryNode.vector)

	// Search from top level to level 1
	for level := entryNode.maxLevel; level >= 1; level-- {
		// Greedy search within current level
		currNode, currDist = h.searchAtLayer(queryVector, currNode, currDist, level)
	}

	// Do a more thorough search at layer 0
	candidates := h.searchLayer(queryVector, currNode, 0, ef)

	// Convert candidates to SearchResults
	results := make(SearchResults, 0, len(candidates))
	for _, candidate := range candidates {
		if len(results) >= k {
			break
		}
		results = append(results, SearchResult{
			ID:       candidate.id,
			Distance: candidate.distance,
			Vector:   candidate.node.vector,
		})
	}

	// Sort results by distance
	sort.Sort(results)

	// Trim to k results if we have more
	if len(results) > k {
		results = results[:k]
	}

	return results, nil
}

// searchAtLayer performs a greedy search within a single layer
// returns the closest node found and its distance
func (h *HNSW) searchAtLayer(queryVector []float64, entryNode *node,
	entryDist float64, level int) (*node, float64) {

	currNode := entryNode
	currDist := entryDist

	for {
		changed := false

		// Check all neighbors at current level
		currNode.lock.RLock()
		neighbors := currNode.neighbors[level]
		currNode.lock.RUnlock()

		for _, neighborID := range neighbors {
			neighbor := h.nodes[neighborID]
			distance := h.distanceToNode(queryVector, neighbor.vector)

			// If we found a closer neighbor, move to it
			if distance < currDist {
				currDist = distance
				currNode = neighbor
				changed = true
				break
			}
		}

		// If we didn't find a closer neighbor, we're done at this level
		if !changed {
			break
		}
	}

	return currNode, currDist
}

// BatchSearch performs multiple searches in parallel
func (h *HNSW) BatchSearch(queryVectors [][]float64, k int) ([]SearchResults, error) {
	results := make([]SearchResults, len(queryVectors))
	errors := make([]error, len(queryVectors))

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(len(queryVectors))

	// Process each query vector in parallel
	for i, queryVector := range queryVectors {
		go func(idx int, query []float64) {
			defer wg.Done()
			results[idx], errors[idx] = h.Search(query, k)
		}(i, queryVector)
	}

	// Wait for all searches to complete
	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
