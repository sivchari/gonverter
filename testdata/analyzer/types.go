package analyzer

// SameFields has same name and type fields
type SameFields struct {
	Name  string
	Email string
	Age   int
}

// TargetSame has same field structure
type TargetSame struct {
	Name  string
	Email string
	Age   int
}

// DifferentFields has different field names
type DifferentFields struct {
	FullName string
	Email    string
}

// TargetDiff has different field names
type TargetDiff struct {
	Name  string
	Email string
}
