package archive

import "os/exec"

// formatCommands maps each archive format to its required commands
var formatCommands = map[ArchiveFormat][]string{
	FormatTar:    {"tar"},
	FormatTarGz:  {"tar", "gzip"},
	FormatTarBz2: {"tar", "bzip2"},
	FormatTarXz:  {"tar", "xz"},
	FormatZip:    {"zip", "unzip"},
	Format7z:     {"7z"},
}

// CheckCommand checks if a command is available in the system PATH
func CheckCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetRequiredCommands returns the list of commands required for a format
func GetRequiredCommands(format ArchiveFormat) []string {
	if cmds, ok := formatCommands[format]; ok {
		// Return a copy to prevent modification
		result := make([]string, len(cmds))
		copy(result, cmds)
		return result
	}
	return []string{}
}

// IsFormatAvailable checks if all required commands for a format are available
func IsFormatAvailable(format ArchiveFormat) bool {
	cmds, ok := formatCommands[format]
	if !ok {
		return false
	}

	for _, cmd := range cmds {
		if !CheckCommand(cmd) {
			return false
		}
	}
	return true
}

// GetAvailableFormats returns a list of formats that are available on the system
func GetAvailableFormats() []ArchiveFormat {
	formats := []ArchiveFormat{
		FormatTar,
		FormatTarGz,
		FormatTarBz2,
		FormatTarXz,
		FormatZip,
		Format7z,
	}

	var available []ArchiveFormat
	for _, format := range formats {
		if IsFormatAvailable(format) {
			available = append(available, format)
		}
	}

	return available
}
