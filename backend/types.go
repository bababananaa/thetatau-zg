package main

// Status values stored in people.status.
const (
	StatusPNM    = "pnm"
	StatusActive = "active"
	StatusAlumni = "alumni"
)

// LoginRequest is the body of POST /api/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PublicPerson is the public view returned by GET /api/people.
// PledgeClass and Roll are null for PNMs.
type PublicPerson struct {
	ID          int64   `json:"id"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	PledgeClass *string `json:"pledgeClass"`
	Roll        *int    `json:"roll"`
}

// PrivatePerson is the authenticated view returned by POST /api/login,
// GET /api/me, and GET /api/directory.
type PrivatePerson struct {
	PublicPerson
	Email    string  `json:"email"`
	Phone    *string `json:"phone"`
	Address  *string `json:"address"`
	Company  *string `json:"company"`
	JobTitle *string `json:"jobTitle"`
	Status   string  `json:"status"`
}

// ErrorResponse is the body of every error response.
type ErrorResponse struct {
	Error string `json:"error"`
}
