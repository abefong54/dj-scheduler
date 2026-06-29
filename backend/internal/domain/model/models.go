package model

type DJ struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	GenreTags []string `json:"genre_tags"`
	CreatedAt string   `json:"created_at"`
}

type Event struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	VenueName string   `json:"venue_name"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	Genres    []string `json:"genres"`
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
}

// PortalSlot is a single booking as seen from a DJ's self-service portal: it
// spans events, so it carries the event and stage names directly (US-009).
type PortalSlot struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	StageName string `json:"stage_name"`
	Genre     string `json:"genre"`
	SlotDate  string `json:"slot_date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Notes     string `json:"notes"`
}
