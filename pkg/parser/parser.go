package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var knownAssertionTypes = map[string]bool{
	"exit_code":          true,
	"output_match":       true,
	"output_contains":    true,
	"output_not_contains": true,
	"output_regex":       true,
	"file_match":         true,
	"custom_script":      true,
	"regex_tolerance":    true,
	"schema":             true,
	"values":             true,
}

// Assertion defines a single test assertion.
type Assertion struct {
	Type         string      `yaml:"type"`
	Expected     interface{} `yaml:"expected"`
	Stream       string      `yaml:"stream"`
	ExpectedFile string      `yaml:"expected_file"`
	Pattern      string      `yaml:"pattern"`
	Tolerance    float64     `yaml:"tolerance"`
}

// TestConfig holds the parsed test.yaml configuration.
type TestConfig struct {
	Name       string            `yaml:"name"`
	Command    string            `yaml:"command"`
	WorkingDir string            `yaml:"working_dir"`
	Timeout    int               `yaml:"timeout"`
	Env        map[string]string `yaml:"env"`
	Assertions []Assertion       `yaml:"assertions"`
}

// ParseTestConfig reads and parses a test.yaml file.
func ParseTestConfig(path string) (*TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config TestConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.WorkingDir == "" {
		config.WorkingDir = "."
	}

	return &config, nil
}

// Validate checks that a parsed config has all required fields and valid assertion types.
func (c *TestConfig) Validate() []string {
	var errs []string

	if c.Command == "" {
		errs = append(errs, "missing required field: command")
	}

	for i, a := range c.Assertions {
		if a.Type == "" {
			errs = append(errs, fmt.Sprintf("assertion[%d]: missing type", i))
			continue
		}
		if !knownAssertionTypes[a.Type] {
			errs = append(errs, fmt.Sprintf("assertion[%d]: unknown type %q", i, a.Type))
			continue
		}
		switch a.Type {
		case "exit_code":
			if a.Expected == nil {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing expected", i, a.Type))
			}
		case "output_match":
			if a.Stream == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing stream", i, a.Type))
			}
			if a.ExpectedFile == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing expected_file", i, a.Type))
			}
		case "output_contains", "output_not_contains", "output_regex":
			if a.Stream == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing stream", i, a.Type))
			}
			if a.Pattern == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing pattern", i, a.Type))
			}
		case "file_match":
			if a.Pattern == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing pattern (filename)", i, a.Type))
			}
			if a.ExpectedFile == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing expected_file", i, a.Type))
			}
		case "custom_script":
			if a.Pattern == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing pattern (script name)", i, a.Type))
			}
		case "regex_tolerance":
			if a.Stream == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing stream", i, a.Type))
			}
			if a.Pattern == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing pattern", i, a.Type))
			}
			if a.Expected == nil {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing expected", i, a.Type))
			}
		case "schema", "values":
			if a.Stream == "" {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing stream", i, a.Type))
			}
			if a.Expected == nil {
				errs = append(errs, fmt.Sprintf("assertion[%d] (%s): missing expected", i, a.Type))
			}
		}
	}

	return errs
}
