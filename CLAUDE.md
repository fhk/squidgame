# Squidgame CLI Testing Framework

## Overview

Squidgame is a CLI testing framework that works on any folder structure by traversing the directory tree. Each folder represents a test case with configurable commands, assertions, setup/teardown scripts, and fixture files.

## Technology Stack

- **Language**: Go
- **Testing Strategy**:
  - Standard Go unit tests for core functionality
  - Dogfooding approach (framework tests itself using its own format)
- **Configuration Format**: YAML
- **Distribution**: Single binary

## Design Principles

1. **Separation of concerns**: Input, output, and expected results are stored separately
2. **External diff support**: Results stored in predictable locations for tools like kdiff3
3. **Plain text output**: Clean, readable output with clear pass/fail indicators
4. **Multiple assertion types**: Support for exit codes, exact matches, contains, and regex patterns
5. **Flexible setup**: Both setup/teardown scripts and fixture files

## Folder Structure

```
tests/
  test_1/
    test.yaml              # Test configuration (required)
    setup.sh               # Optional: runs before test
    teardown.sh            # Optional: runs after test
    fixtures/              # Optional: files copied to temp dir before test
      input.txt
      config.json
    expected/              # Expected outputs
      stdout.txt
      stderr.txt
      exit_code
    .results/              # Generated after test run (gitignored)
      input/               # Captured input state
        test.yaml
        fixtures/
      output/
        stdout.txt
        stderr.txt
        exit_code
      expected/            # Copy of expected for easy diffing
        stdout.txt
        stderr.txt
```

## test.yaml Configuration Format

```yaml
name: "Description of test"
command: "cli-tool --arg value"
working_dir: "."          # Optional: relative to test dir, defaults to "."
timeout: 30               # Optional: timeout in seconds, defaults to 30
env:                      # Optional: environment variables
  API_KEY: "test-key"
  DEBUG: "true"
assertions:
  - type: exit_code
    expected: 0
  - type: output_match
    stream: stdout
    expected_file: expected/stdout.txt
  - type: output_match
    stream: stderr
    expected_file: expected/stderr.txt
  - type: output_contains
    stream: stdout
    pattern: "Success"
  - type: output_regex
    stream: stdout
    pattern: "^Result: \\d+$"
  - type: output_not_contains
    stream: stderr
    pattern: "ERROR"
```

## Assertion Types

1. **exit_code**: Compares the command's exit code
   - `expected: <number>` - Expected exit code

2. **output_match**: Exact match against expected output file
   - `stream: stdout|stderr` - Which output stream to check
   - `expected_file: <path>` - Path to expected output file (relative to test dir)

3. **output_contains**: Check if output contains a pattern
   - `stream: stdout|stderr` - Which output stream to check
   - `pattern: <string>` - String that must be present in output

4. **output_not_contains**: Check if output does NOT contain a pattern
   - `stream: stdout|stderr` - Which output stream to check
   - `pattern: <string>` - String that must NOT be present in output

5. **output_regex**: Match output against regex pattern
   - `stream: stdout|stderr` - Which output stream to check
   - `pattern: <regex>` - Regular expression pattern to match

## Test Execution Flow

1. **Discovery**: Traverse directory tree to find all test.yaml files
2. **For each test**:
   1. Create temporary working directory
   2. Copy fixtures to temp dir (if fixtures/ exists)
   3. Run setup.sh in temp dir (if exists)
   4. Execute command from test.yaml
   5. Capture stdout, stderr, and exit code
   6. Run teardown.sh in temp dir (if exists)
   7. Run all assertions
   8. Save results to .results/ directory
   9. Clean up temp directory
3. **Report**: Display test results with pass/fail status

## Results Directory Structure

After a test run, the `.results/` directory contains:

- `input/` - Snapshot of test inputs for reproducibility
  - Copy of test.yaml
  - Copy of fixtures/ if present
- `output/` - Actual test outputs
  - `stdout.txt` - Captured standard output
  - `stderr.txt` - Captured standard error
  - `exit_code` - Exit code as text
- `expected/` - Copy of expected outputs for easy diffing
  - Copy of files from expected/ directory

This structure enables easy diffing with external tools:
```bash
kdiff3 .results/expected/stdout.txt .results/output/stdout.txt
diff -u .results/expected/stdout.txt .results/output/stdout.txt
```

## CLI Usage

```bash
# Run all tests in current directory
squidgame

# Run tests in specific directory
squidgame /path/to/tests

# Run specific test
squidgame /path/to/tests/test_1

# Verbose output
squidgame -v

# Show diffs on failure
squidgame --show-diffs

# Update expected outputs from actual (after manual verification)
squidgame --update-expected

# Dry run (validate test configs without running)
squidgame --dry-run
```

## Output Format

```
Running tests in /path/to/tests

test_1: Simple command execution
  ✓ exit_code: expected 0, got 0
  ✓ output_match (stdout): matches expected
  PASS (123ms)

test_2: Error handling
  ✓ exit_code: expected 1, got 1
  ✗ output_contains (stderr): expected "Error: invalid input", not found
  FAIL (45ms)

  To view diffs:
    diff -u test_2/.results/expected/stderr.txt test_2/.results/output/stderr.txt

test_3: Regex validation
  ✓ exit_code: expected 0, got 0
  ✓ output_regex (stdout): matches pattern "^Result: \d+$"
  PASS (89ms)

Summary: 2 passed, 1 failed (3 total, 257ms)
```

## Project Structure

```
squidgame/
  cmd/
    squidgame/
      main.go                 # CLI entry point
  pkg/
    runner/
      runner.go               # Test discovery and execution
      runner_test.go
    parser/
      parser.go               # YAML config parsing
      parser_test.go
    assertion/
      assertion.go            # Assertion types and validation
      assertion_test.go
    result/
      result.go               # Result formatting and output
      result_test.go
  tests/                      # Dogfooding tests
    basic_test/
      test.yaml
      expected/
        stdout.txt
    error_handling_test/
      test.yaml
      expected/
        stderr.txt
  go.mod
  go.sum
  README.md
  CLAUDE.md                   # This file
```

## Implementation Tasks

1. Set up Go project structure
2. Implement test case parser (YAML config)
3. Implement test runner core logic
4. Implement assertion types (exit code, match, contains, regex)
5. Add setup/teardown and fixture support
6. Implement result output and diff generation
7. Write Go unit tests for core functionality
8. Create dogfooding integration tests

## Future Enhancements

- Parallel test execution
- Test filtering by tags/patterns
- JUnit XML output for CI integration
- Watch mode for test development
- Interactive mode for updating expected outputs
- Coverage reporting for the CLI under test
- Test retries on failure
- Custom assertion plugins

===

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Universal Development Guidelines

### Code Quality Standards
- Write clean, readable, and maintainable code
- Follow consistent naming conventions across the project
- Use meaningful variable and function names
- Keep functions focused and single-purpose
- Add comments for complex logic and business rules

### Git Workflow
- Use descriptive commit messages following conventional commits format
- Create feature branches for new development
- Keep commits atomic and focused on single changes
- Use pull requests for code review before merging
- Maintain a clean commit history

### Documentation
- Keep README.md files up to date
- Document public APIs and interfaces
- Include usage examples for complex features
- Maintain inline code documentation
- Update documentation when making changes

### Testing Approach
- Write tests for new features and bug fixes
- Maintain good test coverage
- Use descriptive test names that explain the expected behavior
- Organize tests logically by feature or module
- Run tests before committing changes

### Security Best Practices
- Never commit sensitive information (API keys, passwords, tokens)
- Use environment variables for configuration
- Validate input data and sanitize outputs
- Follow principle of least privilege
- Keep dependencies updated

## Project Structure Guidelines

### File Organization
- Group related files in logical directories
- Use consistent file and folder naming conventions
- Separate source code from configuration files
- Keep build artifacts out of version control
- Organize assets and resources appropriately

### Configuration Management
- Use configuration files for environment-specific settings
- Centralize configuration in dedicated files
- Use environment variables for sensitive or environment-specific data
- Document configuration options and their purposes
- Provide example configuration files

## Development Workflow

### Before Starting Work
1. Pull latest changes from main branch
2. Create a new feature branch
3. Review existing code and architecture
4. Plan the implementation approach

### During Development
1. Make incremental commits with clear messages
2. Run tests frequently to catch issues early
3. Follow established coding standards
4. Update documentation as needed

### Before Submitting
1. Run full test suite
2. Check code quality and formatting
3. Update documentation if necessary
4. Create clear pull request description

## Common Patterns

### Error Handling
- Use appropriate error handling mechanisms for the language
- Provide meaningful error messages
- Log errors appropriately for debugging
- Handle edge cases gracefully
- Don't expose sensitive information in error messages

### Performance Considerations
- Profile code for performance bottlenecks
- Optimize database queries and API calls
- Use caching where appropriate
- Consider memory usage and resource management
- Monitor and measure performance metrics

### Code Reusability
- Extract common functionality into reusable modules
- Use dependency injection for better testability
- Create utility functions for repeated operations
- Design interfaces for extensibility
- Follow DRY (Don't Repeat Yourself) principle

## Review Checklist

Before marking any task as complete:
- [ ] Code follows established conventions
- [ ] Tests are written and passing
- [ ] Documentation is updated
- [ ] Security considerations are addressed
- [ ] Performance impact is considered
- [ ] Code is reviewed for maintainability
