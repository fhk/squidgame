package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fhk/squidgame/pkg/assertion"
	"github.com/fhk/squidgame/pkg/parser"
)

// TestResult holds the outcome of a complete test run.
type TestResult struct {
	TestDir    string
	Config     *parser.TestConfig
	Passed     bool
	Assertions []assertion.Result
	Duration   time.Duration
	Error      string
}

// Discover walks rootDir and returns paths of all directories containing test.yaml.
func Discover(rootDir string) ([]string, error) {
	var testDirs []string
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories (e.g. .results)
		if d.IsDir() && len(d.Name()) > 0 && d.Name()[0] == '.' {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "test.yaml" {
			testDirs = append(testDirs, filepath.Dir(path))
		}
		return nil
	})
	return testDirs, err
}

// RunTest executes a single test and returns its result.
func RunTest(testDir string, verbose bool) TestResult {
	start := time.Now()

	configPath := filepath.Join(testDir, "test.yaml")
	config, err := parser.ParseTestConfig(configPath)
	if err != nil {
		return TestResult{TestDir: testDir, Error: fmt.Sprintf("failed to parse config: %v", err)}
	}

	// Use .results/ as the working directory: purge, recreate, copy fixtures in
	resultsDir := filepath.Join(testDir, ".results")
	if err := os.RemoveAll(resultsDir); err != nil {
		return TestResult{TestDir: testDir, Config: config, Error: fmt.Sprintf("failed to clear .results: %v", err)}
	}
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return TestResult{TestDir: testDir, Config: config, Error: fmt.Sprintf("failed to create .results: %v", err)}
	}

	// Copy fixtures into .results/
	fixturesDir := filepath.Join(testDir, "fixtures")
	if info, err := os.Stat(fixturesDir); err == nil && info.IsDir() {
		if err := copyDir(fixturesDir, resultsDir); err != nil {
			return TestResult{TestDir: testDir, Config: config, Error: fmt.Sprintf("failed to copy fixtures: %v", err)}
		}
	}

	// Run setup.sh if present
	setupScript := filepath.Join(testDir, "setup.sh")
	if _, err := os.Stat(setupScript); err == nil {
		if err := runScript(setupScript, resultsDir, config.Env, config.Timeout); err != nil {
			return TestResult{TestDir: testDir, Config: config, Error: fmt.Sprintf("setup.sh failed: %v", err)}
		}
	}

	// Execute the test command in .results/
	workDir := filepath.Join(resultsDir, config.WorkingDir)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return TestResult{TestDir: testDir, Config: config, Error: fmt.Sprintf("failed to create working dir: %v", err)}
	}
	stdout, stderr, exitCode, err := runCommand(config.Command, workDir, config.Env, config.Timeout, verbose)
	if err != nil && exitCode == -1 {
		return TestResult{TestDir: testDir, Config: config, Error: fmt.Sprintf("command execution failed: %v", err)}
	}

	// Run teardown.sh if present (ignore errors)
	teardownScript := filepath.Join(testDir, "teardown.sh")
	if _, err := os.Stat(teardownScript); err == nil {
		_ = runScript(teardownScript, resultsDir, config.Env, config.Timeout)
	}

	// Save captured streams into .results/
	_ = os.WriteFile(filepath.Join(resultsDir, "stdout.txt"), []byte(stdout), 0644)
	_ = os.WriteFile(filepath.Join(resultsDir, "stderr.txt"), []byte(stderr), 0644)
	_ = os.WriteFile(filepath.Join(resultsDir, "exit_code"), []byte(fmt.Sprintf("%d", exitCode)), 0644)

	// Evaluate assertions
	var assertionResults []assertion.Result
	allPassed := true
	for _, a := range config.Assertions {
		result := assertion.Run(a, testDir, workDir, stdout, stderr, exitCode)
		assertionResults = append(assertionResults, result)
		if !result.Passed {
			allPassed = false
		}
	}

	return TestResult{
		TestDir:    testDir,
		Config:     config,
		Passed:     allPassed,
		Assertions: assertionResults,
		Duration:   time.Since(start),
	}
}

func runCommand(command, workDir string, env map[string]string, timeout int, verbose bool) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workDir

	// Setup environment: include project bin/ in PATH
	absBin, _ := filepath.Abs("bin")
	cmd.Env = os.Environ()
	pathSet := false
	for i, envVar := range cmd.Env {
		if strings.HasPrefix(envVar, "PATH=") {
			cmd.Env[i] = fmt.Sprintf("PATH=%s:%s", absBin, envVar[5:])
			pathSet = true
			break
		}
	}
	if !pathSet {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", absBin))
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	if verbose {
		cmd.Stdout = io.MultiWriter(&stdoutBuf, os.Stdout)
		cmd.Stderr = io.MultiWriter(&stderrBuf, os.Stderr)
	} else {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", "", -1, err
		}
	}
	return stdoutBuf.String(), stderrBuf.String(), exitCode, nil
}

func runScript(scriptPath, workDir string, env map[string]string, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", scriptPath)
	cmd.Dir = workDir

	// Setup environment: include project bin/ in PATH
	absBin, _ := filepath.Abs("bin")
	cmd.Env = os.Environ()
	pathSet := false
	for i, envVar := range cmd.Env {
		if strings.HasPrefix(envVar, "PATH=") {
			cmd.Env[i] = fmt.Sprintf("PATH=%s:%s", absBin, envVar[5:])
			pathSet = true
			break
		}
	}
	if !pathSet {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", absBin))
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("exited with code %d", exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

