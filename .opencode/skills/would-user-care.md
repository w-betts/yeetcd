# Would-User-Care Decision Framework

This skill provides a shared decision framework for determining when ambiguity in the spec-tree workflow should be escalated to the user vs decided autonomously by the agent.

## Core Principle

**Conservative escalation: when in doubt, escalate.** Only purely mechanical/line-level details with no behavioral impact are agent-decidable. Everything else should be escalated to the user via the orchestrator (pause-ask-resume).

## The Ambiguity Spectrum

### Tier 1: Mechanical (Agent-Decidable)
Decisions with NO behavioral impact. Purely cosmetic or internal.

Examples:
- Local variable naming in method bodies
- Line wrapping, formatting, whitespace
- Which specific Java collection to use in a private method body (e.g., ArrayList vs LinkedList for a temp variable)
- Import organization
- Comments on obvious/internal code

### Tier 2: Gray Zone (Agent Decides, Flags Uncertainty)
Reasonable teams may disagree. The agent makes the decision but if uncertain about user preference, flags it.

Examples:
- Whether to extract a small private helper method
- Standard library function choice when multiple produce the same result
- Following established codebase patterns that aren't explicitly documented
- Test data values that don't affect test meaning

### Tier 3: User-Significant (Must Escalate)
Decisions the user would care about. Agent MUST STOP and escalate to orchestrator.

## The "Would-User-Care" Litmus Test

Run these 7 criteria before EVERY decision. If ANY is YES, ESCALATE.

| # | Criterion | What it catches |
|---|-----------|-----------------|
| 1 | **Does this change WHAT THE SYSTEM DOES?** | Scope, behavior, outcomes |
| 2 | **Does this change HOW THE SYSTEM IS USED?** | Public API, CLI flags, config format, interfaces |
| 3 | **Does this change HOW THE SYSTEM IS OPERATED?** | Deployment, env vars, monitoring, infrastructure |
| 4 | **Does this change WHAT IS TESTED or HOW?** | Test strategy, coverage, test types |
| 5 | **Does this change the ARCHITECTURE?** | New components, patterns, dependencies |
| 6 | **Does this change DEFAULT BEHAVIOR?** | Default values, fallbacks, error handling |
| 7 | **Am I deciding on ERROR HANDLING STRATEGY?** | What happens on failure |

### Agent-Decidable Test (check this LAST)
| # | Criterion | What it means |
|---|-----------|---------------|
| 8 | **Is this purely formatting / local variables / naming / internal implementation with NO behavioral impact?** | If ALL of these are true, decide autonomously |

## Escalation Protocol

When a user-significant decision is identified:

1. **STOP** — Do not make the decision autonomously
2. **REPORT** — Send details to the orchestrator:
   - What decision needs to be made
   - The options/alternatives
   - Why each option matters (which criteria triggered)
   - What direction is needed from the user
3. **WAIT** — The orchestrator will ask the user and get back to you
4. **RESUME** — Continue with the user's direction

## Concrete Examples

### Would escalate (User-Significant):
- "Should I use JPA or JDBC for this database access?" → Architecture change (#5)
- "Should this endpoint return 404 or 403 when not found?" → Error handling (#7)
- "Default timeout: 30s or 60s?" → Default behavior (#6)
- "Should we add integration tests or keep this as unit tests?" → Test strategy (#4)
- "Should the output be JSON or XML?" → How system is used (#2)

### Would NOT escalate (Agent-Decidable):
- "Should I name this temp variable `result` or `output`?" → Naming (#8)
- "Should I use `for` or `while` for this loop?" → Internal impl (#8)
- "Should I add a blank line before this return statement?" → Formatting (#8)
