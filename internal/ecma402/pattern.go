package ecma402

// Part is the unit of output from PartitionPattern and formatter partition
// algorithms.
type Part struct {
	Type  string
	Value string
}

// Pattern is the slice form returned by PartitionPattern.
type Pattern = []Part
