package result

import (
	"testing"
	"time"
)

func TestFormatDuration_Milliseconds(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{100 * time.Millisecond, "100ms"},
		{500 * time.Millisecond, "500ms"},
		{999 * time.Millisecond, "999ms"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatDuration_Seconds(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{1 * time.Second, "1.00s"},
		{2500 * time.Millisecond, "2.50s"},
		{10 * time.Second, "10.00s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
