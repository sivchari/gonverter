package bidirectional

import "testing"

func TestBidirectionalUserConversion(t *testing.T) {
	// Test API → Domain
	apiUser := &UserAPI{
		ID:       "user123",
		FullName: "John Doe",
		Email:    "john@example.com",
	}

	domainUser := &UserDomain{}
	ConvertUserAPIToUserDomain(apiUser, domainUser)

	if domainUser.ID != "user123" {
		t.Errorf("ID = %q, want %q", domainUser.ID, "user123")
	}

	if domainUser.FullName != "John Doe" {
		t.Errorf("FullName = %q, want %q", domainUser.FullName, "John Doe")
	}

	if domainUser.Email != "john@example.com" {
		t.Errorf("Email = %q, want %q", domainUser.Email, "john@example.com")
	}

	// Test Domain → API (reverse conversion)
	domainUser2 := &UserDomain{
		ID:       "user456",
		FullName: "Jane Doe",
		Email:    "jane@example.com",
	}

	apiUser2 := &UserAPI{}
	ConvertUserDomainToUserAPI(domainUser2, apiUser2)

	if apiUser2.ID != "user456" {
		t.Errorf("ID = %q, want %q", apiUser2.ID, "user456")
	}

	if apiUser2.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want %q", apiUser2.FullName, "Jane Doe")
	}

	if apiUser2.Email != "jane@example.com" {
		t.Errorf("Email = %q, want %q", apiUser2.Email, "jane@example.com")
	}
}

func TestBidirectionalOrderConversion(t *testing.T) {
	// Test API → Domain with nested slice
	apiOrder := &OrderAPI{
		OrderID: "order123",
		Items: []ItemAPI{
			{Name: "Item A", Quantity: 2},
			{Name: "Item B", Quantity: 5},
		},
	}

	domainOrder := &OrderDomain{}
	ConvertOrderAPIToOrderDomain(apiOrder, domainOrder)

	if domainOrder.OrderID != "order123" {
		t.Errorf("OrderID = %q, want %q", domainOrder.OrderID, "order123")
	}

	if len(domainOrder.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(domainOrder.Items))
	}

	if domainOrder.Items[0].Name != "Item A" {
		t.Errorf("Items[0].Name = %q, want %q", domainOrder.Items[0].Name, "Item A")
	}

	if domainOrder.Items[0].Quantity != 2 {
		t.Errorf("Items[0].Quantity = %d, want 2", domainOrder.Items[0].Quantity)
	}

	// Test Domain → API (reverse conversion with nested slice)
	domainOrder2 := &OrderDomain{
		OrderID: "order456",
		Items: []ItemDomain{
			{Name: "Item X", Quantity: 10},
		},
	}

	apiOrder2 := &OrderAPI{}
	ConvertOrderDomainToOrderAPI(domainOrder2, apiOrder2)

	if apiOrder2.OrderID != "order456" {
		t.Errorf("OrderID = %q, want %q", apiOrder2.OrderID, "order456")
	}

	if len(apiOrder2.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(apiOrder2.Items))
	}

	if apiOrder2.Items[0].Name != "Item X" {
		t.Errorf("Items[0].Name = %q, want %q", apiOrder2.Items[0].Name, "Item X")
	}

	if apiOrder2.Items[0].Quantity != 10 {
		t.Errorf("Items[0].Quantity = %d, want 10", apiOrder2.Items[0].Quantity)
	}
}

func TestBidirectionalRoundTrip(t *testing.T) {
	// Test round-trip: API → Domain → API
	original := &UserAPI{
		ID:       "user789",
		FullName: "Round Trip",
		Email:    "roundtrip@example.com",
	}

	domain := &UserDomain{}
	ConvertUserAPIToUserDomain(original, domain)

	restored := &UserAPI{}
	ConvertUserDomainToUserAPI(domain, restored)

	if restored.ID != original.ID {
		t.Errorf("ID = %q, want %q", restored.ID, original.ID)
	}

	if restored.FullName != original.FullName {
		t.Errorf("FullName = %q, want %q", restored.FullName, original.FullName)
	}

	if restored.Email != original.Email {
		t.Errorf("Email = %q, want %q", restored.Email, original.Email)
	}
}

func TestBidirectionalNilSource(t *testing.T) {
	// Test nil source handling for both directions
	dst := &UserDomain{}
	ConvertUserAPIToUserDomain(nil, dst)

	// Should not panic and dst should remain unchanged
	if dst.ID != "" {
		t.Errorf("ID = %q, want empty string", dst.ID)
	}

	dst2 := &UserAPI{}
	ConvertUserDomainToUserAPI(nil, dst2)

	if dst2.ID != "" {
		t.Errorf("ID = %q, want empty string", dst2.ID)
	}
}
