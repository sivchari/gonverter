//go:build gonverter

package simple

import "github.com/sivchari/gonverter/runtime"

// Register conversion pair
var _ = runtime.Register[*Source, *Target]()
