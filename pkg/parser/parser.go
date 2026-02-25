package parser

import (
	"os"

	"gopkg.in/yaml.v3"
)

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
