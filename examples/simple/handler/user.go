// Package handler provides HTTP handler types.
package handler

// UserRequest is handler request.
type UserRequest struct {
	FullName string
	Email    string
	Age      int
	Address  AddressRequest
}

// AddressRequest is handler request for address.
type AddressRequest struct {
	ReqCity string
	ZipCode string
}
