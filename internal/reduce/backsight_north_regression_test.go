package reduce

import (
	"math"
	"testing"

	"CaveLoop/internal/config"
	"CaveLoop/internal/model"
)

// TestReconcilesReciprocalPairAcrossNorth covers a foresight and a backsight
// that agree perfectly but sit on opposite sides of north, plus a pair that
// really does disagree. The first pair must be averaged, the second rejected.
func TestReconcilesReciprocalPairAcrossNorth(t *testing.T) {
	back := func(value float64) *float64 { return &value }
	survey := model.Survey{
		Cave: "North Cave",
		Trips: []model.Trip{{
			ID: "T1", LengthUnit: "m", AngleUnit: "deg",
			Stations: []model.Station{{Name: "A"}, {Name: "B"}, {Name: "C"}},
			Shots: []model.Shot{
				{ID: "S1", From: "A", To: "B", Distance: 10, Azimuth: 359.7, Inclination: 0,
					BackAzimuth: back(180.1)},
				{ID: "S2", From: "B", To: "C", Distance: 10, Azimuth: 10, Inclination: 0,
					BackAzimuth: back(100)},
			},
		}},
	}
	result, err := Reduce(survey, config.Default())
	if err != nil {
		t.Fatalf("Reduce returned %v", err)
	}
	if len(result.Shots) != 2 {
		t.Fatalf("Reduce produced %d legs", len(result.Shots))
	}

	across := result.Shots[0]
	if !across.Reconciliation.HasBacksight {
		t.Fatalf("the reciprocal reading was dropped: %+v", across.Reconciliation)
	}
	if math.Abs(across.Reconciliation.AzimuthDisagreementDeg-0.4) > 1e-9 {
		t.Fatalf("the pair 359.7 / 180.1 disagrees by %v deg, want 0.4",
			across.Reconciliation.AzimuthDisagreementDeg)
	}
	if !across.Reconciliation.AzimuthAveraged || !across.Reconciliation.WithinTolerance {
		t.Fatalf("a reciprocal pair straddling north was rejected: %+v", across.Reconciliation)
	}
	if math.Abs(across.AzimuthDeg-359.9) > 1e-9 {
		t.Fatalf("the reconciled azimuth is %v deg, want 359.9", across.AzimuthDeg)
	}
	if !hasReducedNote(across.Notes, NoteBacksightAveraged) ||
		hasReducedNote(across.Notes, NoteBacksightAzimuthOutOfTolerance) {
		t.Fatalf("notes of the leg across north are %v", across.Notes)
	}

	disagreeing := result.Shots[1]
	if disagreeing.Reconciliation.AzimuthAveraged || disagreeing.Reconciliation.WithinTolerance {
		t.Fatalf("a 90 deg disagreement was accepted: %+v", disagreeing.Reconciliation)
	}
	if math.Abs(disagreeing.Reconciliation.AzimuthDisagreementDeg-90) > 1e-9 {
		t.Fatalf("the pair 10 / 100 disagrees by %v deg, want 90",
			disagreeing.Reconciliation.AzimuthDisagreementDeg)
	}
	if math.Abs(disagreeing.AzimuthDeg-10) > 1e-9 {
		t.Fatalf("a rejected backsight changed the azimuth to %v deg", disagreeing.AzimuthDeg)
	}

	for _, issue := range result.Issues {
		if issue.Path == "survey.trips[T1].shots[S1].backAzimuth" {
			t.Fatalf("the leg across north raised %s: %s", issue.Code, issue.Message)
		}
	}
	if len(result.Issues) != 1 || result.Issues[0].Code != "backsight-azimuth-disagreement" {
		t.Fatalf("issues are %v", result.Issues)
	}
}

// hasReducedNote reports whether a reduced leg carries the given note code.
func hasReducedNote(notes []string, want string) bool {
	for _, note := range notes {
		if note == want {
			return true
		}
	}
	return false
}
