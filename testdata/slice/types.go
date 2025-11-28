package slice

// Source types
type TeamRequest struct {
	Name    string
	Members []MemberRequest
}

type MemberRequest struct {
	Name string
	Role string
}

// Target types
type Team struct {
	Name    string
	Members []Member
}

type Member struct {
	Name string
	Role string
}
