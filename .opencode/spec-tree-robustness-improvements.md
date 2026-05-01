# Spec-Tree Agent Robustness Improvements

## Problem Statement
The spec-tree agent was at risk of **skipping phases** (especially Phase 5.5 Pre-Implementation Review) and not following the workflow exactly.

## Solution: Checklist Gates with Hard Enforcement

### Design Decisions

1. **Strict Sequential Phases**: Phases must complete 100% before the next phase can start (no pipelining)
2. **Phase-Level Gates**: One checklist item per major phase (not per-node, to avoid checklist fatigue)
3. **Hard Enforcement**: Agent MUST check `checklist_status` before each phase and cannot proceed with pending items

### Implementation

#### 1. Session Initialization (MANDATORY)
At the start of EVERY conversation, the agent MUST:
- Call `session_session_start({ workflow_type: "spec" })` to get `session_id`
- Initialize 6 checklist items (one per phase gate)

#### 2. Phase Gates Defined

| Phase | Checklist Item Description | When Completed |
|-------|---------------------------|----------------|
| 1 | `phase1-complete: Problem understood, challenged, ambiguities resolved` | After Phase 1 done |
| 2 | `phase2-complete: Spec-tree created with root node` | After Phase 2 done |
| 3-4 | `phase3-4-complete: All nodes explored and decomposed (breadth-first)` | After breadth-first complete |
| 5 | `phase5-complete: All leaves defined with tests and implementation details` | After all leaves defined |
| 5.5 | `phase5.5-complete: All leaves reviewed and approved by user` | After review complete |
| 6 | `phase6-complete: All leaves implemented, tested, and committed` | After all impl complete |

#### 3. Enforcement Logic (Embedded in Each Phase)

**Before entering ANY phase**:
```javascript
checklist_checklist_status({
  session_id: "<session_id>",
  show_resolved: false
})
```

If ANY prerequisite items are pending:
- **STOP immediately**
- **DO NOT proceed to the next phase**
- Tell user about incomplete prerequisites
- Wait for `checklist_complete` to resolve items

**After completing ANY phase**:
1. Call `checklist_status` to get `item_id` for the phase
2. Call `checklist_complete` with resolution note
3. Only THEN proceed to next phase

#### 4. Changes to Prompt Structure

**Added Sections**:
1. `## MANDATORY: Session Initialization` - Tells agent to start session + init checklist
2. `## MANDATORY: Phase Gate Enforcement` - Explains the enforcement mechanism
3. Updated `## Core Workflow` - Added gate checks and completion markers to EVERY phase
4. Updated `## Workflow Summary` - Visual diagram now shows checklist gates
5. Updated `## Critical Rules` - Added 4 new rules (15-18) about checklist enforcement
6. Updated `## Phase 7` - Added session cleanup (session_end + session_archive)

**Key Additions to Each Phase**:
- **Gate Check**: Code block showing `checklist_status` call
- **Stop Condition**: "If X is pending, STOP and resolve it first"
- **Completion Gate**: Code block showing `checklist_complete` call with resolution note

#### 5. Visual Workflow Diagram

The Workflow Summary now includes ASCII art showing:
- `═══ CHECKLIST GATE ═══` before each phase
- `✅ CHECKLIST GATE: Verify phaseX-complete` call
- `**Phase X Complete** → checklist_checklist_complete(...)` markers

This makes it visually impossible to miss the gates.

### Critical Improvements

1. **Cannot Skip Phase 5.5**: The gate between Phase 5 and 5.5 is now structurally enforced
2. **Item ID Discovery**: Agent is taught to use `checklist_status` to find `item_id` (since `checklist_tick` doesn't return it)
3. **Session Tracking**: All phase completions are tracked via session + checklist tools
4. **Hard Stop**: Agent is explicitly told to **STOP** and **DO NOT proceed** if gates fail

### Edge Cases Handled

1. **Agent doesn't know item_id**: Taught to call `checklist_status` first to discover mappings
2. **Multiple pending items**: Agent lists ALL pending items, not just the first one
3. **User tries to skip**: Agent explicitly refuses and shows pending items
4. **Phase 3-4 complexity**: Breadth-first completion gate handles the recursive nature

### Testing the Robustness

To verify the agent follows the workflow:
1. Start a new spec-tree conversation
2. Verify agent calls `session_session_start` first
3. Verify agent initializes 6 checklist items
4. Try to trick agent into Phase 2 without completing Phase 1
5. Verify agent stops and shows pending items
6. Complete Phase 1, verify agent marks it complete
7. Continue testing phase transitions...

### Files Modified
- `.opencode/prompts/spec-tree-orchestrator.md` - Main prompt with all enforcement logic

### Future Improvements (Not Implemented)
- Add automatic `checklist_status` validation before EVERY tool call (might be overkill)
