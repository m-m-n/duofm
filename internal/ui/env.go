// Package ui provides desktop environment detection for duofm.
// This module detects whether a desktop environment is available
// by checking DISPLAY and WAYLAND_DISPLAY environment variables.
package ui

import "os"

// hasDesktop caches the desktop environment detection result.
// This is determined at package initialization and cached for the lifetime of the application.
var hasDesktop = detectDesktopEnvironment()

// detectDesktopEnvironment checks environment variables to determine
// if a desktop environment is available.
func detectDesktopEnvironment() bool {
	display := os.Getenv("DISPLAY")
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	return detectDesktopEnvironmentWithValues(display, waylandDisplay)
}

// detectDesktopEnvironmentWithValues is the core detection logic.
// It returns true if either DISPLAY or WAYLAND_DISPLAY is non-empty.
// This function is separated for easier testing.
func detectDesktopEnvironmentWithValues(display, waylandDisplay string) bool {
	return display != "" || waylandDisplay != ""
}

// HasDesktopEnvironment returns whether a desktop environment is available.
// This value is cached at application startup.
func HasDesktopEnvironment() bool {
	return hasDesktop
}
