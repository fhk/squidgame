package assertion

import (
	"encoding/csv"
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
func Run(a parser.Assertion, testDir, workDir string, stdout, stderr string, exitCode int) Result {
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
		actualPath := filepath.Join(workDir, a.Pattern)
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
		scriptPath := filepath.Join(testDir, a.Pattern)
		expectedDir := filepath.Join(testDir, "expected")

		cmd := exec.Command("sh", scriptPath, workDir, expectedDir)
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

	case "schema":
		output := getStream(a.Stream, stdout, stderr)
		expectedSchema, ok := a.Expected.(map[string]interface{})
		if !ok {
			// YAML might unmarshal into map[interface{}]interface{}
			// Let's try to convert it if possible or handle the most common case
			return Result{false, "schema: expected value must be a map of column:type"}
		}

		// Simple CSV parsing for DuckDB DESCRIBE output
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) < 2 {
			return Result{false, "schema: output is empty or missing header"}
		}

		actualSchema := make(map[string]string)
		for i, line := range lines {
			if i == 0 {
				continue // Skip header: column_name,column_type,null,key,default,extra
			}
			parts := strings.Split(line, ",")
			if len(parts) >= 2 {
				// DuckDB DESCRIBE -csv output: column_name is parts[0], type is parts[1]
				actualSchema[parts[0]] = parts[1]
			}
		}

		for col, expType := range expectedSchema {
			actualType, exists := actualSchema[col]
			if !exists {
				return Result{false, fmt.Sprintf("schema: column %q not found in output", col)}
			}
			if actualType != fmt.Sprintf("%v", expType) {
				return Result{false, fmt.Sprintf("schema: column %q expected type %v, got %s", col, expType, actualType)}
			}
		}
		return Result{true, "schema: all specified columns matched expected types"}

	case "values":
		output := getStream(a.Stream, stdout, stderr)
		reader := csv.NewReader(strings.NewReader(strings.TrimSpace(output)))
		records, err := reader.ReadAll()
		if err != nil {
			return Result{false, fmt.Sprintf("values: failed to parse CSV: %v", err)}
		}
		if len(records) < 2 {
			return Result{false, "values: output has no data rows"}
		}

		header := records[0]
		data := records[1:]
		colIndices := make(map[string]int)
		for i, name := range header {
			colIndices[name] = i
		}

		constraints, ok := a.Expected.(map[string]interface{})
		if !ok {
			return Result{false, "values: expected must be a map of column:constraints"}
		}

		for colName, colConstRaw := range constraints {
			idx, exists := colIndices[colName]
			if !exists {
				return Result{false, fmt.Sprintf("values: column %q not found", colName)}
			}

			colConst, ok := colConstRaw.(map[string]interface{})
			if !ok {
				return Result{false, fmt.Sprintf("values: constraints for %q must be a map", colName)}
			}

			uniqueValues := make(map[string]bool)
			minVal := math.MaxFloat64
			maxVal := -math.MaxFloat64
			hasNumeric := false

			for _, row := range data {
				val := row[idx]
				uniqueValues[val] = true

				if f, err := strconv.ParseFloat(val, 64); err == nil {
					hasNumeric = true
					if f < minVal {
						minVal = f
					}
					if f > maxVal {
						maxVal = f
					}
				}

				// Check allowed values immediately if specified
				if allowedRaw, ok := colConst["allowed"]; ok {
					allowedList, _ := allowedRaw.([]interface{})
					found := false
					for _, a := range allowedList {
						if val == fmt.Sprintf("%v", a) {
							found = true
							break
						}
					}
					if !found {
						return Result{false, fmt.Sprintf("values: column %q contains unauthorized value %q", colName, val)}
					}
				}
			}

			// Final checks for min/max/unique_count
			if m, ok := toFloat64(colConst["min"]); ok {
				if !hasNumeric || minVal < m {
					return Result{false, fmt.Sprintf("values: column %q min value %g is less than expected %g", colName, minVal, m)}
				}
			}
			if m, ok := toFloat64(colConst["max"]); ok {
				if !hasNumeric || maxVal > m {
					return Result{false, fmt.Sprintf("values: column %q max value %g is greater than expected %g", colName, maxVal, m)}
				}
			}
			if uc, ok := toInt(colConst["unique_count"]); ok {
				if len(uniqueValues) != uc {
					return Result{false, fmt.Sprintf("values: column %q expected %d unique values, got %d", colName, uc, len(uniqueValues))}
				}
			}
		}

		return Result{true, "values: all data constraints satisfied"}

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
