//go:build gonverter

package pointer

import "github.com/sivchari/gonverter/runtime"

//go:generate go run ../../cmd/gonverter/main.go .

var _ = runtime.Register[*UserRequest, *User]()
