# yeetcd

Agent-friendly cd

## OpenCode Structured Development Workflow

This project uses a 4-phase structured workflow for all feature development. The workflow is orchestrated by OpenCode's `build` and `plan` agents, which delegate specialized work to three custom subagents.

### The 4-Phase Workflow

Every feature development session follows this flow:

#### Phase 1: Understand
- **Agent**: Primary agent (Build or Plan)
- **Interaction**: Direct conversation with user
- **Output**: Clear, unambiguous problem statement
- **Actions**:
  - Play back understanding
  - Ask clarifying questions
  - Highlight ambiguities
  - Confirm understanding before moving on

#### Phase 2: Plan
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

#### Phase 3: Write Tests
- **Agent**: `@test-writer` subagent
- **Interaction**: Primary agent delegates to test-writer
- **Output**: Test files matching language conventions
- **Actions**:
  - Read plan via `plan_read`
  - Create test files using patterns from plan
  - Implement all test cases from plan
  - Verify tests compile/load
  - Report test infrastructure created

#### Phase 4: Implement
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

### Agent Roles and Boundaries

#### Primary Agents (Build & Plan)
- **Build**: Full permissions, acts as orchestrator
- **Plan**: Restricted permissions (edit/bash: ask), but same orchestrator role
- **Role**: Guide user through workflow, delegate to subagents, present decisions to user
- **Key Principle**: Do NOT write code directly; orchestrate the process

#### Subagents

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

### Plan File Format

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

### Custom Tools

**`plan_write`**: Writes and validates plan YAML files
- Input: title, structured plan object
- Validation: Ensures all mandatory fields present
- Output: File path + summary
- Used by: Planner

**`plan_read`**: Reads and validates plan files
- Input: optional path (reads most recent if not specified)
- Output: Formatted plan content with all fields
- Used by: Test-writer, Implementer

### Language-Aware Test Patterns

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

### Configuration

Configuration is in `opencode.json`:
- Build agent uses orchestrator prompt
- Plan agent uses same orchestrator prompt (with restricted direct permissions)
- All three subagents are defined and configured
- Tool permissions control what each subagent can do
- Both agents can invoke all three subagents via task permissions

### Working on opencode config

#### Prefer tools over skills

Tools provide more deterministic results and better control. Use tools (Bash, Read, Edit, Write, Glob, Grep, etc.) over skills when possible.

#### Restrict subagent capabilities

When launching subagents, limit their tool access to only the essential tools required for their specific task. Avoid giving broad access that isn't necessary for the job at hand.

### Using the Workflow

To start a feature development session:

1. In OpenCode, ask for a feature
2. Build/Plan agent will engage you in Phase 1 (Understand)
3. Agent will invoke planner for Phase 2
4. Review and approve the plan
5. Agent invokes test-writer for Phase 3
6. Agent invokes implementer for Phase 4
7. If non-trivial issues arise, loop back to planning

The workflow ensures:
- Problems are well-understood before building
- Architecture is approved before coding
- Tests drive implementation
- Non-trivial issues are escalated rather than worked around
