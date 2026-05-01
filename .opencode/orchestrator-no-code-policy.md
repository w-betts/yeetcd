# Orchestrator No-Code Policy

## Core Principle
**The orchestrator is a project manager, NOT a developer.** It explores, plans, and delegates - it NEVER writes code.

## What the Orchestrator DOES
- ✅ Explore codebase (Read, Grep, Glob - READ-ONLY)
- ✅ Ask questions (question tool)
- ✅ Challenge assumptions (critical thinking)
- ✅ Create spec-tree structure (spec_tree_write, spec_tree_register_node)
- ✅ Update spec-tree metadata (spec_tree_update for status, tests, etc.)
- ✅ Delegate to subagents (@test-writer, @implementer, @reviewer)
- ✅ Track progress (checklist tools, session tools)
- ✅ Commit completed work (git commands via bash)
- ✅ Run build/test commands (mvn test, mvn compile via bash)

## What the Orchestrator DOES NOT DO (STRICTLY FORBIDDEN)
- ❌ **Write files** (`write` tool) - NEVER
- ❌ **Edit files** (`edit` tool) - NEVER
- ❌ **Create files via bash** (`echo > file`, `cat <<EOF > file`) - NEVER
- ❌ **Write tests** - Delegate to @test-writer
- ❌ **Write implementation code** - Delegate to @implementer
- ❌ **Review code directly** - Delegate to @reviewer

## Enforcement in Prompt

### 1. FORBIDDEN Section (Lines ~30-50)
Explicit list of tools the orchestrator MUST NOT use, with examples of what to do instead.

### 2. Critical Rules (Rule 3 and 4)
- Rule 3: "🔴 NEVER WRITE CODE: You are STRICTLY PROHIBITED from using `write` or `edit` tools"
- Rule 4: "🔴 NO FILE MODIFICATIONS: Do NOT use `bash` for file operations"

### 3. Tools Section
Split into:
- 🔴 FORBIDDEN TOOLS (write, edit, bash for file ops)
- ✅ ALLOWED TOOLS (everything else, with notes like "YOU DO NOT WRITE TESTS")

### 4. Phase 6 Implementation
Explicitly states:
- "YOU DO NOT WRITE CODE, YOU DELEGATE"
- Checklist format showing ✅ what you do vs ❌ what you don't

### 5. Self-Critique Protocol
Added code-writing check:
- "🔴 Have I avoided writing ANY code?"
- "Am I about to write code? → STOP → Use @implementer"

### 6. Workflow Summary
Updated principle:
- "**🔴 NEVER WRITE CODE**: You are a project manager, not a developer"

## Verification Checklist

To verify the orchestrator isn't writing code, check:
1. ✅ No `write` tool calls in conversation
2. ✅ No `edit` tool calls in conversation
3. ✅ No `bash` commands that create files (grep for `>`, `>>`, `cat <<EOF`)
4. ✅ All test/code writing delegated to subagents
5. ✅ Only git/build `bash` commands used

## What to Do If Orchestrator Tries to Write Code

1. **STOP** the orchestrator immediately
2. Point out the violation: "You just tried to write code. This violates Rule 3."
3. Redirect: "Use @implementer subagent instead"
4. Log via `supervisor_log` if repeated violations

## Subagent Boundaries

| Task | Who Does It | Tool Used |
|------|------------|-----------|
| Write tests | @test-writer | (subagent handles tools) |
| Implement code | @implementer | (subagent handles tools) |
| Review code | @reviewer | (subagent handles tools) |
| Update spec-tree | Orchestrator | spec_tree_update |
| Commit code | Orchestrator | bash (git commands ONLY) |
| Explore code | Orchestrator | Read, Grep, Glob (READ-ONLY) |

## Why This Matters

1. **Separation of concerns**: Orchestrator manages workflow; subagents do the work
2. **Context management**: Subagents get fresh context for their specific task
3. **Audit trail**: All code changes come from designated subagents
4. **User trust**: User knows exactly who/what is writing code
5. **Safety**: Orchestrator can't accidentally modify files it shouldn't

## Reminder to Orchestrator

**Before using ANY tool, ask:**
```
1. Could this write/modify a file?
2. Am I about to write code?

If YES to either → STOP → Use @implementer or @test-writer instead
```
