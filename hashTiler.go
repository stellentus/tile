package tile

import (
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"math"
)

// HashTiler is used for tile coding.
type HashTiler struct {
	numTilings int
	seed       *maphash.Seed
}

// InvalidNumTilingsError is returned
type InvalidNumTilingsError struct {
	NumTilings int
	Reason     string
}

func (err InvalidNumTilingsError) Error() string {
	return fmt.Sprintf("invalid number of tilings (%d): %s", err.NumTilings, err.Reason)
}

// NewHashTiler creates a new tile coder with a unique random seed. The `numTilings` argument determines the number of
// tilings that will be calculated. Tiling is uniform with the displacement vector (1,-1).
func NewHashTiler(numTilings int) (Tiler, error) {
	switch {
	case numTilings < 1:
		return nil, InvalidNumTilingsError{numTilings, "must be at least 1"}
	case (numTilings & (numTilings - 1)) != 0:
		return nil, InvalidNumTilingsError{numTilings, "must be a power of 2"}
	}

	seed := maphash.MakeSeed()
	return &HashTiler{
		numTilings: numTilings,
		seed:       &seed,
	}, nil
}

// Tile returns a vector of length equal to `numTilings` (the argument to `NewHashTiler`). That vector contains hashes
// describing the input data. The length of the input data is not checked, but it is generally expected that the input
// length should always be the same for calls to the same HashTiler.
func (ht HashTiler) Tile(data []float64) []uint64 {
	tiles := make([]uint64, ht.numTilings)
	hash := maphash.Hash{}
	hash.SetSeed(*ht.seed)

	qstate := make([]int, len(data))
	offsets := make([]int, len(data))
	coordinates := make([]uint64, len(data)+1) // one interval number per relevant dimension

	// quantize state to integers (henceforth, tile widths == ht.numTilings)
	for i := 0; i < len(data); i++ {
		qstate[i] = int(math.Floor(data[i] * float64(ht.numTilings)))
	}

	//compute the tile numbers
	for tileNum := 0; tileNum < ht.numTilings; tileNum++ {
		// loop over each relevant dimension
		for i, q := range qstate {
			diff := q - offsets[i]
			// find coordinates of activated tile in tiling space
			if diff >= 0 {
				// This shifts q toward offsets[i] so it's at a multiple of numTilings (plus the offset)
				coordinates[i] = uint64(q - ((diff) % ht.numTilings))
			} else {
				// We always want to shift the value to the multiple of numTilings below its value, so when
				// q < offsets[i], it's necessary to move it away from offsets[i] instead of toward it.
				coordinates[i] = uint64(q - ((diff + 1) % ht.numTilings) - ht.numTilings + 1)
			}
			offsets[i] += 1 + 2*i
		}
		// add additional indices for tiling and hashing_set so they hash differently
		coordinates[len(data)] = uint64(tileNum)

		hash.Reset()
		err := binary.Write(&hash, binary.LittleEndian, coordinates)
		if err != nil {
			panic(err.Error())
		}

		tiles[tileNum] = hash.Sum64()
	}

	return tiles
}
