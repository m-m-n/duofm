package archive

import (
	"path/filepath"
	"strings"
)

// ArchiveFormat represents supported archive formats
type ArchiveFormat int

const (
	FormatUnknown ArchiveFormat = iota
	FormatTar
	FormatTarGz
	FormatTarBz2
	FormatTarXz
	FormatZip
	Format7z
)

// String returns the string representation of the archive format
func (f ArchiveFormat) String() string {
	switch f {
	case FormatTar:
		return "tar"
	case FormatTarGz:
		return "tar.gz"
	case FormatTarBz2:
		return "tar.bz2"
	case FormatTarXz:
		return "tar.xz"
	case FormatZip:
		return "zip"
	case Format7z:
		return "7z"
	default:
		return "unknown"
	}
}

// Extension returns the file extension for the archive format
func (f ArchiveFormat) Extension() string {
	switch f {
	case FormatTar:
		return ".tar"
	case FormatTarGz:
		return ".tar.gz"
	case FormatTarBz2:
		return ".tar.bz2"
	case FormatTarXz:
		return ".tar.xz"
	case FormatZip:
		return ".zip"
	case Format7z:
		return ".7z"
	default:
		return ""
	}
}

// DetectFormat determines archive format from file extension
func DetectFormat(filePath string) (ArchiveFormat, error) {
	lower := strings.ToLower(filePath)

	// Check for double extensions first (tar.gz, tar.bz2, tar.xz)
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return FormatTarGz, nil
	}
	if strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2") || strings.HasSuffix(lower, ".tbz") {
		return FormatTarBz2, nil
	}
	if strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz") {
		return FormatTarXz, nil
	}

	// Check single extensions
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".tar":
		return FormatTar, nil
	case ".zip":
		return FormatZip, nil
	case ".7z":
		return Format7z, nil
	default:
		return FormatUnknown, NewArchiveError(
			ErrArchiveUnsupportedFormat,
			"Unsupported archive format",
			nil,
		)
	}
}
