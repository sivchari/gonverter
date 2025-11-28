// Package converter provides type conversion functions.
package converter

import (
	"github.com/sivchari/gonverter/examples/simple/domain"
	"github.com/sivchari/gonverter/examples/simple/handler"
)

// ConvertUserRequestNameToUserName is custom mapping
// UserRequest.FullName -> User.Name.
func ConvertUserRequestNameToUserName(src *handler.UserRequest, dst *domain.User) {
	dst.Name = src.FullName
}

// ConvertAddressRequestCityToAddressCity is custom mapping
// AddressRequest.ReqCity -> Address.City.
func ConvertAddressRequestCityToAddressCity(src *handler.AddressRequest, dst *domain.Address) {
	dst.City = src.ReqCity
}
