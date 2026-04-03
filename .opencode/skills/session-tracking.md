# Session Tracking Skill

This skill provides session tracking and problem recording capabilities for opencode agent sessions.

## Overview

The session tracking skill enables agents to:
1. Track session metadata (agent type, start/end time, branch/worktree info)
2. Record problems that occur during the session
3. Write structured session reports at session end

## Session Data

Sessions are tracked in YAML files organized by workflow type:
- Location: `.opencode/sessions/<workflow-type>/`
- Naming: `<timestamp>-<session-id>.yaml`
- Each agent invocation creates a new session file

## Session Data Structure

```yaml
session_id: "unique-session-id"
workflow_type: "spec"  # spec, vibe, fix, document
started_at: "2024-01-15T10:30:00Z"
ended_at: "2024-01-15T12:45:00Z"
branch: "work/spec-my-feature"
worktree: ".worktrees/work-spec-my-feature"
problems:
  - type: "tool_failure"
    description: "Edit tool failed to apply changes"
    context:
      file: "src/App.java"
      tool: "edit"
    timestamp: "2024-01-15T11:00:00Z"
    severity: "high"
  - type: "misunderstanding"
    description: "User wanted X but I implemented Y"
    context:
      expected: "Delete file"
      actual: "Clear file contents"
    timestamp: "2024-01-15T11:30:00Z"
    severity: "medium"
  - type: "workflow_friction"
    description: "User had to repeat information multiple times"
    context:
      repeated_info: "project structure"
    timestamp: "2024-01-15T12:00:00Z"
    severity: "low"
summary: "Optional summary of the session"
```

## Problem Types

| Type | Description |
|------|-------------|
| `tool_failure` | A tool (edit, bash, read, etc.) failed unexpectedly |
| `misunderstanding` | Agent misunderstood user intent |
| `workflow_friction` | The workflow felt awkward or inefficient |
| `assumption_wrong` | An assumption the agent made proved incorrect |
| `user_feedback_negative` | User explicitly indicated something was wrong |
| `regression` | Something that previously worked now doesn't |

## Problem Severity

| Level | When to use |
|-------|-------------|
| `critical` | Session had to be aborted or restarted |
| `high` | Significant impact on productivity |
| `medium` | Noticeable impact but work continued |
| `low` | Minor inconvenience |

## Problem Schema

Each problem has the following structure:

```yaml
- type: "tool_failure"
  description: "Edit tool failed to apply changes"
  context:
    file: "src/App.java"
    tool: "edit"
  timestamp: "2024-01-15T11:00:00Z"
  severity: "high"
  analysed: false  # Whether improve agent has processed this
```

## Usage (Tools)

This skill provides TypeScript tools that enforce schema validation:

### session_start

Starts a new session - call this at the beginning of your workflow:

```typescript
session_start(workflow_type: "vibe" | "spec" | "fix" | "document")
```

Returns a session_id - use this for subsequent calls.

### session_record_problem

Records a problem when something goes wrong:

```typescript
session_record_problem(
  session_id: string,  // from session_start
  type: "tool_failure" | "misunderstanding" | "workflow_friction" | "assumption_wrong" | "user_feedback_negative" | "regression",
  description: string,
  context?: Record<string, unknown>,  // optional additional details
  severity: "critical" | "high" | "medium" | "low"
)
```

### session_end

Ends the session - call this before committing:

```typescript
session_end(
  session_id: string,
  summary?: string  // optional session summary
)
```

### session_mark_analysed

Marks problems as analysed (used by improve agent):

```typescript
session_mark_analysed(
  session_id: string,
  problem_indices: number[]  // 0-based indices of problems
)
```

### session_archive

Archives a session by moving it to `.opencode/sessions/archived/<workflow>/`.
Archived sessions are not scanned by the improve agent:

```typescript
session_archive(session_id: string)
```

## Session Data Location

Sessions are stored in:
- `.opencode/sessions/spec/<session-id>.yaml`
- `.opencode/sessions/vibe/<session-id>.yaml`
- `.opencode/sessions/fix/<session-id>.yaml`
- `.opencode/sessions/document/<session-id>.yaml`

Archived sessions go to:
- `.opencode/sessions/archived/spec/<session-id>.yaml`
- `.opencode/sessions/archived/vibe/<session-id>.yaml`
- etc.
