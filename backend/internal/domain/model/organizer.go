package model

// Organizer is an event organizer account, created on first Google sign-in.
type Organizer struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	GoogleID  string `json:"google_id"`
	CreatedAt string `json:"created_at"`
}
