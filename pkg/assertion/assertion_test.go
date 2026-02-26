package assertion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fhk/squidgame/pkg/parser"
)

func TestRun_ExitCode(t *testing.T) {
	a := parser.Assertion{Type: "exit_code", Expected: 0}

	if result := Run(a, "", "", "", "", 0); !result.Passed {
		t.Errorf("expected pass for exit code 0: %s", result.Message)
	}
	if result := Run(a, "", "", "", "", 1); result.Passed {
		t.Error("expected fail for exit code 1")
	}
}

func TestRun_ExitCode_Invalid(t *testing.T) {
	a := parser.Assertion{Type: "exit_code", Expected: "notanumber"}
	result := Run(a, "", "", "", "", 0)
	if result.Passed {
		t.Error("expected fail for invalid expected value")
	}
}

func TestRun_OutputContains(t *testing.T) {
	a := parser.Assertion{Type: "output_contains", Stream: "stdout", Pattern: "hello"}

	if result := Run(a, "", "", "hello world", "", 0); !result.Passed {
		t.Errorf("expected pass when output contains 'hello': %s", result.Message)
	}
	if result := Run(a, "", "", "goodbye", "", 0); result.Passed {
		t.Error("expected fail when output does not contain 'hello'")
	}
}

func TestRun_OutputNotContains(t *testing.T) {
	a := parser.Assertion{Type: "output_not_contains", Stream: "stdout", Pattern: "ERROR"}

	if result := Run(a, "", "", "success", "", 0); !result.Passed {
		t.Errorf("expected pass when output does not contain 'ERROR': %s", result.Message)
	}
	if result := Run(a, "", "", "ERROR: failed", "", 0); result.Passed {
		t.Error("expected fail when output contains 'ERROR'")
	}
}

func TestRun_OutputRegex(t *testing.T) {
	a := parser.Assertion{Type: "output_regex", Stream: "stdout", Pattern: `Result: \d+`}

	if result := Run(a, "", "", "Result: 42\n", "", 0); !result.Passed {
		t.Errorf("expected pass for matching regex: %s", result.Message)
	}
	if result := Run(a, "", "", "Result: abc", "", 0); result.Passed {
		t.Error("expected fail for non-matching regex")
	}
}

func TestRun_OutputRegex_Invalid(t *testing.T) {
	a := parser.Assertion{Type: "output_regex", Stream: "stdout", Pattern: `[invalid`}
	result := Run(a, "", "", "anything", "", 0)
	if result.Passed {
		t.Error("expected fail for invalid regex")
	}
}

func TestRun_OutputMatch(t *testing.T) {
	tmpDir := t.TempDir()
	expectedFile := filepath.Join(tmpDir, "expected", "stdout.txt")
	if err := os.MkdirAll(filepath.Dir(expectedFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(expectedFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	a := parser.Assertion{Type: "output_match", Stream: "stdout", ExpectedFile: "expected/stdout.txt"}

	if result := Run(a, tmpDir, "", "hello world\n", "", 0); !result.Passed {
		t.Errorf("expected pass for matching output: %s", result.Message)
	}
	if result := Run(a, tmpDir, "", "different\n", "", 0); result.Passed {
		t.Error("expected fail for non-matching output")
	}
}

func TestRun_OutputMatch_MissingFile(t *testing.T) {
	a := parser.Assertion{Type: "output_match", Stream: "stdout", ExpectedFile: "expected/stdout.txt"}
	result := Run(a, "/nonexistent", "", "hello", "", 0)
	if result.Passed {
		t.Error("expected fail when expected file is missing")
	}
}

func TestRun_Stderr(t *testing.T) {
	a := parser.Assertion{Type: "output_contains", Stream: "stderr", Pattern: "error"}

	if result := Run(a, "", "", "", "some error occurred", 0); !result.Passed {
		t.Errorf("expected pass when stderr contains 'error': %s", result.Message)
	}
	if result := Run(a, "", "", "error in stdout", "", 0); result.Passed {
		t.Error("expected fail when only stdout contains 'error'")
	}
}

func TestRun_UnknownType(t *testing.T) {
	a := parser.Assertion{Type: "unknown_type"}
	result := Run(a, "", "", "", "", 0)
	if result.Passed {
		t.Error("expected fail for unknown assertion type")
	}
}

// --- file_match ---

func TestRun_FileMatch(t *testing.T) {
	testDir := t.TempDir()
	workDir := t.TempDir()

	// Write actual file into workDir
	if err := os.WriteFile(filepath.Join(workDir, "result.csv"), []byte("a,b\n1,2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write matching expected file
	if err := os.MkdirAll(filepath.Join(testDir, "expected"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "expected", "result.csv"), []byte("a,b\n1,2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	a := parser.Assertion{Type: "file_match", Pattern: "result.csv", ExpectedFile: "expected/result.csv"}

	if result := Run(a, testDir, workDir, "", "", 0); !result.Passed {
		t.Errorf("expected pass for matching file: %s", result.Message)
	}
}

func TestRun_FileMatch_Mismatch(t *testing.T) {
	testDir := t.TempDir()
	workDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(workDir, "result.csv"), []byte("a,b\n1,2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(testDir, "expected"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "expected", "result.csv"), []byte("a,b\n9,9\n"), 0644); err != nil {
		t.Fatal(err)
	}

	a := parser.Assertion{Type: "file_match", Pattern: "result.csv", ExpectedFile: "expected/result.csv"}
	if result := Run(a, testDir, workDir, "", "", 0); result.Passed {
		t.Error("expected fail for mismatched file content")
	}
}

func TestRun_FileMatch_MissingActual(t *testing.T) {
	testDir := t.TempDir()
	workDir := t.TempDir()

	a := parser.Assertion{Type: "file_match", Pattern: "missing.csv", ExpectedFile: "expected/result.csv"}
	if result := Run(a, testDir, workDir, "", "", 0); result.Passed {
		t.Error("expected fail when actual file does not exist")
	}
}

// --- custom_script ---

func TestRun_CustomScript_Pass(t *testing.T) {
	testDir := t.TempDir()
	workDir := t.TempDir()

	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(filepath.Join(testDir, "compare.sh"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	a := parser.Assertion{Type: "custom_script", Pattern: "compare.sh"}
	if result := Run(a, testDir, workDir, "", "", 0); !result.Passed {
		t.Errorf("expected pass for script exiting 0: %s", result.Message)
	}
}

func TestRun_CustomScript_Fail(t *testing.T) {
	testDir := t.TempDir()
	workDir := t.TempDir()

	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(filepath.Join(testDir, "compare.sh"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	a := parser.Assertion{Type: "custom_script", Pattern: "compare.sh"}
	if result := Run(a, testDir, workDir, "", "", 0); result.Passed {
		t.Error("expected fail for script exiting 1")
	}
}

func TestRun_CustomScript_ReceivesWorkDir(t *testing.T) {
	testDir := t.TempDir()
	workDir := t.TempDir()

	// Script checks that $1 (workDir) contains a specific file
	if err := os.WriteFile(filepath.Join(workDir, "output.txt"), []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\ntest -f \"$1/output.txt\"\n"
	if err := os.WriteFile(filepath.Join(testDir, "check.sh"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	a := parser.Assertion{Type: "custom_script", Pattern: "check.sh"}
	if result := Run(a, testDir, workDir, "", "", 0); !result.Passed {
		t.Errorf("expected pass: script should find output.txt in workDir: %s", result.Message)
	}
}

// --- regex_tolerance ---

func TestRun_RegexTolerance_Pass(t *testing.T) {
	a := parser.Assertion{
		Type:      "regex_tolerance",
		Stream:    "stdout",
		Pattern:   `Value: ([\d.]+)`,
		Expected:  100.0,
		Tolerance: 1.0,
	}

	if result := Run(a, "", "", "Value: 100.5\n", "", 0); !result.Passed {
		t.Errorf("expected pass within tolerance: %s", result.Message)
	}
}

func TestRun_RegexTolerance_Fail(t *testing.T) {
	a := parser.Assertion{
		Type:      "regex_tolerance",
		Stream:    "stdout",
		Pattern:   `Value: ([\d.]+)`,
		Expected:  100.0,
		Tolerance: 0.1,
	}

	if result := Run(a, "", "", "Value: 105.0\n", "", 0); result.Passed {
		t.Error("expected fail outside tolerance")
	}
}

func TestRun_RegexTolerance_NoMatch(t *testing.T) {
	a := parser.Assertion{
		Type:      "regex_tolerance",
		Stream:    "stdout",
		Pattern:   `Value: ([\d.]+)`,
		Expected:  100.0,
		Tolerance: 1.0,
	}

	if result := Run(a, "", "", "no match here\n", "", 0); result.Passed {
		t.Error("expected fail when pattern does not match")
	}
}

func TestRun_RegexTolerance_InvalidRegex(t *testing.T) {
	a := parser.Assertion{
		Type:      "regex_tolerance",
		Stream:    "stdout",
		Pattern:   `[invalid`,
		Expected:  1.0,
		Tolerance: 0.1,
	}
	if result := Run(a, "", "", "anything", "", 0); result.Passed {
		t.Error("expected fail for invalid regex")
	}
}

// --- schema ---

func TestRun_Schema_Pass(t *testing.T) {
	stdout := "column_name,column_type,null,key,default,extra\nid,BIGINT,YES,NULL,NULL,NULL\nname,VARCHAR,YES,NULL,NULL,NULL\n"
	a := parser.Assertion{
		Type:   "schema",
		Stream: "stdout",
		Expected: map[string]interface{}{
			"id":   "BIGINT",
			"name": "VARCHAR",
		},
	}

	if result := Run(a, "", "", stdout, "", 0); !result.Passed {
		t.Errorf("expected pass for valid schema: %s", result.Message)
	}
}

func TestRun_Schema_WrongType(t *testing.T) {
	stdout := "column_name,column_type,null,key,default,extra\nid,VARCHAR,YES,NULL,NULL,NULL\n"
	a := parser.Assertion{
		Type:     "schema",
		Stream:   "stdout",
		Expected: map[string]interface{}{"id": "BIGINT"},
	}

	if result := Run(a, "", "", stdout, "", 0); result.Passed {
		t.Error("expected fail for wrong column type")
	}
}

func TestRun_Schema_MissingColumn(t *testing.T) {
	stdout := "column_name,column_type,null,key,default,extra\nid,BIGINT,YES,NULL,NULL,NULL\n"
	a := parser.Assertion{
		Type:     "schema",
		Stream:   "stdout",
		Expected: map[string]interface{}{"missing_col": "TEXT"},
	}

	if result := Run(a, "", "", stdout, "", 0); result.Passed {
		t.Error("expected fail for missing column")
	}
}

func TestRun_Schema_EmptyOutput(t *testing.T) {
	a := parser.Assertion{
		Type:     "schema",
		Stream:   "stdout",
		Expected: map[string]interface{}{"id": "BIGINT"},
	}

	if result := Run(a, "", "", "", "", 0); result.Passed {
		t.Error("expected fail for empty output")
	}
}

// --- values ---

func TestRun_Values_Pass(t *testing.T) {
	stdout := "id,score,name\n1,95.5,Alice\n2,88.0,Bob\n3,92.1,Charlie\n"
	a := parser.Assertion{
		Type:   "values",
		Stream: "stdout",
		Expected: map[string]interface{}{
			"score": map[string]interface{}{"min": 80.0, "max": 100.0},
			"name":  map[string]interface{}{"allowed": []interface{}{"Alice", "Bob", "Charlie"}},
		},
	}

	if result := Run(a, "", "", stdout, "", 0); !result.Passed {
		t.Errorf("expected pass for valid values: %s", result.Message)
	}
}

func TestRun_Values_MinViolation(t *testing.T) {
	stdout := "score\n50.0\n90.0\n"
	a := parser.Assertion{
		Type:     "values",
		Stream:   "stdout",
		Expected: map[string]interface{}{"score": map[string]interface{}{"min": 80.0}},
	}

	if result := Run(a, "", "", stdout, "", 0); result.Passed {
		t.Error("expected fail when value is below min")
	}
}

func TestRun_Values_MaxViolation(t *testing.T) {
	stdout := "score\n90.0\n110.0\n"
	a := parser.Assertion{
		Type:     "values",
		Stream:   "stdout",
		Expected: map[string]interface{}{"score": map[string]interface{}{"max": 100.0}},
	}

	if result := Run(a, "", "", stdout, "", 0); result.Passed {
		t.Error("expected fail when value exceeds max")
	}
}

func TestRun_Values_UniqueCount(t *testing.T) {
	stdout := "status\nactive\ninactive\nactive\n"
	a := parser.Assertion{
		Type:     "values",
		Stream:   "stdout",
		Expected: map[string]interface{}{"status": map[string]interface{}{"unique_count": 2}},
	}

	if result := Run(a, "", "", stdout, "", 0); !result.Passed {
		t.Errorf("expected pass for correct unique count: %s", result.Message)
	}
}

func TestRun_Values_DisallowedValue(t *testing.T) {
	stdout := "name\nAlice\nMallory\n"
	a := parser.Assertion{
		Type:     "values",
		Stream:   "stdout",
		Expected: map[string]interface{}{"name": map[string]interface{}{"allowed": []interface{}{"Alice", "Bob"}}},
	}

	if result := Run(a, "", "", stdout, "", 0); result.Passed {
		t.Error("expected fail for disallowed value")
	}
}

func TestRun_Values_MissingColumn(t *testing.T) {
	stdout := "id\n1\n2\n"
	a := parser.Assertion{
		Type:     "values",
		Stream:   "stdout",
		Expected: map[string]interface{}{"missing": map[string]interface{}{"min": 0.0}},
	}

	if result := Run(a, "", "", stdout, "", 0); result.Passed {
		t.Error("expected fail for missing column")
	}
}
