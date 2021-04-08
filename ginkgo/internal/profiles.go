package internal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/onsi/ginkgo/config"
)

func FinalizeProfilesForSuites(suites []TestSuite, cliConfig config.GinkgoCLIConfigType, goFlagsConfig config.GoFlagsConfigType) ([]string, error) {
	messages := []string{}
	if goFlagsConfig.Cover {
		if cliConfig.KeepSeparateCoverprofiles {
			if cliConfig.OutputDir != "" {
				// move separate cover profiles to the output directory, appropriately namespaced
				for _, suite := range suites {
					src := filepath.Join(suite.Path, goFlagsConfig.CoverProfile)
					dst := filepath.Join(cliConfig.OutputDir, suite.NamespacedName()+"_"+goFlagsConfig.CoverProfile)
					fmt.Println(src, dst)
					err := os.Rename(src, dst)
					if err != nil {
						return messages, err
					}
				}
			}
		} else {
			// merge cover profiles
			coverProfiles := []string{}
			for _, suite := range suites {
				coverProfiles = append(coverProfiles, filepath.Join(suite.Path, goFlagsConfig.CoverProfile))
			}
			dst := goFlagsConfig.CoverProfile
			if cliConfig.OutputDir != "" {
				dst = filepath.Join(cliConfig.OutputDir, goFlagsConfig.CoverProfile)
			}
			err := MergeAndCleanupCoverProfiles(coverProfiles, dst)
			if err != nil {
				return messages, err
			}

			coverage, err := GetCoverageFromCoverProfile(dst)
			if err != nil {
				return messages, err
			}
			if coverage == 0 {
				messages = append(messages, "composite coverage: [no statements]")
			} else {
				messages = append(messages, fmt.Sprintf("composite coverage: %.1f%% of statements", coverage))
			}

		}
	}

	return messages, nil
}

//loads each profile, combines them, deletes them, stores them in destination
func MergeAndCleanupCoverProfiles(profiles []string, destination string) error {
	combined := &bytes.Buffer{}
	modeRegex := regexp.MustCompile(`^mode: .*\n`)
	for i, profile := range profiles {
		contents, err := ioutil.ReadFile(profile)
		if err != nil {
			return fmt.Errorf("Unable to read coverage file %s:\n%s", profile, err.Error())
		}
		os.Remove(profile)

		// remove the cover mode line from every file
		// except the first one
		if i > 0 {
			contents = modeRegex.ReplaceAll(contents, []byte{})
		}

		_, err = combined.Write(contents)

		// Add a newline to the end of every file if missing.
		if err == nil && len(contents) > 0 && contents[len(contents)-1] != '\n' {
			_, err = combined.Write([]byte("\n"))
		}

		if err != nil {
			return fmt.Errorf("Unable to append to coverprofile:\n%s", err.Error())
		}
	}

	err := ioutil.WriteFile(destination, combined.Bytes(), 0666)
	if err != nil {
		return fmt.Errorf("Unable to create combined cover profile:\n%s", err.Error())
	}
	return nil
}

func GetCoverageFromCoverProfile(profile string) (float64, error) {
	cmd := exec.Command("go", "tool", "cover", "-func", profile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("Could not process Coverprofile %s: %s", profile, err.Error())
	}
	re := regexp.MustCompile(`total:\s*\(statements\)\s*(\d*\.\d*)\%`)
	matches := re.FindStringSubmatch(string(output))
	if matches == nil {
		return 0, fmt.Errorf("Could not parse Coverprofile to compute coverage percentage")
	}
	coverageString := matches[1]
	coverage, err := strconv.ParseFloat(coverageString, 64)
	if err != nil {
		return 0, fmt.Errorf("Could not parse Coverprofile to compute coverage percentage: %s", err.Error())
	}

	return coverage, nil
}
