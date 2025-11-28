package nested

import "testing"

func TestNestedConversion(t *testing.T) {
	src := &UserRequest{
		Name: "John Doe",
		Address: AddressRequest{
			City:    "Tokyo",
			ZipCode: "100-0001",
		},
	}

	dst := &User{}
	ConvertUserRequestToUser(src, dst)

	if dst.Name != "John Doe" {
		t.Errorf("Name = %q, want %q", dst.Name, "John Doe")
	}
	if dst.Address.City != "Tokyo" {
		t.Errorf("Address.City = %q, want %q", dst.Address.City, "Tokyo")
	}
	if dst.Address.ZipCode != "100-0001" {
		t.Errorf("Address.ZipCode = %q, want %q", dst.Address.ZipCode, "100-0001")
	}
}

func TestNestedConversionNilSrc(t *testing.T) {
	dst := &User{}
	ConvertUserRequestToUser(nil, dst)

	if dst.Name != "" {
		t.Errorf("Name should be empty, got %q", dst.Name)
	}
}
