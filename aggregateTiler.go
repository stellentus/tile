package tile

// aggregateTiler is used for tile coding when multiple Tilers must work together.
type aggregateTiler struct {
	// tils are the underlying Tilers that generate the hashes.
	tils []Tiler
}

// NewAggregateTiler creates a new Tiler which returns all of the hashes provided
// by the individual Tilers.
func NewAggregateTiler(tils []Tiler) (Tiler, error) {
	return &aggregateTiler{
		tils: tils,
	}, nil
}

// Tile returns a vector of indices describing the input data.
func (til *aggregateTiler) Tile(data []float64) []uint64 {
	output := []uint64{}

	for _, til := range til.tils {
		output = append(output, til.Tile(data)...)
	}

	return output
}

// singleTiler is used for tile coding when multiple Tilers must work together.
type singleTiler struct {
	idx int
	til Tiler
}

func (til *singleTiler) Tile(data []float64) []uint64 {
	return til.Tile([]float64{data[til.idx]})
}

// NewSinglesTiler creates a new Tiler which tiles each dimension individually.
func NewSinglesTiler(numDims, tiles int) (Tiler, error) {
	tils := make([]Tiler, numDims)
	for i := range tils {
		til, err := NewHashTiler(tiles)
		if err != nil {
			return nil, err
		}
		tils[i] = &singleTiler{
			idx: i,
			til: til,
		}
	}
	return NewAggregateTiler(tils)
}

// pairTiler is used for tile coding when multiple Tilers must work together.
type pairTiler struct {
	idx1, idx2 int
	til        Tiler
}

func (til *pairTiler) Tile(data []float64) []uint64 {
	return til.Tile([]float64{data[til.idx1], data[til.idx2]})
}

// NewPairsTiler creates a new Tiler which tiles each pair of dimensions.
func NewPairsTiler(numDims, tiles int) (Tiler, error) {
	tils := make([]Tiler, numDims)
	for i := 0; i < numDims; i++ {
		for j := i + 1; j < numDims; j++ {
			til, err := NewHashTiler(tiles)
			if err != nil {
				return nil, err
			}
			tils[i] = &pairTiler{
				idx1: i,
				idx2: j,
				til:  til,
			}
		}
	}
	return NewAggregateTiler(tils)
}
