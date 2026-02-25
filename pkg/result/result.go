package result

import (
	"fmt"
	"time"

	"github.com/fhk/squidgame/pkg/runner"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

// PrintResult writes a single test result to stdout.
func PrintResult(r runner.TestResult, verbose, showDiffs bool) {
	name := "unknown"
	if r.Config != nil {
		name = r.Config.Name
	}
	fmt.Printf("\n%s: %s\n", r.TestDir, name)

	if r.Error != "" {
		fmt.Printf("  %sERROR: %s%s\n", colorRed, r.Error, colorReset)
		fmt.Printf("  %sFAIL%s\n", colorRed, colorReset)
		return
	}

	for _, a := range r.Assertions {
		mark := "✓"
		color := colorGreen
		if !a.Passed {
			mark = "✗"
			color = colorRed
		}
		fmt.Printf("  %s%s %s%s\n", color, mark, a.Message, colorReset)
	}

	if r.Passed {
		fmt.Printf("  %sPASS%s (%s)\n", colorGreen, colorReset, formatDuration(r.Duration))
	} else {
		fmt.Printf("  %sFAIL%s (%s)\n", colorRed, colorReset, formatDuration(r.Duration))
		if showDiffs {
			fmt.Printf("\n  To view diffs:\n")
			fmt.Printf("    diff -u %s/.results/expected/stdout.txt %s/.results/output/stdout.txt\n", r.TestDir, r.TestDir)
			fmt.Printf("    diff -u %s/.results/expected/stderr.txt %s/.results/output/stderr.txt\n", r.TestDir, r.TestDir)
		}
	}
}

// PrintSummary writes the overall test run summary to stdout.
func PrintSummary(results []runner.TestResult, total time.Duration) {
	passed, failed := 0, 0
	for _, r := range results {
		if r.Passed && r.Error == "" {
			passed++
		} else {
			failed++
		}
	}
	fmt.Printf("\nSummary: %d passed, %d failed (%d total, %s)\n",
		passed, failed, len(results), formatDuration(total))
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
