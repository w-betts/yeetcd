# yeetcd

Agent-friendly cd

## Multi-Workflow Agent System

This project provides two distinct development workflows through the `agent` wrapper script:

1. **Spec Workflow** (`agent spec`) - Structured 4-phase workflow for complex features
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

## The Spec Workflow (4-Phase Structured)

The **spec** workflow is designed for complex features that require careful planning, architecture decisions, and comprehensive testing. It follows a structured 4-phase approach:

### Phase 1: Understand
- **Agent**: Spec agent (primary orchestrator)
- **Interaction**: Direct conversation with user
- **Output**: Clear, unambiguous problem statement
- **Actions**:
  - Play back understanding
  - Ask clarifying questions
  - Highlight ambiguities
  - Confirm understanding before moving on

### Phase 2: Plan
- **Agent**: `@planner` subagent
- **Interaction**: Primary agent delegates to planner
- **Output**: Structured YAML plan file with:
  - Problem statement and goals
  - Technology choices with rationale
  - System architecture and components
  - Test strategy with language-specific patterns
  - Detailed file changes list
- **Actions**:
  - Analyze existing codebase
  - Propose architecture
  - Define testing approach
  - Generate plan via `plan_write` tool
  - Primary agent presents plan to user for approval
  - Loop back to planner if user requests changes
  - Move to Phase 3 once user approves

### Phase 3: Write Tests
- **Agent**: `@test-writer` subagent
- **Interaction**: Primary agent delegates to test-writer
- **Output**: Test files matching language conventions
- **Actions**:
  - Read plan via `plan_read`
  - Create test files using patterns from plan
  - Implement all test cases from plan
  - Verify tests compile/load
  - Report test infrastructure created

### Phase 4: Implement
- **Agent**: `@implementer` subagent
- **Interaction**: Primary agent delegates to implementer
- **Output**: Implementation code passing all tests
- **Actions**:
  - Read plan via `plan_read`
  - Write implementation files
  - Run tests incrementally
  - Apply trivial fixes (formatting, simple bugs)
  - If non-trivial issue found:
    - Report clearly
    - Loop back to Phase 2 via primary agent
    - Continue after plan update
  - Report final status with test results

### When to Use Spec Workflow

Use `agent spec` when:
- Building complex features requiring architecture decisions
- Multiple components need coordination
- You need comprehensive test coverage
- The problem requires careful analysis before implementation
- Working with unfamiliar domains or technologies
- Building production-ready features

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
- Prompt: 4-phase structured workflow
- Permissions: Task delegation only (no direct editing)
- Can delegate to: @planner, @test-writer, @implementer

**Vibe Agent**:
- Mode: Primary implementer
- Prompt: Direct implementation workflow
- Permissions: Full tool access (edit, bash, etc.)
- Can optionally delegate to: @planner, @test-writer, @implementer

### Subagents

**Planner**:
- Reads: Codebase analysis (git, ls, find, grep)
- Writes: Plan YAML files only
- Calls: `plan_write` tool to save plans
- Cannot: Execute code, write tests, write implementation

**Test-Writer**:
- Reads: Plan files via `plan_read`, existing code
- Writes: Test files only (matching language patterns in plan)
- Runs: Test commands to verify tests compile
- Cannot: Write implementation code, modify plan

**Implementer**:
- Reads: Plan files via `plan_read`, existing code
- Writes: Implementation files only
- Runs: Tests to verify implementation
- Applies: Trivial fixes consistent with plan
- Cannot: Write test code, modify plan, ignore test failures

---

## Plan File Format

Plans are stored as `.opencode/plans/<timestamp>-<slug>.yaml` and include:

```yaml
version: 1
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
  test_cases:
    - description: "Parser handles valid CSV"
      type: "unit"
      target_component: "Parser"
file_changes:
  - path: "src/parser.go"
    action: "create"
    description: "CSV parser implementation"
    is_test: false
  - path: "src/parser_test.go"
    action: "create"
    description: "Parser tests"
    is_test: true
status: "draft"  # Changes to "approved" by user
```

All fields are mandatory for validation.

---

## Custom Tools

**`plan_write`**: Writes and validates plan YAML files
- Input: title, structured plan object
- Validation: Ensures all mandatory fields present
- Output: File path + summary
- Used by: Planner

**`plan_read`**: Reads and validates plan files
- Input: optional path (reads most recent if not specified)
- Output: Formatted plan content with all fields
- Used by: Test-writer, Implementer

---

## Language-Aware Test Patterns

The planner must define test file patterns for each language:

| Language | Pattern | Convention |
|----------|---------|-----------|
| Go | `*_test.go` | Same dir as implementation |
| TypeScript | `*.test.ts` or `*.spec.ts` | Per-file or in tests/ dir |
| Python | `test_*.py` or `*_test.py` | Per-file or in tests/ dir |
| Rust | `#[cfg(test)]` or `tests/` | In src/lib.rs or separate |
| Java | `*Test.java` | Same package structure |

When a new language is introduced:
- Planner must research and define appropriate test patterns
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

1. Engage with spec agent in Phase 1 (Understand)
2. Summarize back what you heard
3. Ask clarifying questions
4. Once clear, agent will invoke the planner to draft a solution
5. Review and approve the plan
6. Agent invokes test-writer for Phase 3
7. Agent invokes implementer for Phase 4
8. If non-trivial issues arise, loop back to planning

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
- Tests drive implementation
- Non-trivial issues are escalated rather than worked around

**Vibe Workflow**:
- Rapid iteration and fast feedback
- Minimal overhead for simple tasks
- Full tool access for maximum flexibility
- Optional delegation when needed
