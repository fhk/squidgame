package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	tmpDir := t.TempDir()

	for _, dir := range []string{"test1", "test2", "subdir/test3"} {
		d := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "test.yaml"), []byte("name: test\ncommand: echo\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Non-test directory (no test.yaml)
	if err := os.MkdirAll(filepath.Join(tmpDir, "nottest"), 0755); err != nil {
		t.Fatal(err)
	}

	dirs, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(dirs) != 3 {
		t.Errorf("expected 3 test dirs, got %d: %v", len(dirs), dirs)
	}
}

func TestDiscover_SkipsHiddenDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a visible test
	d := filepath.Join(tmpDir, "visible")
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "test.yaml"), []byte("name: test\ncommand: echo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a .results dir with a test.yaml — should be skipped
	hidden := filepath.Join(tmpDir, ".results", "inner")
	if err := os.MkdirAll(hidden, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hidden, "test.yaml"), []byte("name: hidden\ncommand: echo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(dirs) != 1 {
		t.Errorf("expected 1 test dir (hidden skipped), got %d: %v", len(dirs), dirs)
	}
}

func TestRunTest_BasicEcho(t *testing.T) {
	tmpDir := t.TempDir()

	config := `
name: "Echo test"
command: "echo hello"
assertions:
  - type: exit_code
    expected: 0
  - type: output_contains
    stream: stdout
    pattern: "hello"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	result := RunTest(tmpDir)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if !result.Passed {
		for _, a := range result.Assertions {
			if !a.Passed {
				t.Errorf("failed assertion: %s", a.Message)
			}
		}
	}
}

func TestRunTest_ExitCode(t *testing.T) {
	tmpDir := t.TempDir()

	config := `
name: "Exit code test"
command: "exit 1"
assertions:
  - type: exit_code
    expected: 1
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	result := RunTest(tmpDir)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if !result.Passed {
		t.Error("expected test to pass")
	}
}

func TestRunTest_WithFixtures(t *testing.T) {
	tmpDir := t.TempDir()

	fixturesDir := filepath.Join(tmpDir, "fixtures")
	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixturesDir, "input.txt"), []byte("test data"), 0644); err != nil {
		t.Fatal(err)
	}

	config := `
name: "Fixtures test"
command: "cat input.txt"
assertions:
  - type: exit_code
    expected: 0
  - type: output_contains
    stream: stdout
    pattern: "test data"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	result := RunTest(tmpDir)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if !result.Passed {
		for _, a := range result.Assertions {
			if !a.Passed {
				t.Errorf("failed assertion: %s", a.Message)
			}
		}
	}
}

func TestRunTest_FailingAssertion(t *testing.T) {
	tmpDir := t.TempDir()

	config := `
name: "Failing assertion"
command: "echo hello"
assertions:
  - type: output_contains
    stream: stdout
    pattern: "this is not in the output"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	result := RunTest(tmpDir)

	if result.Passed {
		t.Error("expected test to fail")
	}
}

func TestRunTest_EnvVars(t *testing.T) {
	tmpDir := t.TempDir()

	config := `
name: "Env var test"
command: "echo $MY_VAR"
env:
  MY_VAR: "custom_value"
assertions:
  - type: output_contains
    stream: stdout
    pattern: "custom_value"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	result := RunTest(tmpDir)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if !result.Passed {
		for _, a := range result.Assertions {
			if !a.Passed {
				t.Errorf("failed assertion: %s", a.Message)
			}
		}
	}
}

func TestRunTest_SavesResults(t *testing.T) {
	tmpDir := t.TempDir()

	config := `
name: "Results save test"
command: "echo output_value"
assertions:
  - type: exit_code
    expected: 0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	RunTest(tmpDir)

	// Verify .results directory was created
	for _, path := range []string{
		filepath.Join(tmpDir, ".results", "output", "stdout.txt"),
		filepath.Join(tmpDir, ".results", "output", "stderr.txt"),
		filepath.Join(tmpDir, ".results", "output", "exit_code"),
		filepath.Join(tmpDir, ".results", "input", "test.yaml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected results file %s to exist: %v", path, err)
		}
	}
}

func TestRunTest_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte("invalid: yaml: ::"), 0644); err != nil {
		t.Fatal(err)
	}

	result := RunTest(tmpDir)
	if result.Error == "" {
		t.Error("expected error for invalid config")
	}
}
