//go:build mage

// Mage build automation for the uncaved backend. Run targets with:
//
//	go tool mage <target>      # e.g. go tool mage build
//	go tool mage -l            # list targets
package main

import (
	"fmt"
	"os/exec"
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

// Changelog regenerates CHANGELOG.md at the repo root from the git history
// using git-cliff (config: cliff.toml). git-cliff is a standalone binary, not
// a go tool — install it from https://git-cliff.org/docs/installation
// (e.g. `cargo install git-cliff`). The changelog is repo-wide, so this runs
// against the repository root regardless of mage's working directory.
func Changelog() error {
	if _, err := exec.LookPath("git-cliff"); err != nil {
		return fmt.Errorf("git-cliff not found on PATH; install it: https://git-cliff.org/docs/installation")
	}
	root, err := sh.Output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("locate repo root: %w", err)
	}
	return sh.RunV("git-cliff",
		"--repository", root,
		"--config", filepath.Join(root, "cliff.toml"),
		"--output", filepath.Join(root, "CHANGELOG.md"),
	)
}

// flag renders a single -X linker assignment: -X pkg.Name=value.
func flag(pkg, name, value string) string {
	return fmt.Sprintf("-X %s.%s=%s", pkg, name, value)
}
