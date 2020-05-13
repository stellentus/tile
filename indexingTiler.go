package tile

import (
	"errors"
	"math"
)

// UnlimitedIndices can be provided to NewIndexingTiler to indicate there is no maximum number of indices.
const UnlimitedIndices = math.MaxInt64

// IndexingTiler is used for tile coding when a slice of indexes is desired. It runs slower than HashTiler.
type IndexingTiler struct {
	// ht is the underlying HashTiler that generates the hashes.
	ht *HashTiler

	// mp stores an index for each hash that has been seen so far.
	mp map[uint64]int
	// indexSize is the maximum number of indices to be stored in the map.
	indexSize int
	// currentIndex stores the number that will be used for the next index. (Therefore, it's also the number
	// of elements currently in the map, unless overflow has occurred.)
	currentIndex int

	// err stores any errors that occurred due to an index overflow
	err error
}

// NewIndexingTiler creates a new Indexing Tiler, which returns a slice of indexes based on the tiles' hashes.
// Hashes are calculated by HashTiler. See its documentation for further details regarding usage.
// If indexSize is UnlimitedIndices, then the number of indices is unlimited. Otherwise, the error is provided
// through CheckError().
func NewIndexingTiler(tiles, indexSize int) (*IndexingTiler, error) {
	ht, err := NewHashTiler(tiles)
	return &IndexingTiler{
		indexSize: indexSize,
		ht:        ht,
		mp:        make(map[uint64]int),
	}, err
}

// Tile returns a vector of length equal to tiles (the argument to NewIndexingTiler). That vector contains indices
// describing the input data. The indices range from 0 to indexSize-1 (where indexSize was an argument to NewIndexingTiler).
// The length of the input data is not checked, but it is generally expected that the input
// length should always be the same for calls to the same IndexingTiler.
func (it *IndexingTiler) Tile(data []float64) []int {
	hashes := it.ht.Tile(data)

	indices := make([]int, len(hashes))
	for i, hash := range hashes {
		idx, ok := it.mp[hash]
		if !ok {
			if it.currentIndex >= it.indexSize {
				it.err = errors.New("Too many tile indices were used, so one is being overwritten")
				it.currentIndex = 0
			}
			idx = it.currentIndex
			it.mp[hash] = it.currentIndex
			it.currentIndex++

		}
		indices[i] = idx
	}

	return indices
}

// CheckError returns an error if more indices were used than expected.
// There is no reason to check it if indexSize is UnlimitedIndices.
func (it IndexingTiler) CheckError() error {
	return it.err
}
