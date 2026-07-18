//go:build mage

// Mage build automation for the uncaved backend. Run targets with:
//
//	go tool mage <target>      # e.g. go tool mage build
//	go tool mage -l            # list targets
package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
)

const (
	binDir     = "bin"
	binName    = "uncaved"
	mainPkg    = "./cmd/uncaved"
	versionPkg = "github.com/DoppleDankster/uncaved/internal/version"
)

// Build compiles the server binary into ./bin with version metadata injected.
func Build() error {
	version, err := sh.Output("git", "describe", "--always", "--tags")
	if err != nil {
		return fmt.Errorf("git describe: %w", err)
	}
	// --always guarantees a value even with no commits/tags, but a dirty or
	// empty tree can still yield "", so fall back to something meaningful.
	if version == "" {
		version = "dev"
	}
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		commit = "none"
	}
	date := time.Now().UTC().Format(time.RFC3339)

	ldflags := strings.Join([]string{
		"-s", "-w", // strip symbol table and DWARF for a smaller binary
		flag(versionPkg, "Version", version),
		flag(versionPkg, "Commit", commit),
		flag(versionPkg, "Date", date),
	}, " ")

	out := filepath.Join(binDir, binName)
	fmt.Printf("building %s %s (%s)\n", out, version, commit)
	return sh.RunV("go", "build", "-trimpath", "-ldflags", ldflags, "-o", out, mainPkg)
}

// Clean removes build artifacts.
func Clean() error {
	return sh.Rm(binDir)
}

// flag renders a single -X linker assignment: -X pkg.Name=value.
func flag(pkg, name, value string) string {
	return fmt.Sprintf("-X %s.%s=%s", pkg, name, value)
}
