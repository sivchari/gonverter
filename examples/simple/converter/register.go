//go:build gonverter

//go:generate go run ../../../cmd/gonverter .

package converter

import (
	"github.com/sivchari/gonverter/examples/simple/domain"
	"github.com/sivchari/gonverter/examples/simple/handler"
	"github.com/sivchari/gonverter/runtime"
)

// Register conversion pair
var _ = runtime.Register[*handler.UserRequest, *domain.User]()
