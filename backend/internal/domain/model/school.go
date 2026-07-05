package model

// School is the tenant that owns events and DJs; organizers join it through
// memberships (EL-103).
type School struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	City      string `json:"city"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}
