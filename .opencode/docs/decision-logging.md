# Decision Logging

## Overview

Agent flows log decisions they make that aren't specified in their spec or by an explicit user prompt. These are logged to `.opencode/decision-logs/<session-id>.yaml` so they survive session restarts.

## Scope (STRICT)

Only log **explicit choices between alternatives** that:
- Are NOT in the spec
- Are NOT explicit user prompts

### ✅ DO Log:
- "I chose approach A over B because..." (your judgment call)
- "Suggested breakdown: X, Y, Z" (your analysis, not in spec)

### ❌ DON'T Log:
- Following spec instructions ("add tests" - explicit in spec)
- Following user prompts ("implement X" - explicit user instruction)
- Trivial choices with no alternatives

## Usage

```typescript
decision_log({
  session_id: "<session_id>",
  agent_type: "<spec|vibe|fix|spec-tree|implementer|reviewer|test-writer>",
  decision: "Chose X over Y for this task",
  alternatives_considered: ["Y", "Z"],
  rationale: "X is simpler and meets all requirements"
})
```

## File Format

Decisions are stored as an array in `.opencode/decision-logs/<session-id>.yaml`:

```yaml
session_id: "12345"
decisions:
  - timestamp: "2026-05-01T12:00:00.000Z"
    agent_type: "spec"
    decision: "Chose approach A"
    alternatives_considered: ["B", "C"]
    rationale: "A is simpler"
  - timestamp: "2026-05-01T12:05:00.000Z"
    agent_type: "spec"
    decision: "Suggested breakdown: X, Y, Z"
    alternatives_considered: ["Mark as leaf", "Break down differently"]
    rationale: "X, Y, Z is most natural decomposition"
```

## Tools

- `decision_log`: Write/append a decision to the session's log file
- `decision_read`: Read and format a session's decision log
