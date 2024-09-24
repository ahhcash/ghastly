package search

import (
	"math"
)

func Cosine(vec1, vec2 []float64) float64 {
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}
