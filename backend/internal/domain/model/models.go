package model

type DJ struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	GenreTags []string `json:"genre_tags"`
	// Certifications are the genres this DJ is cleared to perform (EL-019).
	Certifications []string `json:"certifications"`
	// IsStudent is true for active students (the certification gate applies) and
	// false for graduates/pros (gate bypassed).
	IsStudent bool   `json:"is_student"`
	CreatedAt string `json:"created_at"`
}

type Event struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	VenueName string   `json:"venue_name"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	Genres    []string `json:"genres"`
	// LineNotifyEnabled reports whether a LINE Notify token is stored for the
	// event (US-006). The raw token is never exposed through the model.
	LineNotifyEnabled bool `json:"line_notify_enabled"`
}

type Stage struct {
	ID           string `json:"id"`
	EventID      string `json:"event_id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

type Slot struct {
	ID        string `json:"id"`
	EventID   string `json:"event_id"`
	StageID   string `json:"stage_id"`
	StageName string `json:"stage_name"`
	DjID      string `json:"dj_id"`
	DjName    string `json:"dj_name"`
	Genre     string `json:"genre"`
	SlotDate  string `json:"slot_date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Notes     string `json:"notes"`
	// DJConfirmation is the DJ's portal response: "confirmed", "flagged", or nil
	// (no response yet). Pointer so null round-trips as JSON null (US-011).
	DJConfirmation *string `json:"dj_confirmation"`
}

// PortalSlot is a single booking as seen from a DJ's self-service portal: it
// spans events, so it carries the event and stage names directly (US-009).
type PortalSlot struct {
	ID        string `json:"id"`
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	StageName string `json:"stage_name"`
	Genre     string `json:"genre"`
	SlotDate  string `json:"slot_date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Notes     string `json:"notes"`
	// DJConfirmation: "confirmed", "flagged", or nil (no response yet) — US-011.
	DJConfirmation *string `json:"dj_confirmation"`
}
