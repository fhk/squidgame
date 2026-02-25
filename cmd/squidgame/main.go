package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	updateExpected := flag.Bool("update-expected", false, "Update expected/ from .results/output/ on failure")
	dryRun := flag.Bool("dry-run", false, "Discover and validate test configs without executing them")
	filter := flag.String("filter", "", "Only run tests whose folder names contain this substring")
	flag.Parse()

	fmt.Print(asciiArt)

	rootDir := "."
	if flag.NArg() > 0 {
		rootDir = flag.Arg(0)
	}

	fmt.Printf("Running tests in %s\n", rootDir)

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
		for _, dir := range testDirs {
			fmt.Printf("  %s\n", dir)
		}
		os.Exit(0)
	}

	_ = verbose // reserved for future verbose-only output

	start := time.Now()
	var results []runner.TestResult

	for _, testDir := range testDirs {
		r := runner.RunTest(testDir)
		results = append(results, r)
		result.PrintResult(r, *verbose, *showDiffs)

		if *updateExpected && !r.Passed && r.Error == "" {
			if err := updateExpectedOutputs(testDir); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not update expected for %s: %v\n", testDir, err)
			} else {
				fmt.Printf("  Updated expected outputs for %s\n", testDir)
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

// updateExpectedOutputs copies all files from .results/output/* to expected/.
func updateExpectedOutputs(testDir string) error {
	src := filepath.Join(testDir, ".results", "output")
	dst := filepath.Join(testDir, "expected")

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // skip directories for now, or could use copyDir
		}
		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
