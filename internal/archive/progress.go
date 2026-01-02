package archive

import "time"

// ProgressUpdate represents the current state of an archive operation
type ProgressUpdate struct {
	ProcessedFiles int       // Number of files processed so far
	TotalFiles     int       // Total number of files to process
	ProcessedBytes int64     // Number of bytes processed so far
	TotalBytes     int64     // Total number of bytes to process
	CurrentFile    string    // Name of the file currently being processed
	StartTime      time.Time // When the operation started
	Operation      string    // "compress" or "extract"
	ArchivePath    string    // Path to the archive being created/extracted
}

// Percentage calculates the completion percentage (0-100)
func (p *ProgressUpdate) Percentage() int {
	if p.TotalFiles == 0 {
		return 0
	}
	return (p.ProcessedFiles * 100) / p.TotalFiles
}

// ElapsedTime returns the time elapsed since the operation started
func (p *ProgressUpdate) ElapsedTime() time.Duration {
	return time.Since(p.StartTime)
}

// EstimatedRemaining estimates the remaining time based on current progress
func (p *ProgressUpdate) EstimatedRemaining() time.Duration {
	if p.ProcessedFiles == 0 || p.TotalFiles == 0 {
		return 0
	}

	elapsed := p.ElapsedTime()
	avgTimePerFile := elapsed / time.Duration(p.ProcessedFiles)
	remainingFiles := p.TotalFiles - p.ProcessedFiles

	if remainingFiles <= 0 {
		return 0
	}

	return avgTimePerFile * time.Duration(remainingFiles)
}
