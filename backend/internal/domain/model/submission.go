package model

// Submission is a student's attempt at a Challenge (migration 015) — the one
// write-path record that matters: it is simultaneously the feed item and the
// progress driver. State machine: posted -> passed, or posted -> needs_work ->
// posted -> ... One active row per (DJID, ChallengeID). When a core challenge is
// passed, its badge metadata is snapshotted onto EarnedBadgeLabel/EarnedBadgeIcon
// so the earned badge is permanent even if the teacher later edits/archives the
// challenge. Media is a reference (not a file): render-by-kind on the frontend.
type Submission struct {
	ID          string `json:"id"`
	SchoolID    string `json:"school_id"`
	ChallengeID string `json:"challenge_id"`
	DJID        string `json:"dj_id"` // the student
	// Media reference — the seam that lets us own files later without touching the
	// challenge loop. MediaKind is "embed" in v1 (future: "file").
	MediaKind     string `json:"media_kind"`
	MediaProvider string `json:"media_provider"` // youtube | soundcloud | mixcloud
	MediaRef      string `json:"media_ref"`      // extracted id / normalized url
	Caption       string `json:"caption"`
	Status        string `json:"status"`      // "posted" | "passed" | "needs_work"
	ReviewNote    string `json:"review_note"` // nullable
	// EarnedBadge* are the snapshot copied from the challenge when a core
	// challenge is passed — permanent, never re-read from the live challenge row.
	EarnedBadgeLabel string `json:"earned_badge_label"`
	EarnedBadgeIcon  string `json:"earned_badge_icon"`
	ReviewedBy       string `json:"reviewed_by"` // nullable organizer
	ReviewedAt       string `json:"reviewed_at"` // nullable
	CreatedAt        string `json:"created_at"`
}
