package tile

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashTilerEqual(t *testing.T) {
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
			ht, err := NewHashTiler(test.tiles)
			require.NoError(t, err)

			first := ht.Tile(test.data[0])
			for i, d := range test.data[1:] {
				other := ht.Tile(d)
				assert.Equalf(t, other, first, "tilings should match, failed element[%d]", i+1)
			}
		})
	}
}

func TestHashTilerNotEqual(t *testing.T) {
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
			ht, err := NewHashTiler(test.tiles)
			require.NoError(t, err)

			first := ht.Tile(test.data[0])
			for i, d := range test.data[1:] {
				other := ht.Tile(d)
				assert.NotEqualf(t, other, first, "tilings should not match, failed element[%d]", i+1)
			}
		})
	}
}

func TestHashTilerCorrectTileLength(t *testing.T) {
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
			ht, err := NewHashTiler(tiles)
			require.NoError(t, err)
			for inputName, data := range inputDataTest {
				assert.Len(t, ht.Tile(data), tiles, inputName)
			}
		})
	}
}

func TestHashTilerUnitGrid2DWithOffset(t *testing.T) {
	tests := map[string]int{
		"One":  1,
		"Two":  2,
		"Five": 5,
		"Ten":  10,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewHashTiler(num)
			require.NoError(t, err)

			offset := 1 / float64(num)
			smallOffset := offset / 2
			gridOfHashes := [][]uint64{}
			gridOfOffsetHashes := [][]uint64{}

			for i := 0; i < num; i++ {
				for j := 0; j < num; j++ {
					x, y := float64(i)*offset, float64(j)*offset
					gridOfHashes = append(gridOfHashes, ht.Tile([]float64{x, y}))
					gridOfOffsetHashes = append(gridOfOffsetHashes, ht.Tile([]float64{x + smallOffset, y + smallOffset}))
				}
			}

			// Confirm that offset hashes are equal to regular hashes
			for i := range gridOfHashes {
				assert.EqualValues(t, gridOfHashes[i], gridOfOffsetHashes[i], "hashes with small offset should be equal")
			}
		})
	}
}

func TestHashTilerUnitGrid2DRowsAreCorrect(t *testing.T) {
	// In this case, "correct" means each subsequent element has exactly one hash different from the previous ones
	tests := map[string]int{
		// Obviously testing with a single tiling doesn't make sense
		"Two":  2,
		"Five": 5,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewHashTiler(num)
			require.NoError(t, err)

			offset := 1 / float64(num)

			for i := 0; i < num; i++ {
				x := float64(i) * offset
				gridOfHashes := make([][]uint64, num)

				for j := 0; j < num; j++ {
					y := float64(j) * offset
					gridOfHashes[j] = ht.Tile([]float64{x, y})
				}

				verifyGridSlice(t, gridOfHashes)
			}
		})
	}
}

func TestHashTilerUnitGrid2DColumnsAreCorrect(t *testing.T) {
	// In this case, "correct" means each subsequent element has exactly one hash different from the previous ones
	tests := map[string]int{
		// Obviously testing with a single tiling doesn't make sense
		"Two":  2,
		"Five": 5,
	}

	for name, num := range tests {
		t.Run(name, func(t *testing.T) {
			ht, err := NewHashTiler(num)
			require.NoError(t, err)

			offset := 1 / float64(num)

			for i := 0; i < num; i++ {
				y := float64(i) * offset
				gridOfHashes := make([][]uint64, num)

				for j := 0; j < num; j++ {
					x := float64(j) * offset
					gridOfHashes[j] = ht.Tile([]float64{x, y})
				}

				verifyGridSlice(t, gridOfHashes)
			}
		})
	}
}

func verifyGridSlice(t *testing.T, gridOfHashes [][]uint64) {
	// For each box in this row (or column), find the hash which it has in common with all other boxes in the row (or column), and delete it
	lastHashes := gridOfHashes[len(gridOfHashes)-1]
	for i := range gridOfHashes[:len(gridOfHashes)-1] {
		commonHash, err := intersect(gridOfHashes, i, lastHashes) // Expect exactly one
		require.NoError(t, err)

		// If a common hash was found, it's also in all remaining slices, so delete it.
		for j := i; j < len(gridOfHashes); j++ {
			gridOfHashes[j] = deleteHash(gridOfHashes[j], commonHash)
		}
	}

	assert.Len(t, gridOfHashes[len(gridOfHashes)-1], 1, "Final box should have one unique hash")
}

func intersect(gridOfHashes [][]uint64, idx int, lastHashes []uint64) (uint64, error) {
	theseHashes := gridOfHashes[idx]
	foundHash := false
	hash := uint64(0)

	for _, hashFromGrid := range theseHashes {
		for _, hashFromLast := range lastHashes {
			if hashFromGrid == hashFromLast {
				if foundHash {
					return 0, fmt.Errorf("Found multiple shared hashes at index %d", idx)
				}
				hash = hashFromGrid
				foundHash = true
			}
		}
	}

	var err error
	if !foundHash {
		err = fmt.Errorf("No shared hash was found at index %d", idx)
	}
	return hash, err
}

func deleteHash(hashes []uint64, hash uint64) []uint64 {
	newHashes := make([]uint64, len(hashes)-1)
	idx := 0
	for _, val := range hashes {
		if val != hash {
			newHashes[idx] = val
			idx++
		}
	}

	return newHashes
}

func makeValues(size int) []float64 {
	v := make([]float64, size)
	for i := range v {
		v[i] = rand.Float64() * 10
	}
	return v
}

func BenchmarkHashTiler(b *testing.B) {
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
			ht, _ := NewHashTiler(bench.tiles)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ht.Tile(v)
			}
		})
	}
}
