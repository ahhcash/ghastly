package tests

import (
	"github.com/aakashshankar/vexdb/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math"
	"testing"
)

type SearchMetricsTestSuite struct {
	suite.Suite
	vec1           []float64
	vec2           []float64
	orthogonalVec1 []float64
	orthogonalVec2 []float64
	parallelVec1   []float64
	parallelVec2   []float64
}

func (s *SearchMetricsTestSuite) SetupTest() {
	s.vec1 = []float64{1.0, 2.0, 3.0}
	s.vec2 = []float64{4.0, 5.0, 6.0}

	s.orthogonalVec1 = []float64{1.0, 0.0}
	s.orthogonalVec2 = []float64{0.0, 1.0}

	s.parallelVec1 = []float64{2.0, 4.0, 6.0}
	s.parallelVec2 = []float64{4.0, 8.0, 12.0}
}

func (s *SearchMetricsTestSuite) TestCosine() {
	cosine := search.Cosine(s.vec1, s.vec2)
	expected := 0.9746318461970762 // Pre-calculated value
	assert.InDelta(s.T(), expected, cosine, 0.000001)

	cosine = search.Cosine(s.orthogonalVec1, s.orthogonalVec2)
	assert.InDelta(s.T(), 0.0, cosine, 0.000001)

	cosine = search.Cosine(s.parallelVec1, s.parallelVec2)
	assert.InDelta(s.T(), 1.0, cosine, 0.000001)

	cosine = search.Cosine(s.vec1, s.vec1)
	assert.InDelta(s.T(), 1.0, cosine, 0.000001)
}

func (s *SearchMetricsTestSuite) TestDot() {
	dot := search.Dot(s.vec1, s.vec2)
	expected := 32.0 // 1*4 + 2*5 + 3*6
	assert.InDelta(s.T(), expected, dot, 0.000001)

	dot = search.Dot(s.orthogonalVec1, s.orthogonalVec2)
	assert.InDelta(s.T(), 0.0, dot, 0.000001)

	dot = search.Dot(s.vec1, s.vec1)
	expected = 14.0 // 1*1 + 2*2 + 3*3
	assert.InDelta(s.T(), expected, dot, 0.000001)
}

func (s *SearchMetricsTestSuite) TestL2() {
	l2 := search.L2(s.vec1, s.vec2)
	expected := math.Sqrt(27.0) // sqrt((4-1)^2 + (5-2)^2 + (6-3)^2)
	assert.InDelta(s.T(), expected, l2, 0.000001)

	l2 = search.L2(s.vec1, s.vec1)
	assert.InDelta(s.T(), 0.0, l2, 0.000001)

	zeroVec := []float64{0.0, 0.0, 0.0}
	l2 = search.L2(s.vec1, zeroVec)
	expected = math.Sqrt(14.0) // sqrt(1^2 + 2^2 + 3^2)
	assert.InDelta(s.T(), expected, l2, 0.000001)
}

func (s *SearchMetricsTestSuite) TestEdgeCases() {
	emptyVec1 := []float64{}
	emptyVec2 := []float64{}

	assert.NotPanics(s.T(), func() {
		search.Cosine(emptyVec1, emptyVec2)
		search.Dot(emptyVec1, emptyVec2)
		search.L2(emptyVec1, emptyVec2)
	})

	differentLengthVec := []float64{1.0}
	assert.Panics(s.T(), func() {
		search.Cosine(s.vec1, differentLengthVec)
	})
}

func TestSearchMetrics(t *testing.T) {
	suite.Run(t, new(SearchMetricsTestSuite))
}
