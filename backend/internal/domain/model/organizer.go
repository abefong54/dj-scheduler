package model

// Organizer is an event organizer account, created on first sign-in via an
// external provider. GoogleID / LineID hold the provider identities; at least
// one is set (see migration 014's CHECK). An empty string means "no identity for
// that provider" (the column was NULL).
type Organizer struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	GoogleID  string `json:"google_id"`
	LineID    string `json:"line_id"`
	CreatedAt string `json:"created_at"`
}
