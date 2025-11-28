package slice

import "testing"

func TestSliceFieldConversion(t *testing.T) {
	src := &TeamRequest{
		Name: "Engineering",
		Members: []MemberRequest{
			{Name: "Alice", Role: "Lead"},
			{Name: "Bob", Role: "Developer"},
		},
	}

	dst := &Team{}
	ConvertTeamRequestToTeam(src, dst)

	if dst.Name != "Engineering" {
		t.Errorf("Name = %q, want %q", dst.Name, "Engineering")
	}

	if len(dst.Members) != 2 {
		t.Fatalf("len(Members) = %d, want 2", len(dst.Members))
	}

	if dst.Members[0].Name != "Alice" {
		t.Errorf("Members[0].Name = %q, want %q", dst.Members[0].Name, "Alice")
	}

	if dst.Members[0].Role != "Lead" {
		t.Errorf("Members[0].Role = %q, want %q", dst.Members[0].Role, "Lead")
	}

	if dst.Members[1].Name != "Bob" {
		t.Errorf("Members[1].Name = %q, want %q", dst.Members[1].Name, "Bob")
	}
}

func TestSliceFieldNilSlice(t *testing.T) {
	src := &TeamRequest{
		Name:    "Engineering",
		Members: nil,
	}

	dst := &Team{}
	ConvertTeamRequestToTeam(src, dst)

	if dst.Members != nil {
		t.Error("Members should be nil when src.Members is nil")
	}
}

func TestSliceFieldEmptySlice(t *testing.T) {
	src := &TeamRequest{
		Name:    "Engineering",
		Members: []MemberRequest{},
	}

	dst := &Team{}
	ConvertTeamRequestToTeam(src, dst)

	if dst.Members == nil {
		t.Error("Members should not be nil for empty slice")
	}

	if len(dst.Members) != 0 {
		t.Errorf("len(Members) = %d, want 0", len(dst.Members))
	}
}
