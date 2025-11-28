// Package domain provides domain models.
package domain

// User is domain model.
type User struct {
	Name    string
	Email   string
	Age     int
	Address Address
}

// Address is domain model for address.
type Address struct {
	City    string
	ZipCode string
}
