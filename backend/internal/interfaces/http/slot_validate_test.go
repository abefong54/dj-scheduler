package handler

import "testing"

func TestSlotWriteRequestValidate(t *testing.T) {
	valid := slotWriteRequest{
		StageID: "stage-1", SlotDate: "2026-07-25", StartTime: "18:00", EndTime: "19:30",
	}

	tests := []struct {
		name    string
		mutate  func(r *slotWriteRequest)
		wantErr bool
	}{
		{"valid", func(r *slotWriteRequest) {}, false},
		{"valid wrapping past midnight", func(r *slotWriteRequest) { r.StartTime = "23:30"; r.EndTime = "00:30" }, false},
		{"missing stage", func(r *slotWriteRequest) { r.StageID = "" }, true},
		{"missing date", func(r *slotWriteRequest) { r.SlotDate = "" }, true},
		{"malformed start time", func(r *slotWriteRequest) { r.StartTime = "25:00" }, true},
		{"non-numeric end time", func(r *slotWriteRequest) { r.EndTime = "soon" }, true},
		{"impossible date", func(r *slotWriteRequest) { r.SlotDate = "2026-13-40" }, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := valid
			tc.mutate(&r)
			err := r.validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
