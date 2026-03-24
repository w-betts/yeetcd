# yeetcd

Agent-friendly cd

## Multi-Workflow Agent System

This project provides two distinct development workflows through the `agent` wrapper script:

1. **Spec Workflow** (`agent spec`) - Structured workflow for complex features with phase-based execution
2. **Vibe Workflow** (`agent vibe`) - Direct implementation for rapid iteration

### Quick Start

```bash
# For complex features requiring planning and architecture
agent spec

# For quick implementation and rapid iteration
agent vibe
```

On Windows, use `agent.bat` instead:
```cmd
agent.bat spec
agent.bat vibe
```

---

## The Spec Workflow (Phase-Based Execution)

The **spec** workflow is designed for complex features that require careful planning, architecture decisions, and comprehensive testing. It follows a structured approach where the spec agent creates a high-level plan, orchestrates subagents for planning and review, and gets user approval before implementation.

### Phase 1: Understand & Create High-Level Spec

- **Agent**: Spec agent (primary orchestrator)
- **Interaction**: Direct conversation with user
- **Output**: Structured YAML spec file (status: "draft")
- **Actions**:
  - **Problem Understanding**:
    - Play back understanding of what user wants to build
    - Ask clarifying questions to get really clear on details and scope
    - Highlight ambiguities and edge cases
    - Confirm understanding before moving to planning
   - **High-Level Planning** (spec agent does this, NOT planner):
     - Analyze the problem and propose architecture
     - Identify if work should be split into multiple phases:
       - **Distinct components**: Work spanning multiple independent components
       - **Release boundaries**: CRITICAL - Backwards compatibility concerns, API changes requiring deployment before client updates, database migrations, or deployment checkpoints
       - **Risk mitigation**: Early phases can validate assumptions
     - **Release Boundary Identification**:
       - A release boundary marks a point where changes MUST be deployed/released before subsequent phases can proceed
       - Examples: API contract changes, database schema migrations, breaking changes requiring coordination
       - Phases must be ordered so release boundaries come BEFORE phases that depend on them
       - Mark phases with `is_release_boundary: true` when they contain changes that must be released first
     - If single phase is sufficient, create spec with one phase
     - Write spec via `spec_write` tool with:
       - Problem statement and goals
       - Constraints and tech choices
       - Architecture and components
       - Test strategy and patterns
       - Phases with high-level descriptions (file_changes and test_cases empty at this stage)
       - Status: "draft"

### Phase 2: Plan All Phases (Delegate to @planner)

- **Agent**: Spec agent (delegates to @planner)
- **Output**: Spec with all phases filled in (status: "planned")
- **Actions**:
  - Spec agent invokes @planner subagent with:
    - The spec path
    - Instructions to fill in file_changes and test_cases for ALL phases
  - The planner will:
    - Read spec via `spec_read`
    - Analyze codebase
    - Define specific file_changes and test_cases for each phase
    - Update each phase via `spec_update`
  - Spec agent reviews planner's output
  - If changes needed, loop back to planner
  - Once all phases are planned, update spec status to "planned"

### Phase 3: Adversarial Review (Delegate to @reviewer)

- **Agent**: Spec agent (delegates to @reviewer)
- **Output**: Spec with review recorded (status: "reviewed" if passed)
- **Actions**:
  - Spec agent invokes @reviewer subagent with:
    - The spec path
  - The reviewer will:
    - Read spec via `spec_read`
    - Examine codebase to inform review
    - Review spec for:
      - **Incorrectness**: Does the plan contradict the problem statement?
      - **Incompleteness**: Are there missing phases, file changes, or test cases?
      - **Over-complexity**: Is the plan unnecessarily complicated?
    - Record review via `spec_update` (only reviewer can set review status)
  - Spec agent reviews reviewer's output:
    - If review passed: Update spec status to "reviewed" and proceed to Phase 4
    - If review failed:
      - Re-invoke @planner with review feedback to address issues
      - Once planner completes, re-invoke @reviewer for re-review
      - Loop until review passes

### Phase 4: User Approval (CRITICAL)

- **Agent**: Spec agent
- **Interaction**: Direct conversation with user
- **Output**: Approved spec (status: "approved")
- **Actions**:
  - Present complete spec to user (including all planned phases and review results)
  - Use question tool to ask: "Are you happy with this spec? Should I proceed with implementation?"
  - User MUST explicitly confirm before proceeding
  - If user requests changes:
    - Update spec as needed
    - Re-invoke @planner if phase details need updating
    - Re-invoke @reviewer for re-review
    - Loop back to this phase for user approval
  - Once approved, update spec status to "approved"

### Phase 5: Execute Phases Iteratively

For each phase (up to the next release boundary, if any):

**5a. Write Tests (Delegate to @test-writer)**:
- Spec agent updates phase status to "in_progress" via `spec_update`
- Spec agent invokes @test-writer subagent with:
  - The spec path
  - The phase index to implement
- The test-writer will:
  - Read spec via `spec_read`
  - Write test files for the phase's test_cases
  - Verify tests compile/run
- Spec agent reviews test-writer's output

**5b. Implement (Delegate to @implementer)**:
- Spec agent invokes @implementer subagent with:
  - The spec path
  - The phase index to implement
- The implementer will:
  - Read spec via `spec_read`
  - Write implementation code for the phase's file_changes
  - Run tests and apply trivial fixes
  - Report any non-trivial issues
- If non-trivial issues arise:
  - Report clearly to user
  - Loop back to Phase 1 for spec revision
- Once phase is complete:
  - Update phase status to "completed" via `spec_update`

**5c. Release Boundary Check (MANDATORY STOP)**:
- If the phase has `is_release_boundary: true`:
  - **STOP IMPLEMENTATION IMMEDIATELY**
  - Inform the user that this phase marks a release boundary
  - Explain that ALL changes from this phase (and any prior phases since the last release boundary) must be fully released before continuing
  - "Fully released" means: deployed to production, merged to main branch, or otherwise made available as specified by the user
  - Ask the user to confirm when the release is complete
  - Update the phase status to "released" via `spec_update`
  - Only after explicit user confirmation of release, continue to the next phase
  - If the user cannot confirm release, STOP and report progress - do not proceed to the next phase

### Phase 6: Completion

- When all phases are complete, update spec status to "completed"
- Report final status to user

### When to Use Spec Workflow

Use `agent spec` when:
- Building complex features requiring architecture decisions
- Multiple components need coordination
- You need comprehensive test coverage
- The problem requires careful analysis before implementation
- Working with unfamiliar domains or technologies
- Building production-ready features
- Work spans multiple phases or release boundaries
- Quality assurance through adversarial review is needed

---

## The Vibe Workflow (Direct Implementation)

The **vibe** workflow is designed for rapid iteration and simpler tasks. It skips formal planning and goes straight to implementation with full tool access.

### Phase 1: Understand
- **Agent**: Vibe agent (primary)
- **Interaction**: Direct conversation with user
- **Output**: Clear understanding of what to build
- **Actions**:
  - Quickly understand requirements
  - Ask clarifying questions
  - Move to implementation immediately

### Phase 2: Implement and Test (Iterative)
- **Agent**: Vibe agent (with optional subagent delegation)
- **Interaction**: Direct implementation
- **Output**: Working implementation
- **Actions**:
  - Implement solution directly
  - Test as you go
  - Iterate based on results
  - Optionally delegate to subagents for complex sub-tasks

### When to Use Vibe Workflow

Use `agent vibe` when:
- Making quick fixes or small changes
- Prototyping and experimenting
- The task is simple and well-understood
- You need rapid iteration
- Working on proof-of-concepts
- Time is more important than perfect architecture

### Optional Subagent Delegation

Even in vibe mode, you can delegate to subagents if needed:
- `@planner` - If architecture becomes complex
- `@test-writer` - If comprehensive tests are needed
- `@implementer` - To parallelize work

---

## Choosing the Right Workflow

| Scenario | Recommended Workflow |
|----------|---------------------|
| Complex feature requiring architecture | `agent spec` |
| Multiple components to coordinate | `agent spec` |
| Production-ready implementation | `agent spec` |
| Work spans multiple phases | `agent spec` |
| Release boundaries needed | `agent spec` |
| Quick bug fix | `agent vibe` |
| Prototype/MVP | `agent vibe` |
| Simple, well-understood task | `agent vibe` |
| Experimentation | `agent vibe` |
| Unfamiliar domain | `agent spec` |

**Rule of thumb**: Start with `agent vibe` for speed. If you find yourself needing formal planning, switch to `agent spec`.

---

## Agent Roles and Boundaries

### Primary Agents

**Spec Agent**:
- Mode: Primary orchestrator
- Prompt: Phase-based execution workflow
- Permissions: Task delegation only (no direct editing)
- Creates high-level spec itself (does NOT delegate to planner)
- Identifies phases and release boundaries
- Gets explicit user approval
- Can delegate to: @planner, @reviewer, @test-writer, @implementer

**Vibe Agent**:
- Mode: Primary implementer
- Prompt: Direct implementation workflow
- Permissions: Full tool access (edit, bash, etc.)
- Can optionally delegate to: @planner, @reviewer, @test-writer, @implementer

### Subagents

**Planner**:
- Reads: Spec files via `spec_read`, codebase analysis (git, ls, find, grep)
- Writes: Updates phases via `spec_update` (fills in file_changes and test_cases)
- Cannot: Execute code, write tests, write implementation, create new specs
- Called by: Spec agent for low-level phase planning

**Reviewer**:
- Reads: Spec files via `spec_read`, codebase analysis
- Writes: Updates review field via `spec_update` (only agent that can set review status)
- Cannot: Write code, modify phases, create specs
- Called by: Spec agent for adversarial review

**Test-Writer**:
- Reads: Spec files via `spec_read`, existing code
- Writes: Test files only (matching language conventions in spec)
- Runs: Test commands to verify tests compile
- Cannot: Write implementation code, modify spec
- Called by: Spec agent for test writing

**Implementer**:
- Reads: Spec files via `spec_read`, existing code
- Writes: Implementation files only
- Runs: Tests to verify implementation
- Applies: Trivial fixes consistent with spec
- Cannot: Write test code, modify spec, ignore test failures
- Called by: Spec agent for implementation

---

## Spec File Format

Specs are stored as `.opencode/specs/<timestamp>-<slug>.yaml` and include:

```yaml
version: 2
problem_statement: "Clear description of the problem"
goals:
  - "Goal 1"
  - "Goal 2"
constraints:
  - "Constraint 1"
tech_choices:
  - area: "database"
    choice: "PostgreSQL"
    rationale: "Needed for ACID transactions"
architecture:
  description: "High-level architecture description"
  components:
    - name: "Parser"
      responsibility: "Parses CSV files"
      interfaces: ["parse(file) -> DataFrame"]
test_strategy:
  approach: "Unit and integration tests"
  test_patterns:
    - language: "go"
      pattern: "*_test.go"
phases:
  - name: "Phase 1: Parser Implementation"
    description: "Build the CSV parser"
    status: "pending"
    is_release_boundary: false
    file_changes: []  # Filled by planner
    test_cases: []    # Filled by planner
  - name: "Phase 2: API Layer"
    description: "Add HTTP endpoints"
    status: "pending"
    is_release_boundary: true  # STOP here - changes must be released before Phase 3
    file_changes: []
    test_cases: []
  - name: "Phase 3: Client Updates"
    description: "Update clients to use new API"
    status: "pending"
    is_release_boundary: false  # Can only proceed after Phase 2 is released
    file_changes: []
    test_cases: []
review:  # Added by reviewer
  status: "passed"
  reviewer: "reviewer"
  timestamp: "2024-01-15T10:30:00Z"
  feedback: "Optional feedback"
status: "draft"  # draft → planned → reviewed → approved → in_progress → completed
```

All fields are mandatory for validation.

---

## Custom Tools

**`spec_write`**: Writes and validates spec YAML files
- Input: title, structured spec object
- Validation: Ensures all mandatory fields present
- Output: File path + summary
- Used by: Spec agent (for high-level spec creation)

**`spec_read`**: Reads and validates spec files
- Input: optional path (reads most recent if not specified)
- Output: Formatted spec content with all fields
- Used by: Spec agent, planner, reviewer, test-writer, implementer

**`spec_update`**: Updates portions of a spec file
- Input: path, optional status, optional phase updates, optional review updates
- Supports: Overall status changes, phase status, phase file_changes, phase test_cases, review status/feedback
- Used by: Spec agent (status changes), planner (phase details), reviewer (review fields)

---

## Language-Aware Test Patterns

The spec must define test file patterns for each language:

| Language | Pattern | Convention |
|----------|---------|-----------|
| Go | `*_test.go` | Same dir as implementation |
| TypeScript | `*.test.ts` or `*.spec.ts` | Per-file or in tests/ dir |
| Python | `test_*.py` or `*_test.py` | Per-file or in tests/ dir |
| Rust | `#[cfg(test)]` or `tests/` | In src/lib.rs or separate |
| Java | `*Test.java` | Same package structure |

When a new language is introduced:
- Spec agent must research and define appropriate test patterns
- If uncertain, ask user to clarify conventions
- Include patterns in `test_strategy.test_patterns` list

---

## Configuration

Configuration is in `opencode.json`:
- Spec agent uses orchestrator prompt with subagent delegation
- Vibe agent uses direct implementation prompt with full permissions
- Subagents (planner, test-writer, implementer) are defined and configured
- Tool permissions control what each agent can do

### Working on opencode config

#### Prefer tools over skills

Tools provide more deterministic results and better control. Use tools (Bash, Read, Edit, Write, Glob, Grep, etc.) over skills when possible.

#### Restrict subagent capabilities

When launching subagents, limit their tool access to only the essential tools required for their specific task. Avoid giving broad access that isn't necessary for the job at hand.

---

## Using the Workflow

### Starting a Spec Session

```bash
agent spec
```

1. Engage with spec agent in Phase 1 (Understand & Create High-Level Spec)
2. Work through problem understanding together
3. Spec agent creates high-level spec with phases
4. Spec agent delegates to @planner to fill in all phase details
5. Spec agent delegates to @reviewer for adversarial review
6. If review fails, planner is re-invoked with feedback, then re-review
7. Review and explicitly approve the spec
8. Spec agent iterates through phases:
   - Delegates test writing to @test-writer
   - Delegates implementation to @implementer
9. If non-trivial issues arise, loop back to planning
10. At release boundaries, spec agent pauses for user confirmation of release

### Starting a Vibe Session

```bash
agent vibe
```

1. Quickly explain what you need
2. Agent starts implementing immediately
3. Agent tests as they go
4. Iterate until it works

---

## Wrapper Scripts

The `agent` script (and `agent.bat` on Windows) provides a simple CLI for selecting agents:

- `agent spec` - Launch spec agent (structured workflow)
- `agent vibe` - Launch vibe agent (direct implementation)

The wrapper scripts:
- Map subcommands to agent names
- Launch opencode with `--agent <agent-name>` flag
- Do NOT pass any prompt - you enter it in the TUI
- Handle invalid subcommands gracefully with helpful error messages

---

## Workflow Benefits

**Spec Workflow**:
- Problems are well-understood before building
- Architecture is approved before coding
- All phases are planned before any implementation
- Adversarial review catches issues early
- Tests drive implementation
- Non-trivial issues are escalated rather than worked around
- Phase-based execution with release boundary support
- High-level planning done by spec agent (not delegated)

**Vibe Workflow**:
- Rapid iteration and fast feedback
- Minimal overhead for simple tasks
- Full tool access for maximum flexibility
- Optional delegation when needed
