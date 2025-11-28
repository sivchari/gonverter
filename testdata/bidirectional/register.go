//go:build gonverter

package bidirectional

import "github.com/sivchari/gonverter/runtime"

//go:generate go run ../../cmd/gonverter/main.go .

// RegisterBidirectional generates both UserAPI→UserDomain and UserDomain→UserAPI
var _ = runtime.RegisterBidirectional[*UserAPI, *UserDomain]()

// RegisterBidirectional generates both OrderAPI→OrderDomain and OrderDomain→OrderAPI
var _ = runtime.RegisterBidirectional[*OrderAPI, *OrderDomain]()
