package model

// Challenge is a teacher-authored task (migration 015). A "core" challenge is
// school-wide and carries badge metadata students earn by passing it; a "custom"
// challenge is assigned to one specific student (AssignedDJID) and carries no
// badge. Challenges are a library — progress and badges are read-model queries
// over Submission, not their own tables.
type Challenge struct {
	ID           string `json:"id"`
	SchoolID     string `json:"school_id"`
	Type         string `json:"type"` // "core" | "custom"
	Title        string `json:"title"`
	Description  string `json:"description"`
	Skill        string `json:"skill"`          // optional tag
	AssignedDJID string `json:"assigned_dj_id"` // custom only, else ""
	BadgeLabel   string `json:"badge_label"`    // core only
	BadgeIcon    string `json:"badge_icon"`     // core only
	SortOrder    int    `json:"sort_order"`
	CreatedBy    string `json:"created_by"` // organizer
	Archived     bool   `json:"archived"`
	CreatedAt    string `json:"created_at"`
}
