package assertion

import (
	"fmt"
	"os"
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
