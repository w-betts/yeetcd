# yeetcd

Agent-friendly cd

## Project Context

This project provides a multi-workflow agent system with two primary modes:
- **spec**: Structured workflow for complex features (planning, review, phased implementation)
- **vibe**: Direct implementation for rapid iteration

Both modes have access to the same subagents and tools.

## Language-Aware Test Patterns

When writing tests, follow language conventions:

| Language | Pattern | Convention |
|----------|---------|-----------|
| Go | `*_test.go` | Same dir as implementation |
| TypeScript | `*.test.ts` or `*.spec.ts` | Per-file or in tests/ dir |
| Python | `test_*.py` or `*_test.py` | Per-file or in tests/ dir |
| Rust | `#[cfg(test)]` or `tests/` | In src/lib.rs or separate |
| Java | `*Test.java` | Same package structure |

## Configuration

Agent configuration is in `opencode.json`. Subagent prompts are in `.opencode/prompts/`.

### Working on opencode config

#### Prefer tools over skills

Tools provide more deterministic results and better control. Use tools (Bash, Read, Edit, Write, Glob, Grep, etc.) over skills when possible.

#### Restrict subagent capabilities

When launching subagents, limit their tool access to only the essential tools required for their specific task. Avoid giving broad access that isn't necessary for the job at hand.
