package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fhk/squidgame/pkg/parser"
	"github.com/fhk/squidgame/pkg/result"
	"github.com/fhk/squidgame/pkg/runner"
)

const asciiArt = `
 █████████   █████████  ███     ███  █████  ██████████      █████████   █████████  ██████████████  ██████████
███░░░░░███ ███░░░░░███░███    ░███ ░░███  ░░███░░░░░███    ███░░░░░███ ███░░░░░███░░███░░███░░███ ░░███░░░░░ 
░███    ░░░ ░███    ░███░███    ░███  ░███   ░███    ░███   ░███    ░░░ ░███    ░███ ░███ ░███ ░███  ░███      
░░█████████ ░███    ░███░███    ░███  ░███   ░███    ░███   ░███  █████ ░███████████ ░███ ░███ ░███  ░████████ 
 ░░░░░░░░███░███    ░███░███    ░███  ░███   ░███    ░███   ░███ ░░███  ░███░░░░░███ ░███ ░███ ░███  ░███░░░░  
 ███    ░███░░███  █████░███    ░███  ░███   ░███    ███    ░███  ░███  ░███    ░███ ░███ ░███ ░███  ░███      
░░█████████  ░░████████ ░░█████████   █████  ██████████     ░░█████████ █████   █████ █████░███ █████ ██████████
 ░░░░░░░░░    ░░░░░░░░░  ░░░░░░░░░   ░░░░░  ░░░░░░░░░░       ░░░░░░░░░  ░░░░░   ░░░░░ ░░░░░ ░░░ ░░░░░ ░░░░░░░░░░

          ○               △               □
       ------- SQUIDGAME CLI TESTING -------
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, asciiArt)
		fmt.Fprintf(os.Stderr, "\nUsage: squidgame [options] [test_dir]\n\n")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nEvery folder is a game. Survive or be ELIMINATED.")
	}

	verbose := flag.Bool("v", false, "Verbose output: show details for all assertions")
	showDiffs := flag.Bool("show-diffs", false, "Show diff commands for failed tests")
	updateExpected := flag.Bool("update-expected", false, "Copy .results/ into expected/ for all tests (pass or fail)")
	dryRun := flag.Bool("dry-run", false, "Discover and validate test configs without executing them")
	filter := flag.String("filter", "", "Only run tests whose folder names contain this substring")
	concurrency := flag.Int("j", 4, "Number of tests to run in parallel")
	flag.Parse()

	fmt.Print(asciiArt)

	rootDir := "."
	if flag.NArg() > 0 {
		rootDir = flag.Arg(0)
	}

	fmt.Printf("Running tests in %s (concurrency: %d)\n", rootDir, *concurrency)

	testDirs, err := runner.Discover(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
		os.Exit(1)
	}

	// Filter testDirs if filter is provided
	if *filter != "" {
		var filtered []string
		for _, dir := range testDirs {
			if strings.Contains(filepath.Base(dir), *filter) {
				filtered = append(filtered, dir)
			}
		}
		testDirs = filtered
	}

	if len(testDirs) == 0 {
		fmt.Println("No tests found")
		os.Exit(0)
	}

	if *dryRun {
		fmt.Printf("Found %d test(s) (dry run)\n", len(testDirs))
		invalid := 0
		for _, dir := range testDirs {
			configPath := filepath.Join(dir, "test.yaml")
			cfg, err := parser.ParseTestConfig(configPath)
			if err != nil {
				fmt.Printf("  INVALID %s: %v\n", dir, err)
				invalid++
				continue
			}
			if errs := cfg.Validate(); len(errs) > 0 {
				fmt.Printf("  INVALID %s:\n", dir)
				for _, e := range errs {
					fmt.Printf("    - %s\n", e)
				}
				invalid++
			} else {
				fmt.Printf("  OK      %s\n", dir)
			}
		}
		if invalid > 0 {
			fmt.Printf("\n%d invalid config(s)\n", invalid)
			os.Exit(1)
		}
		os.Exit(0)
	}

	start := time.Now()
	resultsChan := make(chan runner.TestResult, len(testDirs))
	dirsChan := make(chan string, len(testDirs))

	// Start workers
	for i := 0; i < *concurrency; i++ {
		go func() {
			for dir := range dirsChan {
				resultsChan <- runner.RunTest(dir, *verbose)
			}
		}()
	}

	// Send jobs
	for _, dir := range testDirs {
		dirsChan <- dir
	}
	close(dirsChan)

	var results []runner.TestResult
	for i := 0; i < len(testDirs); i++ {
		r := <-resultsChan
		results = append(results, r)
		result.PrintResult(r, *verbose, *showDiffs)

		if *updateExpected && r.Error == "" {
			if err := updateExpectedOutputs(r.TestDir); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not update expected for %s: %v\n", r.TestDir, err)
			} else {
				fmt.Printf("  Updated expected outputs for %s\n", r.TestDir)
			}
		}
	}

	result.PrintSummary(results, time.Since(start))

	for _, r := range results {
		if !r.Passed || r.Error != "" {
			os.Exit(1)
		}
	}
}

// updateExpectedOutputs copies .results/ contents into expected/.
func updateExpectedOutputs(testDir string) error {
	src := filepath.Join(testDir, ".results")
	dst := filepath.Join(testDir, "expected")

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return copyDir(src, dst)
}

// copyDir copies flat files (non-recursive) from src into dst.
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
