package archive

import (
	"testing"
	"time"
)

func TestProgressUpdate_Percentage(t *testing.T) {
	tests := []struct {
		name           string
		processedFiles int
		totalFiles     int
		want           int
	}{
		{
			name:           "0% progress",
			processedFiles: 0,
			totalFiles:     100,
			want:           0,
		},
		{
			name:           "50% progress",
			processedFiles: 50,
			totalFiles:     100,
			want:           50,
		},
		{
			name:           "100% progress",
			processedFiles: 100,
			totalFiles:     100,
			want:           100,
		},
		{
			name:           "single file complete",
			processedFiles: 1,
			totalFiles:     1,
			want:           100,
		},
		{
			name:           "no files",
			processedFiles: 0,
			totalFiles:     0,
			want:           0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProgressUpdate{
				ProcessedFiles: tt.processedFiles,
				TotalFiles:     tt.totalFiles,
			}
			if got := p.Percentage(); got != tt.want {
				t.Errorf("ProgressUpdate.Percentage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProgressUpdate_ElapsedTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		startTime time.Time
		elapsed   time.Duration
	}{
		{
			name:      "1 second elapsed",
			startTime: now.Add(-1 * time.Second),
			elapsed:   1 * time.Second,
		},
		{
			name:      "1 minute elapsed",
			startTime: now.Add(-1 * time.Minute),
			elapsed:   1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProgressUpdate{
				StartTime: tt.startTime,
			}
			got := p.ElapsedTime()
			// Allow 100ms tolerance for test execution time
			diff := got - tt.elapsed
			if diff < 0 {
				diff = -diff
			}
			if diff > 100*time.Millisecond {
				t.Errorf("ProgressUpdate.ElapsedTime() = %v, want ~%v (diff: %v)", got, tt.elapsed, diff)
			}
		})
	}
}

func TestProgressUpdate_EstimatedRemaining(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		startTime      time.Time
		processedFiles int
		totalFiles     int
		wantApprox     time.Duration
	}{
		{
			name:           "50% done in 10s, 10s remaining",
			startTime:      now.Add(-10 * time.Second),
			processedFiles: 50,
			totalFiles:     100,
			wantApprox:     10 * time.Second,
		},
		{
			name:           "25% done in 5s, 15s remaining",
			startTime:      now.Add(-5 * time.Second),
			processedFiles: 25,
			totalFiles:     100,
			wantApprox:     15 * time.Second,
		},
		{
			name:           "no progress yet",
			startTime:      now,
			processedFiles: 0,
			totalFiles:     100,
			wantApprox:     0,
		},
		{
			name:           "completed",
			startTime:      now.Add(-10 * time.Second),
			processedFiles: 100,
			totalFiles:     100,
			wantApprox:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProgressUpdate{
				StartTime:      tt.startTime,
				ProcessedFiles: tt.processedFiles,
				TotalFiles:     tt.totalFiles,
			}
			got := p.EstimatedRemaining()
			// Allow 1 second tolerance
			diff := got - tt.wantApprox
			if diff < 0 {
				diff = -diff
			}
			if diff > 1*time.Second {
				t.Errorf("ProgressUpdate.EstimatedRemaining() = %v, want ~%v (diff: %v)", got, tt.wantApprox, diff)
			}
		})
	}
}
