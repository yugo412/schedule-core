package models

import (
	"testing"
	"time"
)

func TestHasStarted(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	testCases := []struct {
		name      string
		startedAt *time.Time
		want      bool
	}{
		{
			name:      "started in the past",
			startedAt: &past,
			want:      true,
		},
		{
			name:      "starts in the future",
			startedAt: &future,
			want:      false,
		},
		{
			name:      "unknown start time",
			startedAt: nil,
			want:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			schedule := Schedule{
				StartedAt: testCase.startedAt,
			}

			if got := schedule.HasStarted(); got != testCase.want {
				t.Errorf(
					"expected %v, got %v",
					testCase.want,
					got,
				)
			}
		})
	}
}
