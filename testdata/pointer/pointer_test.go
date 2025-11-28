package pointer

import "testing"

func TestPointerFieldConversion(t *testing.T) {
	src := &UserRequest{
		Name: "John Doe",
		Profile: &ProfileRequest{
			Bio: "Developer",
			Age: 30,
		},
	}

	dst := &User{}
	ConvertUserRequestToUser(src, dst)

	if dst.Name != "John Doe" {
		t.Errorf("Name = %q, want %q", dst.Name, "John Doe")
	}

	if dst.Profile == nil {
		t.Fatal("Profile should not be nil")
	}

	if dst.Profile.Bio != "Developer" {
		t.Errorf("Profile.Bio = %q, want %q", dst.Profile.Bio, "Developer")
	}

	if dst.Profile.Age != 30 {
		t.Errorf("Profile.Age = %d, want %d", dst.Profile.Age, 30)
	}
}

func TestPointerFieldNilSrc(t *testing.T) {
	src := &UserRequest{
		Name:    "John Doe",
		Profile: nil,
	}

	dst := &User{}
	ConvertUserRequestToUser(src, dst)

	if dst.Name != "John Doe" {
		t.Errorf("Name = %q, want %q", dst.Name, "John Doe")
	}

	if dst.Profile != nil {
		t.Error("Profile should be nil when src.Profile is nil")
	}
}
