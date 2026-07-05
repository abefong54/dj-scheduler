package model

// Membership links an organizer to a school with a role (EL-103). OrganizerID is
// nil while an invite is pending — it's matched to the person on their first
// sign-in (email match). Role is owner|admin|member; Status is pending|active.
type Membership struct {
	ID          string  `json:"id"`
	SchoolID    string  `json:"school_id"`
	OrganizerID *string `json:"organizer_id"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}
