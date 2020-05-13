package tile

import (
	"encoding/binary"
	"hash/maphash"
	"math"
)

// HashTiler is used for tile coding.
type HashTiler struct {
	tiles int
	seed  *maphash.Seed
}

// NewHashTiler creates a new tile coder with a unique random seed. The `tiles` argument determines the number of tiles
// that will be calculated. Tiling is uniform with the displacement vector (1,-1).
func NewHashTiler(tiles int) (*HashTiler, error) {
	seed := maphash.MakeSeed()
	return &HashTiler{
		tiles: tiles,
		seed:  &seed,
	}, nil
}

// Tile returns a vector of length equal to `tiles` (the argument to `NewHashTiler`). That vector contains hashes
// describing the input data. The length of the input data is not checked, but it is generally expected that the input
// length should always be the same for calls to the same HashTiler.
func (ht HashTiler) Tile(data []float64) []uint64 {
	tiles := make([]uint64, ht.tiles)
	hash := maphash.Hash{}
	hash.SetSeed(*ht.seed)

	qstate := make([]int, len(data))
	base := make([]int, len(data))
	coordinates := make([]uint64, len(data)+1) /* one interval number per relevant dimension */

	/* quantize state to integers (henceforth, tile widths == ht.tiles) */
	for i := 0; i < len(data); i++ {
		qstate[i] = int(math.Floor(data[i] * float64(ht.tiles)))
	}

	/*compute the tile numbers */
	for j := 0; j < ht.tiles; j++ {
		/* loop over each relevant dimension */
		i := 0
		for ; i < len(data); i++ {

			/* find coordinates of activated tile in tiling space */
			if qstate[i] >= base[i] {
				coordinates[i] = uint64(qstate[i] - ((qstate[i] - base[i]) % ht.tiles))
			} else {
				coordinates[i] = uint64(qstate[i] + 1 + ((base[i] - qstate[i] - 1) % ht.tiles) - ht.tiles)
			}

			/* compute displacement of next tiling in quantized space */
			base[i] += 1 + (2 * i)
		}
		/* add additional indices for tiling and hashing_set so they hash differently */
		coordinates[i] = uint64(j)

		hash.Reset()
		err := binary.Write(&hash, binary.LittleEndian, coordinates)
		if err != nil {
			panic(err.Error())
		}

		tiles[j] = hash.Sum64()
	}

	return tiles
}
