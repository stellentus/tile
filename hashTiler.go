package tile

import (
	"encoding/binary"
	"hash/maphash"
	"math"
)

type HashTiler struct {
	tiles int
	hash  *maphash.Hash
}

func NewHashTiler(tiles int) (*HashTiler, error) {
	return &HashTiler{
		tiles: tiles,
		hash:  &maphash.Hash{},
	}, nil
}

func (ht HashTiler) Tile(data []float64) []uint64 {
	tiles := make([]uint64, ht.tiles)

	qstate := make([]int, len(data))
	base := make([]int, len(data))
	coordinates := make([]uint64, len(data)+1) /* one interval number per relevant dimension */

	/* quantize state to integers (henceforth, tile widths == ht.tiles) */
	for i := 0; i < len(data); i++ {
		qstate[i] = int(math.Floor(float64(data[i]) * float64(ht.tiles)))
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

		ht.hash.Reset()
		err := binary.Write(ht.hash, binary.LittleEndian, coordinates)
		if err != nil {
			panic(err.Error())
		}

		tiles[j] = ht.hash.Sum64()
	}

	return tiles
}
