package ui

import (
	"io/fs"
	"testing"
	"time"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"0 bytes", 0, "0 B"},
		{"1 byte", 1, "1 B"},
		{"512 bytes", 512, "512 B"},
		{"1023 bytes", 1023, "1023 B"},
		{"1 KiB", 1024, "1.0 KiB"},
		{"1.5 KiB", 1536, "1.5 KiB"},
		{"1 MiB", 1048576, "1.0 MiB"},
		{"1.2 MiB", 1258291, "1.2 MiB"},
		{"1 GiB", 1073741824, "1.0 GiB"},
		{"2.5 GiB", 2684354560, "2.5 GiB"},
		{"1 TiB", 1099511627776, "1.0 TiB"},
		{"large size", 1234567890123, "1.1 TiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "specific date time",
			time: time.Date(2024, 12, 17, 22, 28, 0, 0, time.UTC),
			want: "2024-12-17 22:28",
		},
		{
			name: "single digit month and day",
			time: time.Date(2024, 1, 5, 9, 3, 0, 0, time.UTC),
			want: "2024-01-05 09:03",
		},
		{
			name: "midnight",
			time: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			want: "2024-06-15 00:00",
		},
		{
			name: "23:59",
			time: time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC),
			want: "2024-12-31 23:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimestamp(tt.time)
			if got != tt.want {
				t.Errorf("FormatTimestamp(%v) = %q, want %q", tt.time, got, tt.want)
			}
		})
	}
}

func TestFormatPermissions(t *testing.T) {
	tests := []struct {
		name string
		mode fs.FileMode
		want string
	}{
		{
			name: "rwxrwxrwx",
			mode: 0777,
			want: "rwxrwxrwx",
		},
		{
			name: "rw-r--r--",
			mode: 0644,
			want: "rw-r--r--",
		},
		{
			name: "rwxr-xr-x",
			mode: 0755,
			want: "rwxr-xr-x",
		},
		{
			name: "---------",
			mode: 0000,
			want: "---------",
		},
		{
			name: "rwx------",
			mode: 0700,
			want: "rwx------",
		},
		{
			name: "r--r--r--",
			mode: 0444,
			want: "r--r--r--",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPermissions(tt.mode)
			if got != tt.want {
				t.Errorf("FormatPermissions(%o) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

func TestCalculateColumnWidths(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		wantHasSpace  bool
		minNameWidth  int
	}{
		{
			name:          "wide terminal",
			terminalWidth: 120,
			wantHasSpace:  true,
			minNameWidth:  60, // 広い端末では十分な幅がある
		},
		{
			name:          "narrow terminal",
			terminalWidth: 50,
			wantHasSpace:  false,
			minNameWidth:  30, // 狭い端末でも最小限の幅
		},
		{
			name:          "minimum width",
			terminalWidth: 40,
			wantHasSpace:  false,
			minNameWidth:  20, // 最小限の幅
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nameWidth, hasSpace := CalculateColumnWidths(tt.terminalWidth)
			if nameWidth < tt.minNameWidth {
				t.Errorf("CalculateColumnWidths(%d) nameWidth = %d, want >= %d", tt.terminalWidth, nameWidth, tt.minNameWidth)
			}
			if hasSpace != tt.wantHasSpace {
				t.Errorf("CalculateColumnWidths(%d) hasSpace = %v, want %v", tt.terminalWidth, hasSpace, tt.wantHasSpace)
			}
			t.Logf("Terminal width: %d, Name width: %d, Has space: %v", tt.terminalWidth, nameWidth, hasSpace)
		})
	}
}
