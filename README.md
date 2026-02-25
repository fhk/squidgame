# 🦑 Squidgame

> **"You have a chance to enter a game that could change your life. Will you participate?"**

```text
 █████████   █████████  ███     ███  █████  ██████████      █████████   █████████  ██████████████  ██████████
███░░░░░███ ███░░░░░███░███    ░███ ░░███  ░░███░░░░░███    ███░░░░░███ ███░░░░░███░░███░░███░░███ ░░███░░░░░ 
░███    ░░░ ░███    ░███░███    ░███  ░███   ░███    ░███   ░███    ░░░ ░███    ░███ ░███ ░███ ░███  ░███      
░░█████████ ░███    ░███░███    ░███  ░███   ░███    ░███   ░███  █████ ░███████████ ░███ ░███ ░███  ░████████ 
 ░░░░░░░░███░███    ░███░███    ░███  ░███   ░███    ░███   ░███ ░░███  ░███░░░░░███ ░███ ░███ ░███  ░███░░░░  
 ███    ░███░░███  █████░███    ░███  ░███   ░███    ███    ░███  ░███  ░███    ░███ ░███ ░███ ░███  ░███      
░░█████████  ░░████████ ░░█████████   █████  ██████████     ░░█████████ █████   █████ █████░███ █████ ██████████
 ░░░░░░░░░    ░░░░░░░░░  ░░░░░░░░░   ░░░░░  ░░░░░░░░░░       ░░░░░░░░░  ░░░░░   ░░░░░ ░░░░░ ░░░ ░░░░░ ░░░░░░░░░░

          ○               △               □
```

**Squidgame** is a minimalist, high-stakes CLI testing framework designed for applications that are notoriously difficult to test:

- Mathematical Solvers (MILP, SAT)
- Large Language Models (LLMs)
- Machine Learning (ML) pipelines

---

## 🟥 The Three Pillars of the Arena

Squidgame is built to handle the **Big Three** of non-deterministic testing:

### 🧮 Mathematical Solvers (Optimization)

Extract dual bounds, optimality gaps, or objective values from solvers like **HiGHS**, Gurobi, or SCIP. Validate them within ± tolerances.

---

### 🤖 LLMs (Generative)

- Validate structured outputs (JSON/YAML)
- Use regex to verify specific semantic markers or confidence scores

---

### 📊 Machine Learning (Inference)

- Wrap CLI-based inference engines
- Ensure model weights haven't drifted
- Verify prediction probabilities remain within expected bounds

---

## 🎮 How the Game is Played

Every folder in your `tests/` directory is a **"Game Room."**

### 1️⃣ The Challenge (`test.yaml`)

Define the command and the survival criteria.

```yaml
name: "Round 1: The MILP Solver Challenge"
command: "highs model.mps"
assertions:
  - type: exit_code
    expected: 0
  - type: output_contains
    stream: stdout
    pattern: "Model status        : Optimal"
  - type: file_match
    pattern: "solution.txt"
    expected_file: "expected/solution.txt"
```

### 🗂 Assertion Types

1.  **`exit_code`**: Compares the command's exit code.
2.  **`output_match`**: Exact match against expected output file.
3.  **`output_contains`**: Check if output contains a pattern.
4.  **`output_not_contains`**: Check if output does NOT contain a pattern.
5.  **`output_regex`**: Match output against regex pattern.
6.  **`file_match`**: Compare *any* file generated in the working directory against an expected file.
7.  **`custom_script`**: Run a custom shell script to validate results. The script receives `actual_dir` and `expected_dir` as arguments.

### 2️⃣ The Preparation (Scripts)

- **`setup.sh`** — Prepare the environment, fetch weights, or pull matrix files.
- **`teardown.sh`** — Kill background inference servers or wipe temporary files.

### 3️⃣ The Footage (`.results/`)

Win or lose, every byte is recorded. Squidgame captures **all files** generated in the temporary working directory into `.results/output/`.

---

## 🏆 CLI Usage

```bash
# Run all tests in current directory
squidgame

# Run tests in a specific directory
squidgame ./my_tests

# Filter tests by folder name
squidgame -filter solver tests/

# View detailed scoreboard (Verbose)
squidgame -v

# Show diff commands for failed tests
squidgame --show-diffs

# Update expected/ files from actual results
squidgame --update-expected
```
