package nested

// Source types
type UserRequest struct {
	Name    string
	Address AddressRequest
}

type AddressRequest struct {
	City    string
	ZipCode string
}

// Target types
type User struct {
	Name    string
	Address Address
}

type Address struct {
	City    string
	ZipCode string
}
