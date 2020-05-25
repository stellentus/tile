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
	ht Tiler

	// mp stores an index for each hash that has been seen so far.
	mp map[uint64]int
	// indexSize is the maximum number of indices to be stored in the map.
	indexSize int
	// currentIndex stores the number that will be used for the next index. (Therefore, it's also the number
	// of elements currently in the map, unless overflow has occurred or offset is used.)
	currentIndex int
	// offset is the offset added to every index. Indices are stored with this offset.
	offset int

	// err stores any errors that occurred due to an index overflow
	err error
}

// NewIndexingTiler creates a new Indexing Tiler, which returns a slice of indexes based on the tiles' hashes.
// Hashes are calculated by HashTiler. See its documentation for further details regarding usage.
// If indexSize is UnlimitedIndices, then the number of indices is unlimited. Otherwise, the error is provided
// through CheckError().
func NewIndexingTiler(til Tiler, indexSize int) (IndexTiler, error) {
	return NewIndexingTilerWithOffset(til, 0, indexSize)
}

// NewIndexingTilerWithOffset creates a new indexing tiler, but with an offset added to each provided index.
// Indices output by Tile will be in the range [offset, indexSize+offset).
func NewIndexingTilerWithOffset(til Tiler, offset, indexSize int) (IndexTiler, error) {
	return &IndexingTiler{
		ht:           til,
		indexSize:    indexSize,
		offset:       offset,
		currentIndex: offset,
		mp:           make(map[uint64]int),
	}, nil
}

// Tile returns a vector of indices describing the input data.
// The indices range from 0 to indexSize-1 (where indexSize was an argument to NewIndexingTiler).
// The length of the input data is not checked, but it is generally expected that the input
// length should always be the same for calls to the same IndexingTiler.
func (it *IndexingTiler) Tile(data []float64) []int {
	hashes := it.ht.Tile(data)

	indices := make([]int, len(hashes))
	for i, hash := range hashes {
		idx, ok := it.mp[hash]
		if !ok {
			if it.indexSize != UnlimitedIndices && it.currentIndex >= it.indexSize+it.offset {
				it.err = errors.New("Too many tile indices were used, so one is being overwritten")
				it.currentIndex = it.offset
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

// MaxIndices returns the maximum number of indices with a given maximum range of input data, number of dimensions
// in the input data, and number of tilings. It assumes the number of tilings is the same in each dimension.
// For example, when tiling a 4-dimensional input, with each input value ranging from -3 to 6, and 32 tilings, the
// call would be `MaxIndices(6-(-3), 4, 32)`.
func MaxIndices(maxRange, numDims, numTilings int) int {
	ranges := make([]int, numDims)
	for i := range ranges {
		ranges[i] = maxRange
	}
	return MaxIndicesForRanges(ranges, numTilings)
}

// MaxIndicesForRanges returns the maximum number of indices with a given maximum range of input data and number of
// tilings. The ranges slice contains the maximum range for each of the input values being hashed.
func MaxIndicesForRanges(ranges []int, numTilings int) int {
	result := 1
	for _, val := range ranges {
		// Add +1 because the most values along the "edges" of the region will include tiles that are otherwise entirely out of the region.
		result *= val + 1
	}

	result *= numTilings
	return result
}
