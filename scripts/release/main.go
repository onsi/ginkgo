// Command release holds the version and changelog steps of scripts/release.sh, which runs it from
// the repository root:
//
//	go run ./scripts/release current            print VERSION
//	go run ./scripts/release next patch|minor   print the version a bump would produce
//	go run ./scripts/release prepare X.Y.Z      release ## Unreleased as ## X.Y.Z, set VERSION, and
//	                                            stamp the plugin manifests
//	go run ./scripts/release notes X.Y.Z        print the body of ## X.Y.Z (the GitHub release notes)
//
// See RELEASING.md.
package main

import (
	"fmt"
	"os"
)

const versionPath = "types/version.go"
const changelogPath = "CHANGELOG.md"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "release:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: release current | next patch|minor | prepare X.Y.Z | notes X.Y.Z")
	}
	command, args := args[0], args[1:]
	switch {
	case command == "current" && len(args) == 0:
		current, err := readVersion()
		if err != nil {
			return err
		}
		fmt.Println(current)
		return nil
	case command == "next" && len(args) == 1:
		current, err := readVersion()
		if err != nil {
			return err
		}
		next, err := NextVersion(current, args[0])
		if err != nil {
			return err
		}
		fmt.Println(next)
		return nil
	case command == "prepare" && len(args) == 1:
		return prepare(args[0])
	case command == "notes" && len(args) == 1:
		changelog, err := os.ReadFile(changelogPath)
		if err != nil {
			return err
		}
		notes, err := Notes(string(changelog), args[0])
		if err != nil {
			return err
		}
		fmt.Print(notes)
		return nil
	}
	return fmt.Errorf("unknown or malformed command: %q", append([]string{command}, args...))
}

func readVersion() (string, error) {
	source, err := os.ReadFile(versionPath)
	if err != nil {
		return "", err
	}
	return CurrentVersion(string(source))
}

// prepare computes both rewrites before writing either file, so a failure leaves the tree untouched.
func prepare(version string) error {
	if !semver.MatchString(version) {
		return fmt.Errorf("%q is not a plain X.Y.Z version", version)
	}
	changelog, err := os.ReadFile(changelogPath)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(versionPath)
	if err != nil {
		return err
	}
	newChangelog, err := Prepare(string(changelog), version)
	if err != nil {
		return err
	}
	newSource, err := SetVersion(string(source), version)
	if err != nil {
		return err
	}
	if err := os.WriteFile(changelogPath, []byte(newChangelog), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(versionPath, []byte(newSource), 0o644); err != nil {
		return err
	}
	return stampPluginManifests(version)
}
