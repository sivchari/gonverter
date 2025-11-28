//go:build gonverter

package analyzer

import "github.com/sivchari/gonverter/runtime"

var _ = runtime.Register[*SameFields, *TargetSame]()
var _ = runtime.Register[*DifferentFields, *TargetDiff]()
