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

## Usage

### Starting a Session

At the start of a session, agents should:
1. Generate a unique session ID (timestamp + random suffix)
2. Create the session file with metadata
3. Store the session ID for later use

### Recording Problems

When something goes unexpectedly wrong:
1. Capture the problem details (type, description, context, severity)
2. Append the problem to the session file
3. Include relevant context (file, tool, timestamps)

### Ending a Session

At session end (before commit):
1. Update the session file with end time
2. Add an optional summary
3. Ensure the file is saved

## Tools

The skill uses the standard `write` tool to create session files. Use it with paths like:
- `.opencode/sessions/vibe/20240115-103000-abc123.yaml`
- `.opencode/sessions/spec/20240115-104500-def456.yaml`

### write

Writes session data to YAML files:
```python
write(
    content: str,  # YAML content
    filePath: str  # Path like ".opencode/sessions/spec/20240115-103000-abc123.yaml"
)
```

### read

Read existing session files:
```python
read(
    filePath: str  # Path to the session file
)
```

## Integration

This skill should be loaded into all primary agent types:
- `spec` agent
- `vibe` agent
- `fix` agent
- `document` agent

The skill is automatically available when loaded via the skill tool.
