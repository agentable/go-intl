package ecma402types

// Part is the unit of output from PartitionPattern and the various Partition*
// formatter algorithms (PartitionNumberPattern, PartitionDateTimePattern, …).
// Type identifies the role of the segment ("literal", "integer", "decimal",
// "currency", …) and Value carries the raw text.
type Part struct {
	Type  string
	Value string
}

// Pattern is the slice form returned by PartitionPattern and consumed by
// formatter Partition* algorithms.
type Pattern = []Part
