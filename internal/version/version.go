// Package version provides the application version information.
// The Version variable is set at build time via ldflags.
package version

// Version is the application version, set via ldflags at build time.
// Default value "dev" is used when not injected.
var Version = "dev"
