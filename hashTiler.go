package tile

import (
	"encoding/binary"
	"errors"
	"hash/maphash"
	"math"
)

type HashTiler struct {
	tiles  int
	buffer []uint64
	hash   *maphash.Hash
}

type opt func(ht *HashTiler) error

func BufferOpt(buf []uint64) opt {
	return func(ht *HashTiler) error {
		switch {
		case ht.buffer != nil:
			return errors.New("HashTiler already has buffer")
		case len(buf) != ht.tiles:
			return errors.New("Hashtiler tile size does not match buffer")
		default:
			ht.buffer = buf
			return nil
		}
	}
}

func NewHashTiler(tiles int, opts ...opt) (*HashTiler, error) {
	ht := &HashTiler{
		tiles: tiles,
		hash:  &maphash.Hash{},
	}
	for _, o := range opts {
		if err := o(ht); err != nil {
			return nil, err
		}
	}
	return ht, nil
}

func (ht HashTiler) Tile(data []float64) []uint64 {
	tiles := ht.buffer
	if tiles == nil {
		tiles = make([]uint64, ht.tiles)
	}

	qstate := make([]int, len(data))
	base := make([]int, len(data))
	coordinates := make([]uint64, len(data)+1) /* one interval number per relevant dimension */

	/* quantize state to integers (henceforth, tile widths == ht.tiles) */
	for i := 0; i < len(data); i++ {
		qstate[i] = int(math.Floor(float64(data[i]) * float64(ht.tiles)))
		base[i] = 0
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
