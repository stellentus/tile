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
