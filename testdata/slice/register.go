//go:build gonverter

package slice

import "github.com/sivchari/gonverter/runtime"

//go:generate go run ../../cmd/gonverter/main.go .

var _ = runtime.Register[*TeamRequest, *Team]()
