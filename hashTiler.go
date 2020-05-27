package tile

import (
	"fmt"
	"math"
	"math/rand"
)

// HashTiler is used for tile coding.
type HashTiler struct {
	numTilings int
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
func NewHashTiler(numTilings int) (*HashTiler, error) {
	ct.data = map[uint64]uint64{}
	switch {
	case numTilings < 1:
		return nil, InvalidNumTilingsError{numTilings, "must be at least 1"}
	case (numTilings & (numTilings - 1)) != 0:
		return nil, InvalidNumTilingsError{numTilings, "must be a power of 2"}
	}

	return &HashTiler{
		numTilings: numTilings,
	}, nil
}

// Tile returns a vector of length equal to `numTilings` (the argument to `NewHashTiler`). That vector contains hashes
// describing the input data. The length of the input data is not checked, but it is generally expected that the input
// length should always be the same for calls to the same HashTiler.
func (ht HashTiler) Tile(data []float64) []uint64 {
	tiles := make([]uint64, ht.numTilings)

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

		tiles[tileNum] = hash(coordinates)
	}

	return tiles
}

type CollisionTable struct {
	data       map[uint64]uint64
	max        uint64
	clearhits  int
	safe       int
	calls      int
	collisions int
}

var ct = CollisionTable{
	data: map[uint64]uint64{},
	max:  1000000,
}

func hash(ints []uint64) uint64 {
	ct.calls++
	j := hash_UNH(ints, ct.max, 449)
	ccheck := hash_UNH(ints, ^uint64(0), 457)

	old, ok := ct.data[j]
	switch {
	case !ok:
		ct.data[j] = ccheck
		ct.clearhits++
	case ccheck == old:
		ct.clearhits++
	case ct.safe == 0:
		ct.collisions++
		panic("Collision!")
	default:
		panic("Collision handling not implemented")
		// long h2 = 1 + 2 * hash_UNH(ints,num_ints,(MaxLONGINT)/4,449);
		// int i = 0;
		// while (++i) {
		// 	ct->collisions++;
		// 	j = (j+h2) % (ct->m);
		// 	/*printf("collision (%d) \n",j);*/
		// 	if (i > ct->m) {printf("\nTiles: Collision table out of Memory"); exit(0);}
		// 	if (ccheck == ct->data[j]) break;
		// 	if (ct->data[j] == -1) {ct->data[j] = ccheck; break;}
		// }
	}
	return j
}

var rndseq = make([]uint64, 2048)

func init() {
	for i := range rndseq {
		rndseq[i] = rand.Uint64()
	}
}

func hash_UNH(ints []uint64, max, increment uint64) uint64 {
	var sum uint64

	for i, v := range ints {
		/* add random table offset for this dimension and wrap around */
		v += (increment * uint64(i))
		v %= uint64(len(rndseq))

		/* add selected random number to sum */
		sum += rndseq[int(v)]
	}

	return sum % max
}
