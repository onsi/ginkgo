package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ErrEmptyUnreleased is returned by Prepare when `## Unreleased` has no entries: nothing to release.
var ErrEmptyUnreleased = errors.New("## Unreleased in CHANGELOG.md has no entries - nothing to release")

// The fresh section Prepare puts back at the top of the changelog.
const freshUnreleased = "## Unreleased\n\n### Features\n\n### Fixes\n\n### Maintenance\n\n"

var versionLine = regexp.MustCompile(`(?m)^const VERSION = "([^"]*)"$`)
var semver = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

// CurrentVersion reads VERSION out of types/version.go's source.
func CurrentVersion(versionGo string) (string, error) {
	matches := versionLine.FindAllStringSubmatch(versionGo, -1)
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one `const VERSION = \"X.Y.Z\"` line in %s, found %d", versionPath, len(matches))
	}
	return matches[0][1], nil
}

// NextVersion bumps an X.Y.Z version: patch -> X.Y.(Z+1), minor -> X.(Y+1).0.
func NextVersion(current string, bump string) (string, error) {
	parts := semver.FindStringSubmatch(current)
	if parts == nil {
		return "", fmt.Errorf("VERSION %q is not a plain X.Y.Z version", current)
	}
	major, _ := strconv.Atoi(parts[1])
	minor, _ := strconv.Atoi(parts[2])
	patch, _ := strconv.Atoi(parts[3])
	switch bump {
	case "patch":
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1), nil
	case "minor":
		return fmt.Sprintf("%d.%d.0", major, minor+1), nil
	}
	return "", fmt.Errorf("bump must be patch or minor, got %q", bump)
}

// SetVersion rewrites the VERSION line in types/version.go's source, leaving every other byte alone.
func SetVersion(versionGo string, version string) (string, error) {
	if _, err := CurrentVersion(versionGo); err != nil {
		return "", err
	}
	return versionLine.ReplaceAllLiteralString(versionGo, `const VERSION = "`+version+`"`), nil
}

// Prepare turns `## Unreleased` into `## <version>` and puts a fresh, empty `## Unreleased` above
// it.  Empty `###` subsections are dropped from the released section.  Everything before and after
// the Unreleased section is left byte-for-byte as it was.
func Prepare(changelog string, version string) (string, error) {
	before, body, after, err := splitSection(changelog, "Unreleased")
	if err != nil {
		return "", err
	}
	if isEmpty(body) {
		return "", ErrEmptyUnreleased
	}
	if _, _, _, err := splitSection(changelog, version); err == nil {
		return "", fmt.Errorf("CHANGELOG.md already has a ## %s section", version)
	}
	body = dropEmptySubsections(body)
	if !strings.HasSuffix(body, "\n\n") && after != "" {
		body = strings.TrimRight(body, "\n") + "\n\n"
	}
	return before + freshUnreleased + "## " + version + "\n" + body + after, nil
}

// Notes returns the body of the `## <version>` section, trimmed of surrounding blank lines - the
// GitHub release notes.
func Notes(changelog string, version string) (string, error) {
	_, body, _, err := splitSection(changelog, version)
	if err != nil {
		return "", err
	}
	notes := strings.Trim(body, "\n")
	if notes == "" {
		return "", fmt.Errorf("## %s in CHANGELOG.md is empty", version)
	}
	return notes + "\n", nil
}

// splitSection finds the `## <name>` heading and returns the text before the heading, the section
// body (the lines after the heading, up to the next `## ` heading or the end of the file), and the
// text from the next `## ` heading on.
func splitSection(changelog string, name string) (before string, body string, after string, err error) {
	lines := strings.SplitAfter(changelog, "\n")
	offset, start := 0, -1
	for _, line := range lines {
		if start == -1 && strings.TrimRight(line, "\n") == "## "+name {
			start = offset
			offset += len(line)
			continue
		}
		if start != -1 && strings.HasPrefix(line, "## ") {
			return changelog[:start], changelog[start+len("## "+name+"\n") : offset], changelog[offset:], nil
		}
		offset += len(line)
	}
	if start == -1 {
		return "", "", "", fmt.Errorf("CHANGELOG.md has no ## %s section", name)
	}
	bodyStart := min(start+len("## "+name+"\n"), len(changelog))
	return changelog[:start], changelog[bodyStart:], "", nil
}

// isEmpty reports whether a section body holds nothing but blank lines and `###` headings.
func isEmpty(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) != "" && !isSubheading(line) {
			return false
		}
	}
	return true
}

func isSubheading(line string) bool {
	return strings.HasPrefix(line, "### ")
}

// dropEmptySubsections removes each `###` heading that is followed only by blank lines before the
// next `###` heading or the end of the body.
func dropEmptySubsections(body string) string {
	var chunks []string
	for _, line := range strings.SplitAfter(body, "\n") {
		if line == "" {
			continue
		}
		if isSubheading(line) || len(chunks) == 0 {
			chunks = append(chunks, line)
		} else {
			chunks[len(chunks)-1] += line
		}
	}
	var kept strings.Builder
	for _, chunk := range chunks {
		heading, rest, _ := strings.Cut(chunk, "\n")
		if isSubheading(heading) && strings.TrimSpace(rest) == "" {
			continue
		}
		kept.WriteString(chunk)
	}
	return kept.String()
}

var manifestVersionField = regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)

// StampManifest sets the "version" field of a Claude Code plugin manifest to version.  A manifest
// with no "version" field tracks the commit SHA instead and is left alone (stamped is false).
func StampManifest(manifest string, version string) (stamped string, ok bool) {
	if !manifestVersionField.MatchString(manifest) {
		return manifest, false
	}
	return manifestVersionField.ReplaceAllString(manifest, `${1}"`+version+`"`), true
}

// FindPluginManifests returns every plugin.json living in a .claude-plugin directory under root.
func FindPluginManifests(root string) ([]string, error) {
	manifests := []string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "vendor", "node_modules":
				if path != root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if entry.Name() == "plugin.json" && filepath.Base(filepath.Dir(path)) == ".claude-plugin" {
			manifests = append(manifests, path)
		}
		return nil
	})
	return manifests, err
}

// stampPluginManifests keeps every shipped plugin manifest's version in lockstep with the release.
// It is a no-op in a repo with no plugins.
func stampPluginManifests(version string) error {
	manifests, err := FindPluginManifests(".")
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		contents, err := os.ReadFile(manifest)
		if err != nil {
			return err
		}
		stamped, ok := StampManifest(string(contents), version)
		if !ok {
			fmt.Printf("%s has no version field (it tracks the commit SHA) - leaving it alone\n", manifest)
			continue
		}
		if err := os.WriteFile(manifest, []byte(stamped), 0o644); err != nil {
			return err
		}
		fmt.Printf("stamped %s to %s\n", manifest, version)
	}
	return nil
}
