package archive

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     ArchiveFormat
		wantErr  bool
	}{
		{
			name:     "tar file",
			filePath: "test.tar",
			want:     FormatTar,
			wantErr:  false,
		},
		{
			name:     "tar.gz file",
			filePath: "test.tar.gz",
			want:     FormatTarGz,
			wantErr:  false,
		},
		{
			name:     "tar.bz2 file",
			filePath: "test.tar.bz2",
			want:     FormatTarBz2,
			wantErr:  false,
		},
		{
			name:     "tar.xz file",
			filePath: "test.tar.xz",
			want:     FormatTarXz,
			wantErr:  false,
		},
		{
			name:     "zip file",
			filePath: "test.zip",
			want:     FormatZip,
			wantErr:  false,
		},
		{
			name:     "7z file",
			filePath: "test.7z",
			want:     Format7z,
			wantErr:  false,
		},
		{
			name:     "uppercase extension",
			filePath: "TEST.TAR.GZ",
			want:     FormatTarGz,
			wantErr:  false,
		},
		{
			name:     "unsupported extension",
			filePath: "test.rar",
			want:     FormatUnknown,
			wantErr:  true,
		},
		{
			name:     "no extension",
			filePath: "testfile",
			want:     FormatUnknown,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFormat(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArchiveFormat_Extension(t *testing.T) {
	tests := []struct {
		name   string
		format ArchiveFormat
		want   string
	}{
		{
			name:   "tar extension",
			format: FormatTar,
			want:   ".tar",
		},
		{
			name:   "tar.gz extension",
			format: FormatTarGz,
			want:   ".tar.gz",
		},
		{
			name:   "tar.bz2 extension",
			format: FormatTarBz2,
			want:   ".tar.bz2",
		},
		{
			name:   "tar.xz extension",
			format: FormatTarXz,
			want:   ".tar.xz",
		},
		{
			name:   "zip extension",
			format: FormatZip,
			want:   ".zip",
		},
		{
			name:   "7z extension",
			format: Format7z,
			want:   ".7z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.Extension(); got != tt.want {
				t.Errorf("ArchiveFormat.Extension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArchiveFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format ArchiveFormat
		want   string
	}{
		{
			name:   "tar string",
			format: FormatTar,
			want:   "tar",
		},
		{
			name:   "tar.gz string",
			format: FormatTarGz,
			want:   "tar.gz",
		},
		{
			name:   "tar.bz2 string",
			format: FormatTarBz2,
			want:   "tar.bz2",
		},
		{
			name:   "tar.xz string",
			format: FormatTarXz,
			want:   "tar.xz",
		},
		{
			name:   "zip string",
			format: FormatZip,
			want:   "zip",
		},
		{
			name:   "7z string",
			format: Format7z,
			want:   "7z",
		},
		{
			name:   "unknown string",
			format: FormatUnknown,
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("ArchiveFormat.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArchiveFormat_Extension_Unknown(t *testing.T) {
	format := FormatUnknown
	ext := format.Extension()
	if ext != "" {
		t.Errorf("FormatUnknown.Extension() = %q, want empty string", ext)
	}
}

func TestDetectFormat_TGZ(t *testing.T) {
	// Test .tgz extension (alias for .tar.gz)
	format, err := DetectFormat("test.tgz")
	if err != nil {
		t.Errorf("DetectFormat() error = %v", err)
	}
	if format != FormatTarGz {
		t.Errorf("DetectFormat() = %v, want %v", format, FormatTarGz)
	}
}
