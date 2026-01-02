package archive

import (
	"testing"
)

func TestCheckCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		// We can't reliably test the actual result since it depends on the system
		// Just verify it doesn't panic and returns a boolean value
	}{
		{
			name:    "check tar command",
			command: "tar",
		},
		{
			name:    "check gzip command",
			command: "gzip",
		},
		{
			name:    "check nonexistent command",
			command: "definitely_not_a_real_command_12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			_ = CheckCommand(tt.command)
		})
	}
}

func TestGetAvailableFormats(t *testing.T) {
	// This test verifies the function runs without error
	// Actual available formats depend on the system
	formats := GetAvailableFormats()

	// At minimum, tar should be available on most Linux systems
	// But we can't guarantee it in all test environments, so we just
	// verify the function returns a slice (may be empty)
	if formats == nil {
		t.Error("GetAvailableFormats() returned nil, expected non-nil slice")
	}
}

func TestIsFormatAvailable(t *testing.T) {
	// Test that the function works correctly based on GetAvailableFormats
	availableFormats := GetAvailableFormats()
	formatMap := make(map[ArchiveFormat]bool)
	for _, f := range availableFormats {
		formatMap[f] = true
	}

	tests := []struct {
		name   string
		format ArchiveFormat
	}{
		{
			name:   "check tar availability",
			format: FormatTar,
		},
		{
			name:   "check tar.gz availability",
			format: FormatTarGz,
		},
		{
			name:   "check zip availability",
			format: FormatZip,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFormatAvailable(tt.format)
			want := formatMap[tt.format]
			if got != want {
				t.Errorf("IsFormatAvailable(%v) = %v, want %v", tt.format, got, want)
			}
		})
	}
}

func TestGetRequiredCommands(t *testing.T) {
	tests := []struct {
		name   string
		format ArchiveFormat
		want   []string
	}{
		{
			name:   "tar commands",
			format: FormatTar,
			want:   []string{"tar"},
		},
		{
			name:   "tar.gz commands",
			format: FormatTarGz,
			want:   []string{"tar", "gzip"},
		},
		{
			name:   "tar.bz2 commands",
			format: FormatTarBz2,
			want:   []string{"tar", "bzip2"},
		},
		{
			name:   "tar.xz commands",
			format: FormatTarXz,
			want:   []string{"tar", "xz"},
		},
		{
			name:   "zip commands",
			format: FormatZip,
			want:   []string{"zip", "unzip"},
		},
		{
			name:   "7z commands",
			format: Format7z,
			want:   []string{"7z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRequiredCommands(tt.format)
			if len(got) != len(tt.want) {
				t.Errorf("GetRequiredCommands(%v) returned %d commands, want %d", tt.format, len(got), len(tt.want))
				return
			}
			for i, cmd := range got {
				if cmd != tt.want[i] {
					t.Errorf("GetRequiredCommands(%v)[%d] = %v, want %v", tt.format, i, cmd, tt.want[i])
				}
			}
		})
	}
}

func TestGetRequiredCommands_Unknown(t *testing.T) {
	cmds := GetRequiredCommands(FormatUnknown)
	if len(cmds) != 0 {
		t.Errorf("GetRequiredCommands(FormatUnknown) = %v, want empty slice", cmds)
	}
}

func TestIsFormatAvailable_Unknown(t *testing.T) {
	available := IsFormatAvailable(FormatUnknown)
	if available {
		t.Error("IsFormatAvailable(FormatUnknown) = true, want false")
	}
}
