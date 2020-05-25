package tile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUnlimitedIndexTiler(tiles int) (IndexTiler, error) {
	til, err := NewHashTiler(tiles)
	if err != nil {
		return nil, err
	}
	return NewIndexingTiler(til, UnlimitedIndices)
}

func ExampleIndexingTiler_Tile() {
	ht, err := newUnlimitedIndexTiler(1)
	if err != nil {
		fmt.Println(err.Error()) // HashTiler/IndexingTiler test code should have caught all errors
	}
	for _, data := range [][]float64{{3.14, 2.718}, {4, 2}, {3, 3}, {3, 2}} {
		fmt.Println("The index for", data, "is", ht.Tile(data))
	}
	// Output:
	// The index for [3.14 2.718] is [0]
	// The index for [4 2] is [1]
	// The index for [3 3] is [2]
	// The index for [3 2] is [0]
}

func ExampleIndexingTiler_Tile_second() {
	ht, err := newUnlimitedIndexTiler(4)
	if err != nil {
		fmt.Println(err.Error()) // HashTiler/IndexingTiler test code should have caught all errors
	}
	for _, data := range [][]float64{{4.99}, {5.24}, {5.25}, {5.49}} {
		fmt.Println("The indices for", data, "are", ht.Tile(data))
	}
	// Output:
	// The indices for [4.99] are [0 1 2 3]
	// The indices for [5.24] are [4 1 2 3]
	// The indices for [5.25] are [4 5 2 3]
	// The indices for [5.49] are [4 5 2 3]
}

func ExampleIndexingTiler_Tile_third() {
	// Test indexing with a constant offset added to each output.
	til, err := NewHashTiler(4)
	if err != nil {
		fmt.Println(err.Error()) // HashTiler test code should have caught all errors
	}
	ht, _ := NewIndexingTilerWithOffset(til, 15, UnlimitedIndices)
	for _, data := range [][]float64{{4.99}, {5.24}, {5.25}, {5.49}} {
		fmt.Println("The indices for", data, "are", ht.Tile(data))
	}
	// Output:
	// The indices for [4.99] are [15 16 17 18]
	// The indices for [5.24] are [19 16 17 18]
	// The indices for [5.25] are [19 20 17 18]
	// The indices for [5.49] are [19 20 17 18]
}

func TestIndexingTilerEqual(t *testing.T) {
	tests := map[string]struct {
		tiles int
		data  [][]float64
	}{
		"Same":            {16, [][]float64{{5}, {5}, {5}, {5}, {5}}},
		"Halfway":         {16, [][]float64{{5.003125}, {5}}},
		"Range":           {16, [][]float64{{5}, {5.03}, {5.0624}}},
		"Big step":        {4, [][]float64{{5.25}, {5.35}, {5.49}}},
		"multi-dimension": {1, [][]float64{{3.14, 2.718}, {3, 2}}},
		"range-dimension": {4, [][]float64{{3.14, 2.718}, {3.2, 2.7}, {3, 2.5}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := newUnlimitedIndexTiler(test.tiles)
			require.NoError(t, err)

			first := ht.Tile(test.data[0])
			for i, d := range test.data[1:] {
				other := ht.Tile(d)
				assert.Equalf(t, other, first, "tilings should match, failed element[%d]", i+1)
			}
		})
	}
}

func TestIndexingTilerNotEqual(t *testing.T) {
	tests := map[string]struct {
		tiles int
		data  [][]float64
	}{
		"Different":       {16, [][]float64{{5}, {6}, {7}, {8}, {9}}},
		"Range":           {16, [][]float64{{5}, {5.07}, {4.99}}},
		"Big step":        {4, [][]float64{{5.24}, {5.25}, {4.99}}},
		"multi-dimension": {1, [][]float64{{3.14, 2.718}, {4, 2}, {3, 3}}},
		"range-dimension": {4, [][]float64{{3.14, 2.718}, {3.2, 2.8}, {3.3, 2.5}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := newUnlimitedIndexTiler(test.tiles)
			require.NoError(t, err)

			first := ht.Tile(test.data[0])
			for i, d := range test.data[1:] {
				other := ht.Tile(d)
				assert.NotEqualf(t, other, first, "tilings should not match, failed element[%d]", i+1)
			}
		})
	}
}

func TestIndexingTilerUnequalAppearEqualWithSmallMap(t *testing.T) {
	tests := map[string]struct {
		tiles int
		data  [][]float64
	}{
		"Different":       {10, [][]float64{{5}, {6}, {7}, {8}, {9}}},
		"multi-dimension": {1, [][]float64{{3.14, 2.718}, {4, 2}, {3, 3}}},
		"range-dimension": {4, [][]float64{{3.14, 2.718}, {3.2, 2.8}, {3.3, 2.5}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			til, _ := NewHashTiler(test.tiles) // Assume HashTiler test code will catch errors, so don't check here
			ht, err := NewIndexingTiler(til, 1)
			require.NoError(t, err)

			for i, d := range test.data {
				indices := ht.Tile(d)
				for j, index := range indices {
					assert.Equalf(t, 0, index, "tilings should match because all indices are 0, failed element[%d:%d]", i+1, j)
				}
			}

			err = ht.CheckError()
			require.Error(t, err, "IndexingTiler should overflow when only 1 index is enabled")
		})
	}
}

func TestIndexingTilerCorrectTileLength(t *testing.T) {
	numberOfTilesTest := map[string]int{
		"One Tile":  1,
		"Two Tile":  2,
		"Four Tile": 4,
		"512 Tile":  512,
	}
	inputDataTest := map[string][]float64{
		"single value": []float64{5},
		"long list":    []float64{5, 1, 4, 5, 63, 46, 37},
	}

	for name, tiles := range numberOfTilesTest {
		t.Run(name, func(t *testing.T) {
			ht, err := newUnlimitedIndexTiler(tiles)
			require.NoError(t, err)
			for inputName, data := range inputDataTest {
				assert.Len(t, ht.Tile(data), tiles, inputName)
			}
		})
	}
}

func TestIndexingTilerUnitGrid2DWithOffset(t *testing.T) {
	tests := map[string]int{
		"One":  1,
		"Two":  2,
		"Four": 4,
		"Ten":  10,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := newUnlimitedIndexTiler(num)
			require.NoError(t, err)

			offset := 1 / float64(num)
			smallOffset := offset / 2
			gridOfIndices := [][]int{}
			gridOfOffsetIndices := [][]int{}

			for i := 0; i < num; i++ {
				for j := 0; j < num; j++ {
					x, y := float64(i)*offset, float64(j)*offset
					gridOfIndices = append(gridOfIndices, ht.Tile([]float64{x, y}))
					gridOfOffsetIndices = append(gridOfOffsetIndices, ht.Tile([]float64{x + smallOffset, y + smallOffset}))
				}
			}

			// Confirm that offset indices are equal to regular indices
			for i := range gridOfIndices {
				assert.EqualValues(t, gridOfIndices[i], gridOfOffsetIndices[i], "indices with small offset should be equal")
			}
		})
	}
}

const epsilon = 1e-5

func TestIndexingTilerUnitGrid2DRowsAreCorrect(t *testing.T) {
	// In this case, "correct" means each subsequent element has exactly one index different from the previous ones
	tests := map[string]int{
		// Obviously testing with a single tiling doesn't make sense
		"Two":   2,
		"Four":  4,
		"Eight": 8,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := newUnlimitedIndexTiler(num)
			require.NoError(t, err)

			offset := 1/float64(num) + epsilon

			for i := 0; i < num; i++ {
				x := float64(i) * offset
				gridOfIndices := make([][]int, num)

				for j := 0; j < num; j++ {
					y := float64(j) * offset
					gridOfIndices[j] = ht.Tile([]float64{x, y})
				}

				verifyIndexGridSlice(t, gridOfIndices)
			}
		})
	}
}

func TestIndexingTilerUnitGrid2DColumnsAreCorrect(t *testing.T) {
	// In this case, "correct" means each subsequent element has exactly one index different from the previous ones
	tests := map[string]int{
		// Obviously testing with a single tiling doesn't make sense
		"Two":   2,
		"Four":  4,
		"Eight": 8,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := newUnlimitedIndexTiler(num)
			require.NoError(t, err)

			offset := 1/float64(num) + epsilon

			for i := 0; i < num; i++ {
				y := float64(i) * offset
				gridOfIndices := make([][]int, num)

				for j := 0; j < num; j++ {
					x := float64(j) * offset
					gridOfIndices[j] = ht.Tile([]float64{x, y})
				}

				verifyIndexGridSlice(t, gridOfIndices)
			}
		})
	}
}

func verifyIndexGridSlice(t *testing.T, gridOfIndices [][]int) {
	// For each box in this row (or column), find the index which it has in common with all other boxes in the row (or column), and delete it
	lastIndices := gridOfIndices[len(gridOfIndices)-1]
	for i := range gridOfIndices[:len(gridOfIndices)-1] {
		commonIndex, err := intersectIntSlices(gridOfIndices, i, lastIndices) // Expect exactly one
		require.NoError(t, err)

		// If a common index was found, it's also in all remaining slices, so delete it.
		for j := i; j < len(gridOfIndices); j++ {
			gridOfIndices[j] = deleteIndex(gridOfIndices[j], commonIndex)
		}
	}

	assert.Len(t, gridOfIndices[len(gridOfIndices)-1], 1, "Final box should have one unique index")
}

func intersectIntSlices(gridOfIndices [][]int, idx int, lastIndices []int) (int, error) {
	theseIndices := gridOfIndices[idx]
	foundIndex := false
	index := int(0)

	for _, indexFromGrid := range theseIndices {
		for _, indexFromLast := range lastIndices {
			if indexFromGrid == indexFromLast {
				if foundIndex {
					return 0, fmt.Errorf("Found multiple shared indices at index %d", idx)
				}
				index = indexFromGrid
				foundIndex = true
			}
		}
	}

	var err error
	if !foundIndex {
		err = fmt.Errorf("No shared index was found at index %d", idx)
	}
	return index, err
}

func deleteIndex(indices []int, index int) []int {
	newIndices := make([]int, len(indices)-1)
	idx := 0
	for _, val := range indices {
		if val != index {
			newIndices[idx] = val
			idx++
		}
	}

	return newIndices
}

func BenchmarkIndexingTiler(b *testing.B) {
	benchmarks := []struct {
		name          string
		values, tiles int
	}{
		{"1x1", 1, 1},
		{"4x10", 4, 10},
		{"100x128", 100, 128},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			v := makeValues(bench.values)
			ht, _ := newUnlimitedIndexTiler(bench.tiles)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ht.Tile(v)
			}
		})
	}
}

func TestMaxIndices(t *testing.T) {
	tests := map[string]struct {
		maxRange, numDims, numTilings int
		expected                      int
	}{
		"Trivial case": {1, 1, 1, 2},
		"Simple case":  {4, 2, 8, 200},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := MaxIndices(test.maxRange, test.numDims, test.numTilings)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestMaxIndicesForRanges(t *testing.T) {
	tests := map[string]struct {
		ranges     []int
		numTilings int
		expected   int
	}{
		"Trivial case":              {[]int{1}, 1, 2},
		"Simple case":               {[]int{4, 4}, 8, 200},
		"Different dimensions case": {[]int{4, 2}, 16, 240},
		"Overly complicated case":   {[]int{4, 2, 7, 1, 12}, 16, 49920},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := MaxIndicesForRanges(test.ranges, test.numTilings)
			assert.Equal(t, test.expected, actual)
		})
	}
}
