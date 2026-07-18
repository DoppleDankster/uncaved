// Package version exposes build metadata. The vars below are overridden at
// build time via -ldflags "-X ...", see the magefile Build target. When built
// with plain `go build` they keep their dev defaults.
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the release identity, from `git describe --always --tags`.
	Version = "dev"
	// Commit is the short git SHA the binary was built from.
	Commit = "none"
	// Date is the build timestamp in RFC3339 (UTC).
	Date = "unknown"
)

// Info returns a one-line human-readable build summary, including the OS, arch
// and Go toolchain the binary was compiled with.
func Info() string {
	return fmt.Sprintf("uncaved %s (commit %s, built %s, %s/%s, %s)",
		Version, Commit, Date, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
