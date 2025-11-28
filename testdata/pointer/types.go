package pointer

// Source types
type UserRequest struct {
	Name    string
	Profile *ProfileRequest
}

type ProfileRequest struct {
	Bio string
	Age int
}

// Target types
type User struct {
	Name    string
	Profile *Profile
}

type Profile struct {
	Bio string
	Age int
}
