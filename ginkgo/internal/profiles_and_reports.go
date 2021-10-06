package internal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/google/pprof/profile"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

func FinalizeProfilesAndReportsForSuites(suites TestSuites, cliConfig types.CLIConfig, suiteConfig types.SuiteConfig, reporterConfig types.ReporterConfig, goFlagsConfig types.GoFlagsConfig) ([]string, error) {
	messages := []string{}
	if goFlagsConfig.Cover {
		suitesWithCoverProfiles := suites.WithState(TestSuiteStatePassed, TestSuiteStateFailed) //anything else won't have actually run and generated a cover-profile
		if cliConfig.KeepSeparateCoverprofiles {
			if cliConfig.OutputDir != "" {
				for _, suite := range suitesWithCoverProfiles {
					src := filepath.Join(suite.Path, goFlagsConfig.CoverProfile)
					dst := filepath.Join(cliConfig.OutputDir, suite.NamespacedName()+"_"+goFlagsConfig.CoverProfile)
					err := os.Rename(src, dst)
					if err != nil {
						return messages, err
					}
				}
			}
		} else {
			// merge cover profiles
			coverProfiles := []string{}
			for _, suite := range suitesWithCoverProfiles {
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

	if cliConfig.OutputDir != "" {
		suitesWithProfiles := suites.WithState(TestSuiteStatePassed, TestSuiteStateFailed) //anything else won't have actually run and generated a profile

		for _, suite := range suitesWithProfiles {
			if goFlagsConfig.BinaryMustBePreserved() {
				src := suite.PathToCompiledTest
				dst := filepath.Join(cliConfig.OutputDir, suite.NamespacedName()+".test")
				if suite.Precompiled {
					if err := CopyFile(src, dst); err != nil {
						return messages, err
					}
				} else {
					if err := os.Rename(src, dst); err != nil {
						return messages, err
					}
				}
			}
			profiles := []string{goFlagsConfig.BlockProfile, goFlagsConfig.CPUProfile, goFlagsConfig.MemProfile, goFlagsConfig.MutexProfile}
			for _, profile := range profiles {
				if profile != "" {
					src := filepath.Join(suite.Path, profile)
					dst := filepath.Join(cliConfig.OutputDir, suite.NamespacedName()+"_"+profile)
					err := os.Rename(src, dst)
					if err != nil {
						return messages, err
					}
				}
			}
		}
	}

	type reportFormat struct {
		Filename     string
		GenerateFunc func(types.Report, string) error
		MergeFunc    func([]string, string) ([]string, error)
	}
	reportFormats := []reportFormat{}
	if reporterConfig.JSONReport != "" {
		reportFormats = append(reportFormats, reportFormat{Filename: reporterConfig.JSONReport, GenerateFunc: reporters.GenerateJSONReport, MergeFunc: reporters.MergeAndCleanupJSONReports})
	}
	if reporterConfig.JUnitReport != "" {
		reportFormats = append(reportFormats, reportFormat{Filename: reporterConfig.JUnitReport, GenerateFunc: reporters.GenerateJUnitReport, MergeFunc: reporters.MergeAndCleanupJUnitReports})
	}
	if reporterConfig.TeamcityReport != "" {
		reportFormats = append(reportFormats, reportFormat{Filename: reporterConfig.TeamcityReport, GenerateFunc: reporters.GenerateTeamcityReport, MergeFunc: reporters.MergeAndCleanupTeamcityReports})
	}

	reportableSuites := suites.ThatAreGinkgoSuites()
	for _, suite := range reportableSuites.WithState(TestSuiteStateFailedToCompile, TestSuiteStateFailedDueToTimeout, TestSuiteStateSkippedDueToPriorFailures, TestSuiteStateSkippedDueToEmptyCompilation) {
		report := types.Report{
			SuitePath:      suite.AbsPath(),
			SuiteConfig:    suiteConfig,
			SuiteSucceeded: false,
		}
		switch suite.State {
		case TestSuiteStateFailedToCompile:
			report.SpecialSuiteFailureReasons = append(report.SpecialSuiteFailureReasons, suite.CompilationError.Error())
		case TestSuiteStateFailedDueToTimeout:
			report.SpecialSuiteFailureReasons = append(report.SpecialSuiteFailureReasons, TIMEOUT_ELAPSED_FAILURE_REASON)
		case TestSuiteStateSkippedDueToPriorFailures:
			report.SpecialSuiteFailureReasons = append(report.SpecialSuiteFailureReasons, PRIOR_FAILURES_FAILURE_REASON)
		case TestSuiteStateSkippedDueToEmptyCompilation:
			report.SpecialSuiteFailureReasons = append(report.SpecialSuiteFailureReasons, EMPTY_SKIP_FAILURE_REASON)
			report.SuiteSucceeded = true
		}

		for _, format := range reportFormats {
			format.GenerateFunc(report, filepath.Join(suite.Path, format.Filename))
		}
	}

	for _, format := range reportFormats {
		if cliConfig.KeepSeparateReports {
			if cliConfig.OutputDir != "" {
				// move separate reports to the output directory, appropriately namespaced
				for _, suite := range reportableSuites {
					src := filepath.Join(suite.Path, format.Filename)
					dst := filepath.Join(cliConfig.OutputDir, suite.NamespacedName()+"_"+format.Filename)
					err := os.Rename(src, dst)
					if err != nil {
						return messages, err
					}
				}
			}
		} else {
			//merge reports
			reports := []string{}
			for _, suite := range reportableSuites {
				reports = append(reports, filepath.Join(suite.Path, format.Filename))
			}
			dst := format.Filename
			if cliConfig.OutputDir != "" {
				dst = filepath.Join(cliConfig.OutputDir, format.Filename)
			}
			mergeMessages, err := format.MergeFunc(reports, dst)
			messages = append(messages, mergeMessages...)
			if err != nil {
				return messages, err
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
		contents, err := os.ReadFile(profile)
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

	err := os.WriteFile(destination, combined.Bytes(), 0666)
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

func MergeProfiles(profilePaths []string, destination string) error {
	profiles := []*profile.Profile{}
	for _, profilePath := range profilePaths {
		proFile, err := os.Open(profilePath)
		if err != nil {
			return fmt.Errorf("Could not open profile: %s\n%s", profilePath, err.Error())
		}
		prof, err := profile.Parse(proFile)
		if err != nil {
			return fmt.Errorf("Could not parse profile: %s\n%s", profilePath, err.Error())
		}
		profiles = append(profiles, prof)
		os.Remove(profilePath)
	}

	mergedProfile, err := profile.Merge(profiles)
	if err != nil {
		return fmt.Errorf("Could not merge profiles:\n%s", err.Error())
	}

	outFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Could not create merged profile %s:\n%s", destination, err.Error())
	}
	err = mergedProfile.Write(outFile)
	if err != nil {
		return fmt.Errorf("Could not write merged profile %s:\n%s", destination, err.Error())
	}
	err = outFile.Close()
	if err != nil {
		return fmt.Errorf("Could not close merged profile %s:\n%s", destination, err.Error())
	}

	return nil
}
