---                                                                                                                            
  High impact, easy to fix:                                                                                                      
                                                                                                                                 
  1. Diff hints in result output are wrong — result.go still points to .results/expected/ and .results/output/ subdirs that no   
  longer exist. The diff command should just reference expected/ vs .results/ directly.                                          
  2. verbose flag does nothing — -v is wired up but the body of PrintResult ignores it. In verbose mode it should show passing   
  assertion details; currently it always shows everything the same way.                                                          
  3. --dry-run doesn't validate configs — it only discovers test.yaml files but never parses them. A typo in an assertion type
  won't be caught until the test runs.                                                                                           
  4. .results/ dirs are committed to git — there's no .gitignore excluding **/.results/. These should be gitignored.             
                                                                                                                                 
  ---                                                                                                                            
  Medium impact:                                                                                                                 
                                                                                                                                 
  5. schema assertion uses strings.Split instead of encoding/csv — values already uses the proper CSV parser (imported), but
  schema does naive splitting which would break on quoted fields.
  6. custom_script has no timeout — a hanging script blocks a goroutine forever, especially bad with parallel execution.
  7. bin/ path is resolved relative to process cwd — if you run squidgame /some/other/dir from outside the repo root,
  HiGHS/DuckDB won't be found. Should resolve relative to the executable or an env var.
  8. No ANSI color detection — color codes are always emitted, making CI logs noisy. Should check isatty or respect a NO_COLOR
  env var.
  9. --update-expected runs without confirmation on failing tests — you can accidentally overwrite your expected files with bad
  output.

  ---
  Unit test gaps:

  10. TestRunTest_SavesResults tests paths that no longer exist — it checks .results/output/stdout.txt and
  .results/input/test.yaml which are from the old nested layout.
  11. file_match, custom_script, schema, values, regex_tolerance have zero unit tests — the newer/more complex assertion types
  are entirely untested.

  ---
  Nice to have:

  12. Non-deterministic output ordering — with -j 4, results print as goroutines finish. Sorting by test name would make output
  reproducible.
  13. No file_contains / file_regex assertion — generated files can only be compared with exact file_match or a full
  custom_script. A simple contains/regex check on a file would be useful.
  14. expected/ dirs have noisy committed files — --update-expected copies everything from .results/ including fixture files
  (e.g. model.mps, data.csv) that aren't asserted against. The update should only copy stdout.txt, stderr.txt, and exit_code by
  default.

