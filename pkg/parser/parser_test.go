package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTestConfig(t *testing.T) {
	tmpDir := t.TempDir()

	yaml := `
name: "Test case"
command: "echo hello"
timeout: 10
env:
  FOO: "bar"
assertions:
  - type: exit_code
    expected: 0
  - type: output_contains
    stream: stdout
    pattern: "hello"
`
	path := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := ParseTestConfig(path)
	if err != nil {
		t.Fatalf("ParseTestConfig failed: %v", err)
	}

	if config.Name != "Test case" {
		t.Errorf("expected name 'Test case', got %q", config.Name)
	}
	if config.Command != "echo hello" {
		t.Errorf("expected command 'echo hello', got %q", config.Command)
	}
	if config.Timeout != 10 {
		t.Errorf("expected timeout 10, got %d", config.Timeout)
	}
	if config.Env["FOO"] != "bar" {
		t.Errorf("expected env FOO=bar, got %q", config.Env["FOO"])
	}
	if len(config.Assertions) != 2 {
		t.Errorf("expected 2 assertions, got %d", len(config.Assertions))
	}
}

func TestParseTestConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()

	yaml := `
name: "Minimal test"
command: "echo hello"
`
	path := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := ParseTestConfig(path)
	if err != nil {
		t.Fatalf("ParseTestConfig failed: %v", err)
	}

	if config.Timeout != 30 {
		t.Errorf("expected default timeout 30, got %d", config.Timeout)
	}
	if config.WorkingDir != "." {
		t.Errorf("expected default working_dir '.', got %q", config.WorkingDir)
	}
}

func TestParseTestConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(path, []byte("invalid: yaml: : :"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseTestConfig(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParseTestConfig_FileNotFound(t *testing.T) {
	_, err := ParseTestConfig("/nonexistent/test.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
