package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fhk/squidgame/pkg/result"
	"github.com/fhk/squidgame/pkg/runner"
)

func main() {
	verbose := flag.Bool("v", false, "Verbose output")
	showDiffs := flag.Bool("show-diffs", false, "Show diff commands on failure")
	updateExpected := flag.Bool("update-expected", false, "Update expected outputs from actual outputs")
	dryRun := flag.Bool("dry-run", false, "Validate test configs without running them")
	flag.Parse()

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

// updateExpectedOutputs copies .results/output/* to expected/.
func updateExpectedOutputs(testDir string) error {
	src := testDir + "/.results/output"
	dst := testDir + "/expected"

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, name := range []string{"stdout.txt", "stderr.txt"} {
		data, err := os.ReadFile(src + "/" + name)
		if err != nil {
			continue // not all files may exist
		}
		if err := os.WriteFile(dst+"/"+name, data, 0644); err != nil {
			return err
		}
	}
	return nil
}
