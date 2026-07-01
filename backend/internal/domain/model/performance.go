package model

// GenreStat is one genre's slice of a DJ's performance reps (EL-043).
type GenreStat struct {
	Genre        string `json:"genre"`
	Reps         int    `json:"reps"`
	TotalMinutes int    `json:"total_minutes"`
}

// DJPerformance aggregates one DJ's stage time across all of an organizer's
// events (EL-043). Reps is the number of slots played; TotalMinutes sums their
// real durations (a set crossing midnight counts its true length). LastPlayed is
// the most recent slot_date, or "" if the DJ has never played.
type DJPerformance struct {
	DjID         string      `json:"dj_id"`
	DjName       string      `json:"dj_name"`
	Reps         int         `json:"reps"`
	TotalMinutes int         `json:"total_minutes"`
	LastPlayed   string      `json:"last_played"`
	ByGenre      []GenreStat `json:"by_genre"`
}

// RosterPerformance is one student's rep summary in the roster-wide view
// (EL-043). Every active student appears, including those with zero reps — the
// whole point is to surface who is being under-served.
type RosterPerformance struct {
	DjID         string `json:"dj_id"`
	DjName       string `json:"dj_name"`
	IsStudent    bool   `json:"is_student"`
	Reps         int    `json:"reps"`
	TotalMinutes int    `json:"total_minutes"`
	LastPlayed   string `json:"last_played"`
}
