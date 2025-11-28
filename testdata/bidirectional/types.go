package bidirectional

// API type (used for external communication)
type UserAPI struct {
	ID       string
	FullName string
	Email    string
}

// Domain type (used internally)
type UserDomain struct {
	ID       string
	FullName string
	Email    string
}

// Nested types for bidirectional conversion
type OrderAPI struct {
	OrderID string
	Items   []ItemAPI
}

type ItemAPI struct {
	Name     string
	Quantity int
}

type OrderDomain struct {
	OrderID string
	Items   []ItemDomain
}

type ItemDomain struct {
	Name     string
	Quantity int
}
