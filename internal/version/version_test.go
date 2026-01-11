package version

import (
	"testing"
)

func TestVersion_DefaultValue(t *testing.T) {
	// The default value should be "dev" when not set via ldflags
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Default value when not injected at build time
	if Version != "dev" {
		// This is not necessarily an error, as Version might be set at build time
		t.Logf("Version is set to: %s", Version)
	}
}

func TestVersion_IsString(t *testing.T) {
	// Verify Version can be used as a string
	var _ string = Version

	// Verify it's not empty
	if len(Version) == 0 {
		t.Error("Version should have non-zero length")
	}
}

func TestVersion_Comparable(t *testing.T) {
	// Test that Version can be compared
	if Version == "dev" {
		t.Log("Running with default development version")
	}

	// Test string operations work
	if Version != "" && Version[0] != 0 {
		t.Logf("Version starts with: %c", Version[0])
	}
}

func TestVersion_CanBeAssigned(t *testing.T) {
	// Store original value
	original := Version

	// Verify it can be assigned (this is how ldflags work)
	Version = "1.0.0"
	if Version != "1.0.0" {
		t.Errorf("Version assignment failed, got: %s", Version)
	}

	// Restore original value
	Version = original
}
