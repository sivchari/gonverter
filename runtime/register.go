// Package runtime provides runtime utilities for gonverter.
package runtime

// Registration represents conversion registration type.
type Registration struct{}

// Register registers conversion between two types.
// This function does nothing at runtime; it's used as a marker for gonverter to extract type information for code generation.
func Register[From, To any]() Registration {
	return Registration{}
}

// RegisterBidirectional registers bidirectional conversion between two types.
// This generates both From→To and To→From conversion functions.
// This function does nothing at runtime; it's used as a marker for gonverter to extract type information for code generation.
func RegisterBidirectional[From, To any]() Registration {
	return Registration{}
}
