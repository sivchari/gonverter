//go:build gonverter

package maptype

import "github.com/sivchari/gonverter/runtime"

//go:generate go run ../../cmd/gonverter/main.go .

var _ = runtime.Register[*ConfigRequest, *Config]()
