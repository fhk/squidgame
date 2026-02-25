package assertion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fhk/squidgame/pkg/parser"
)

func TestRun_ExitCode(t *testing.T) {
	a := parser.Assertion{Type: "exit_code", Expected: 0}

	if result := Run(a, "", "", "", 0); !result.Passed {
		t.Errorf("expected pass for exit code 0: %s", result.Message)
	}
	if result := Run(a, "", "", "", 1); result.Passed {
		t.Error("expected fail for exit code 1")
	}
}

func TestRun_ExitCode_Invalid(t *testing.T) {
	a := parser.Assertion{Type: "exit_code", Expected: "notanumber"}
	result := Run(a, "", "", "", 0)
	if result.Passed {
		t.Error("expected fail for invalid expected value")
	}
}

func TestRun_OutputContains(t *testing.T) {
	a := parser.Assertion{Type: "output_contains", Stream: "stdout", Pattern: "hello"}

	if result := Run(a, "", "hello world", "", 0); !result.Passed {
		t.Errorf("expected pass when output contains 'hello': %s", result.Message)
	}
	if result := Run(a, "", "goodbye", "", 0); result.Passed {
		t.Error("expected fail when output does not contain 'hello'")
	}
}

func TestRun_OutputNotContains(t *testing.T) {
	a := parser.Assertion{Type: "output_not_contains", Stream: "stdout", Pattern: "ERROR"}

	if result := Run(a, "", "success", "", 0); !result.Passed {
		t.Errorf("expected pass when output does not contain 'ERROR': %s", result.Message)
	}
	if result := Run(a, "", "ERROR: failed", "", 0); result.Passed {
		t.Error("expected fail when output contains 'ERROR'")
	}
}

func TestRun_OutputRegex(t *testing.T) {
	a := parser.Assertion{Type: "output_regex", Stream: "stdout", Pattern: `Result: \d+`}

	if result := Run(a, "", "Result: 42\n", "", 0); !result.Passed {
		t.Errorf("expected pass for matching regex: %s", result.Message)
	}
	if result := Run(a, "", "Result: abc", "", 0); result.Passed {
		t.Error("expected fail for non-matching regex")
	}
}

func TestRun_OutputRegex_Invalid(t *testing.T) {
	a := parser.Assertion{Type: "output_regex", Stream: "stdout", Pattern: `[invalid`}
	result := Run(a, "", "anything", "", 0)
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

	if result := Run(a, tmpDir, "hello world\n", "", 0); !result.Passed {
		t.Errorf("expected pass for matching output: %s", result.Message)
	}
	if result := Run(a, tmpDir, "different\n", "", 0); result.Passed {
		t.Error("expected fail for non-matching output")
	}
}

func TestRun_OutputMatch_MissingFile(t *testing.T) {
	a := parser.Assertion{Type: "output_match", Stream: "stdout", ExpectedFile: "expected/stdout.txt"}
	result := Run(a, "/nonexistent", "hello", "", 0)
	if result.Passed {
		t.Error("expected fail when expected file is missing")
	}
}

func TestRun_Stderr(t *testing.T) {
	a := parser.Assertion{Type: "output_contains", Stream: "stderr", Pattern: "error"}

	if result := Run(a, "", "", "some error occurred", 0); !result.Passed {
		t.Errorf("expected pass when stderr contains 'error': %s", result.Message)
	}
	if result := Run(a, "", "error in stdout", "", 0); result.Passed {
		t.Error("expected fail when only stdout contains 'error'")
	}
}

func TestRun_UnknownType(t *testing.T) {
	a := parser.Assertion{Type: "unknown_type"}
	result := Run(a, "", "", "", 0)
	if result.Passed {
		t.Error("expected fail for unknown assertion type")
	}
}
