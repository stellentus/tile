package tile

type Tiler interface {
	// Tile returns a vector of hashes describing the input data. The length of the input data is not checked,
	// but it is generally expected that the input length should always be the same for calls to the same Tiler.
	Tile(data []float64) []uint64
}

type IndexTiler interface {
	// Tile returns a vector of indices describing the input data. The length of the input data is not checked,
	// but it is generally expected that the input length should always be the same for calls to the same IndexTiler.
	Tile(data []float64) []int

	// CheckError returns an error if any errors have occurred.
	CheckError() error
}
