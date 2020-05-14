package tile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAggregateTiler() (Tiler, error) {
	til1, _ := NewHashTiler(3)
	til2, _ := NewHashTiler(2)
	return NewAggregateTiler([]Tiler{til1, til2})
}

func ExampleAggregateTiler_Tile() {
	til, _ := newAggregateTiler()
	it, _ := NewIndexingTiler(til, UnlimitedIndices)
	test := [][]float64{
		{3, 4},
		{3.35, 4},
		{3.68, 4},

		{3, 4.35},
		{3.35, 4.35},
		{3.68, 4.35},

		{3, 4.68},
		{3.35, 4.68},
		{3.68, 4.68},
	}
	for _, data := range test {
		fmt.Println("The index for", data, "is", it.Tile(data))
	}
	// Output:
	// The index for [3 4] is [0 1 2 3 4]
	// The index for [3.35 4] is [0 5 2 3 4]
	// The index for [3.68 4] is [0 5 6 3 7]
	// The index for [3 4.35] is [0 8 2 3 4]
	// The index for [3.35 4.35] is [0 9 2 3 4]
	// The index for [3.68 4.35] is [0 9 6 3 7]
	// The index for [3 4.68] is [0 8 10 3 11]
	// The index for [3.35 4.68] is [0 9 10 3 11]
	// The index for [3.68 4.68] is [0 9 12 3 13]
}

func TestCreateAggregateTiler(t *testing.T) {
	_, err := newAggregateTiler()
	require.NoError(t, err)
}

func TestComplexAggregateTiler(t *testing.T) {
	numDims := 4
	numTiles := 2
	data := []float64{1, 2, 3, 4}
	expectedLen := (len(data) + len(data)*(len(data)-1)/2) * numTiles // The number of Tilers for the singles and the pairs, times numTiles

	// Create tilers for each single dimension
	singles, err := NewSinglesTiler(numDims, numTiles)
	require.NoError(t, err)

	// Create tilers for each pair of dimensions
	pairs, err := NewPairsTiler(numDims, numTiles)
	require.NoError(t, err)

	// Create a mega-tiler that appends pairs and singles
	til, err := NewAggregateTiler([]Tiler{singles, pairs})
	require.NoError(t, err)

	// Put it all through an IndexingTiler to make it deterministic
	it, err := NewIndexingTiler(til, expectedLen)
	require.NoError(t, err)

	result := it.Tile(data)
	assert.Len(t, result, expectedLen)
	for i, resI := range result {
		assert.Equal(t, i, resI)
	}
}
