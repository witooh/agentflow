package buildinfo

import "fmt"

// These variables are intended to be set via -ldflags at build time.
// Example:
//
//	go build -ldflags "-X agentflow/internal/buildinfo.Version=v0.1.0 -X agentflow/internal/buildinfo.Commit=abcdef -X agentflow/internal/buildinfo.Date=2025-09-15"
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Summary returns a one-line human-readable version string.
func Summary() string {
	return fmt.Sprintf("agentflow %s (commit %s, built %s)", Version, Commit, Date)
}
