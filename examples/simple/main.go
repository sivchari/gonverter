// Package main provides an example of using gonverter.
package main

import (
	"fmt"

	"github.com/sivchari/gonverter/examples/simple/converter"
	"github.com/sivchari/gonverter/examples/simple/domain"
	"github.com/sivchari/gonverter/examples/simple/handler"
)

func main() {
	// Create request with nested struct
	req := &handler.UserRequest{
		FullName: "Takuma Shibuya",
		Email:    "takuma@example.com",
		Age:      30,
		Address: handler.AddressRequest{
			ReqCity: "Tokyo",
			ZipCode: "100-0001",
		},
	}

	// Convert
	user := &domain.User{}
	converter.ConvertUserRequestToUser(req, user)

	// Show result
	fmt.Printf("Name:    %s\n", user.Name)
	fmt.Printf("Email:   %s\n", user.Email)
	fmt.Printf("Age:     %d\n", user.Age)
	fmt.Printf("City:    %s\n", user.Address.City)
	fmt.Printf("ZipCode: %s\n", user.Address.ZipCode)
}
