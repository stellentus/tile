package tile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexingTilerEqual(t *testing.T) {
	tests := map[string]struct {
		tiles int
		data  [][]float64
	}{
		"Same":            {10, [][]float64{{5}, {5}, {5}, {5}, {5}}},
		"Halfway":         {10, [][]float64{{5.05}, {5}}},
		"Range":           {10, [][]float64{{5.11}, {5.15}, {5.19}}},
		"Big step":        {3, [][]float64{{5.34}, {5.5}, {5.65}}},
		"multi-dimension": {1, [][]float64{{3.14, 2.718}, {3, 2}}},
		"range-dimension": {4, [][]float64{{3.14, 2.718}, {3.2, 2.7}, {3, 2.5}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewIndexingTiler(test.tiles, UnlimitedIndices)
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
		"Different":       {10, [][]float64{{5}, {6}, {7}, {8}, {9}}},
		"Range":           {10, [][]float64{{5.09}, {5.11}, {4.99}}},
		"Big step":        {3, [][]float64{{5.32}, {5.5}, {4.99}}},
		"multi-dimension": {1, [][]float64{{3.14, 2.718}, {4, 2}, {3, 3}}},
		"range-dimension": {4, [][]float64{{3.14, 2.718}, {3.2, 2.8}, {3.3, 2.5}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewIndexingTiler(test.tiles, UnlimitedIndices)
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
		"Range":           {10, [][]float64{{5.09}, {5.11}, {4.99}}},
		"Big step":        {3, [][]float64{{5.32}, {5.5}, {4.99}}},
		"multi-dimension": {1, [][]float64{{3.14, 2.718}, {4, 2}, {3, 3}}},
		"range-dimension": {4, [][]float64{{3.14, 2.718}, {3.2, 2.8}, {3.3, 2.5}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewIndexingTiler(test.tiles, 1)
			require.NoError(t, err)

			for i, d := range test.data {
				indices := ht.Tile(d)
				for j, index := range indices {
					assert.Equalf(t, 0, index, "tilings should match because all indices are 0, failed element[%d:%d]", i+1, j)
				}
			}
		})
	}
}

func TestIndexingTilerCorrectTileLength(t *testing.T) {
	numberOfTilesTest := map[string]int{
		"One Tile":  1,
		"Two Tile":  2,
		"Four Tile": 4,
		"500 Tile":  500,
	}
	inputDataTest := map[string][]float64{
		"single value": []float64{5},
		"long list":    []float64{5, 1, 4, 5, 63, 46, 37},
	}

	for name, tiles := range numberOfTilesTest {
		t.Run(name, func(t *testing.T) {
			ht, err := NewIndexingTiler(tiles, UnlimitedIndices)
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
		"Five": 5,
		"Ten":  10,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewIndexingTiler(num, UnlimitedIndices)
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

func TestIndexingTilerUnitGrid2DRowsAreCorrect(t *testing.T) {
	// In this case, "correct" means each subsequent element has exactly one index different from the previous ones
	tests := map[string]int{
		// Obviously testing with a single tiling doesn't make sense
		"Two":  2,
		"Five": 5,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewIndexingTiler(num, UnlimitedIndices)
			require.NoError(t, err)

			offset := 1 / float64(num)

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
		"Two":  2,
		"Five": 5,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewIndexingTiler(num, UnlimitedIndices)
			require.NoError(t, err)

			offset := 1 / float64(num)

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
		{"100x100", 100, 100},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			v := makeValues(bench.values)
			ht, _ := NewIndexingTiler(bench.tiles, UnlimitedIndices)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ht.Tile(v)
			}
		})
	}
}
