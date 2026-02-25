# 🦑 Squidgame

> **"You have a chance to enter a game that could change your life. Will you participate?"**

**Squidgame** is a minimalist, high-stakes CLI testing framework designed for applications that are notoriously difficult to test:

- Mathematical Solvers (MILP, SAT)
- Large Language Models (LLMs)
- Machine Learning (ML) pipelines

In the real world, **"Success"** isn't always a binary string match. It’s a range, a tolerance, or a specific behavior hidden in a noisy log.

Squidgame provides the arena, the rules, and the **"Pink Soldier" automation** to ensure your non-deterministic apps survive the round.

---

## 🟥 The Three Pillars of the Arena

Squidgame is built to handle the **Big Three** of non-deterministic testing:

### 🧮 Mathematical Solvers (Optimization)

Extract dual bounds, optimality gaps, or objective values from:

- Gurobi
- CPLEX
- SCIP

Validate them within ± tolerances.

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

Every folder in your \`tests/\` directory is a **"Game Room."**

If the app inside follows the rules, it moves on.  
If not…

**ELIMINATED.**

---

### 1️⃣ The Challenge (\`test.yaml\`)

Define the command and the specific survival criteria.

```yaml
name: "Round 1: The MILP Solver Challenge"
command: "gurobi_cl TimeLimit=30 model.mps"
assertions:

- type: exit_code
  expected: 0
- type: regex_tolerance
  stream: stdout
  pattern: "Objective: ([-+]?\\d\*\\.?\\d+)"
  expected: 14500.50
  tolerance: 0.05 # Survived if within range
```

---

### 2️⃣ The Preparation (Scripts)

- **\`setup.sh\`** — The "Uniform."
  Prepare the environment, fetch weights, or pull matrix files.

- **\`teardown.sh\`** — The "Cleanup."
  Kill background inference servers or wipe 10GB temp files.

---

### 3️⃣ The Footage (\`.results/\`)

Win or lose, every byte is recorded.

Use your own diff tools to inspect the **"CCTV footage"** of why a player failed.

---

## 🛠 Feature Backlog (The Prize Pool)

We are building the ultimate non-deterministic engine.
Here is what's coming to the **VIP Lounge**:

---

### 📈 Resource & Performance Benchmarking

- [ ] **The "Exhaustion" Metric**
      Automatically capture peak CPU, Memory (RSS), and GPU VRAM usage for every run.

- [ ] **Regression Detection**
      Fail the test if the current run is >20% slower or heavier than the "Golden" benchmark.

- [ ] **Wall-Clock Watchdog**
      Track solver time across different hardware architectures.

---

### 🧪 Advanced Evaluation

- [ ] **LLM-as-a-Judge**
      An \`llm_eval\` assertion type to use a "Front Man" (like Claude or GPT-4) to judge the quality of an output.

- [ ] **Statistical Survival**
      Run a test N times and assert on the distribution (mean/variance) of the results.

- [ ] **JSON Schema Validation**
      Ensure ML inference outputs strictly follow a contract before checking values.

---

## 🏗 Project Structure

```text
squidgame/
├── cmd/squidgame/ # The Front Man (CLI Entry Point)
├── pkg/
│ ├── runner/ # The Guards (Execution & Discovery)
│ ├── assertion/ # The Judges (Regex & Math Logic)
│ └── result/ # The Scoreboard (Formatting & Diffs)
└── tests/ # Dogfooding (The framework testing itself)
```

---

## 🏆 CLI Usage

```bash

# Start the games in a directory

squidgame ./my_tests

# View the detailed scoreboard (Verbose)

squidgame -v

# "Change the Reality" - Update expected files after manual verification

squidgame --update-expected

# Review the footage (Open diff tool on failure)

squidgame --show-diffs
```

---

> **"Out there, the world is just as hard as it is in here. But in here, we play by the rules."**
