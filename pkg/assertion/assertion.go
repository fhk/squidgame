package assertion

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fhk/squidgame/pkg/parser"
)

// Result holds the outcome of a single assertion.
type Result struct {
	Passed  bool
	Message string
}

// Run evaluates a single assertion against the command output.
func Run(a parser.Assertion, testDir string, stdout, stderr string, exitCode int) Result {
	switch a.Type {
	case "exit_code":
		expected, ok := toInt(a.Expected)
		if !ok {
			return Result{false, fmt.Sprintf("exit_code: invalid expected value %v", a.Expected)}
		}
		if exitCode == expected {
			return Result{true, fmt.Sprintf("exit_code: expected %d, got %d", expected, exitCode)}
		}
		return Result{false, fmt.Sprintf("exit_code: expected %d, got %d", expected, exitCode)}

	case "output_match":
		output := getStream(a.Stream, stdout, stderr)
		expectedPath := filepath.Join(testDir, a.ExpectedFile)
		expectedBytes, err := os.ReadFile(expectedPath)
		if err != nil {
			return Result{false, fmt.Sprintf("output_match (%s): cannot read expected file: %v", a.Stream, err)}
		}
		if output == string(expectedBytes) {
			return Result{true, fmt.Sprintf("output_match (%s): matches expected", a.Stream)}
		}
		return Result{false, fmt.Sprintf("output_match (%s): output does not match expected", a.Stream)}

	case "output_contains":
		output := getStream(a.Stream, stdout, stderr)
		if strings.Contains(output, a.Pattern) {
			return Result{true, fmt.Sprintf("output_contains (%s): found %q", a.Stream, a.Pattern)}
		}
		return Result{false, fmt.Sprintf("output_contains (%s): expected %q, not found", a.Stream, a.Pattern)}

	case "output_not_contains":
		output := getStream(a.Stream, stdout, stderr)
		if !strings.Contains(output, a.Pattern) {
			return Result{true, fmt.Sprintf("output_not_contains (%s): %q not found", a.Stream, a.Pattern)}
		}
		return Result{false, fmt.Sprintf("output_not_contains (%s): %q found but should not be", a.Stream, a.Pattern)}

	case "output_regex":
		output := getStream(a.Stream, stdout, stderr)
		re, err := regexp.Compile(a.Pattern)
		if err != nil {
			return Result{false, fmt.Sprintf("output_regex (%s): invalid regex %q: %v", a.Stream, a.Pattern, err)}
		}
		if re.MatchString(output) {
			return Result{true, fmt.Sprintf("output_regex (%s): matches pattern %q", a.Stream, a.Pattern)}
		}
		return Result{false, fmt.Sprintf("output_regex (%s): does not match pattern %q", a.Stream, a.Pattern)}

	case "file_match":
		// This assertion assumes the file was created in the temp dir during RunTest
		// and has been captured into .results/output/
		// Since we don't have the temp dir path here, and saveResults already ran,
		// we can check .results/output/<filename>
		actualPath := filepath.Join(testDir, ".results", "output", a.Pattern) // Pattern used as filename for simplicity, or we could add a new field
		expectedPath := filepath.Join(testDir, a.ExpectedFile)

		actualBytes, err := os.ReadFile(actualPath)
		if err != nil {
			return Result{false, fmt.Sprintf("file_match: cannot read actual file %s: %v", a.Pattern, err)}
		}
		expectedBytes, err := os.ReadFile(expectedPath)
		if err != nil {
			return Result{false, fmt.Sprintf("file_match: cannot read expected file: %v", err)}
		}

		if string(actualBytes) == string(expectedBytes) {
			return Result{true, fmt.Sprintf("file_match: %s matches expected", a.Pattern)}
		}
		return Result{false, fmt.Sprintf("file_match: %s does not match expected", a.Pattern)}

	case "custom_script":
		scriptPath := filepath.Join(testDir, a.Pattern) // Pattern used as script name
		actualDir := filepath.Join(testDir, ".results", "output")
		expectedDir := filepath.Join(testDir, ".results", "expected")

		cmd := exec.Command("sh", scriptPath, actualDir, expectedDir)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return Result{true, fmt.Sprintf("custom_script (%s): passed", a.Pattern)}
		}
		return Result{false, fmt.Sprintf("custom_script (%s): failed with error: %v\nOutput: %s", a.Pattern, err, string(output))}

	case "regex_tolerance":
		output := getStream(a.Stream, stdout, stderr)
		re, err := regexp.Compile(a.Pattern)
		if err != nil {
			return Result{false, fmt.Sprintf("regex_tolerance (%s): invalid regex %q: %v", a.Stream, a.Pattern, err)}
		}
		matches := re.FindStringSubmatch(output)
		if len(matches) < 2 {
			return Result{false, fmt.Sprintf("regex_tolerance (%s): pattern %q matched but no capture group found", a.Stream, a.Pattern)}
		}

		actualVal, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return Result{false, fmt.Sprintf("regex_tolerance (%s): could not parse captured value %q as float: %v", a.Stream, matches[1], err)}
		}

		expectedVal, ok := toFloat64(a.Expected)
		if !ok {
			return Result{false, fmt.Sprintf("regex_tolerance (%s): invalid expected value %v", a.Stream, a.Expected)}
		}

		diff := math.Abs(actualVal - expectedVal)
		if diff <= a.Tolerance {
			return Result{true, fmt.Sprintf("regex_tolerance (%s): found %g, expected %g (diff %g <= tolerance %g)", a.Stream, actualVal, expectedVal, diff, a.Tolerance)}
		}
		return Result{false, fmt.Sprintf("regex_tolerance (%s): found %g, expected %g (diff %g > tolerance %g)", a.Stream, actualVal, expectedVal, diff, a.Tolerance)}

	default:
		return Result{false, fmt.Sprintf("unknown assertion type: %s", a.Type)}
	}
}

func getStream(stream, stdout, stderr string) string {
	if stream == "stderr" {
		return stderr
	}
	return stdout
}

func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case string:
		n, err := strconv.Atoi(val)
		return n, err == nil
	}
	return 0, false
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	}
	return 0, false
}
