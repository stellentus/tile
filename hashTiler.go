package tile

import (
	"encoding/binary"
	"hash/maphash"
	"math"
)

// hashTiler is used for tile coding.
type hashTiler struct {
	numTilings int
	seed       *maphash.Seed
}

// NewHashTiler creates a new tile coder with a unique random seed. The `numTilings` argument determines the number of
// tilings that will be calculated. Tiling is uniform with the displacement vector (1,-1).
func NewHashTiler(numTilings int) (Tiler, error) {
	seed := maphash.MakeSeed()
	return &hashTiler{
		numTilings: numTilings,
		seed:       &seed,
	}, nil
}

// Tile returns a vector of length equal to `numTilings` (the argument to `NewHashTiler`). That vector contains hashes
// describing the input data. The length of the input data is not checked, but it is generally expected that the input
// length should always be the same for calls to the same hashTiler.
func (ht hashTiler) Tile(data []float64) []uint64 {
	tiles := make([]uint64, ht.numTilings)
	hash := maphash.Hash{}
	hash.SetSeed(*ht.seed)

	qstate := make([]int, len(data))
	coordinates := make([]uint64, len(data)+1) // one interval number per relevant dimension

	// quantize state to integers (henceforth, tile widths == ht.numTilings)
	for i := 0; i < len(data); i++ {
		qstate[i] = int(math.Floor(data[i] * float64(ht.numTilings)))
	}

	//compute the tile numbers
	for tileOffset := 0; tileOffset < ht.numTilings; tileOffset++ {
		// loop over each relevant dimension
		for i, q := range qstate {
			diff := q - tileOffset
			// find coordinates of activated tile in tiling space
			if diff >= 0 {
				// This shifts q toward tileOffset so it's at a multiple of 4 (plus the offset)
				coordinates[i] = uint64(q - ((diff) % ht.numTilings))
			} else {
				// We always want to shift the value to the multiple of 4 below its value, so when
				// q < tileOffset, it's necessary to move it away from tileOffset instead of toward it.
				coordinates[i] = uint64(q - ((diff + 1) % ht.numTilings) - ht.numTilings + 1)
			}
		}
		// add additional indices for tiling and hashing_set so they hash differently
		coordinates[len(data)] = uint64(tileOffset)

		hash.Reset()
		err := binary.Write(&hash, binary.LittleEndian, coordinates)
		if err != nil {
			panic(err.Error())
		}

		tiles[tileOffset] = hash.Sum64()
	}

	return tiles
}
