import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"
import crypto from "crypto"

// --- Schema ---

const ProblemSeveritySchema = z.enum(["critical", "high", "medium", "low"])
const ProblemTypeSchema = z.enum([
  "tool_failure",
  "misunderstanding",
  "workflow_friction",
  "assumption_wrong",
  "user_feedback_negative",
  "regression",
])

const ProblemContextSchema = z.record(z.string(), z.unknown())

const ProblemSchema = z.object({
  type: ProblemTypeSchema.describe("Category of the problem"),
  description: z.string().min(1).describe("What happened"),
  context: ProblemContextSchema.describe("Relevant details about the problem"),
  timestamp: z.string().describe("ISO timestamp when the problem occurred"),
  severity: ProblemSeveritySchema.describe("Impact level of the problem"),
  analysed: z.boolean().optional().describe("Whether this problem has been analyzed by improve agent"),
})

const SessionSchema = z.object({
  session_id: z.string().describe("Unique session identifier"),
  workflow_type: z.enum(["spec", "vibe", "fix", "document"]).describe("Type of workflow"),
  started_at: z.string().describe("ISO timestamp when session started"),
  ended_at: z.string().optional().describe("ISO timestamp when session ended"),
  branch: z.string().optional().describe("Git branch name if in worktree"),
  worktree: z.string().optional().describe("Worktree path if applicable"),
  problems: z.array(ProblemSchema).describe("Problems recorded during the session"),
  summary: z.string().optional().describe("Optional summary of the session"),
})

type Session = z.infer<typeof SessionSchema>
type Problem = z.infer<typeof ProblemSchema>

// --- Helpers ---

function formatValidationErrors(error: z.ZodError): string {
  return error.issues
    .map((issue) => {
      const path = issue.path.join(".")
      return `  - ${path ? path + ": " : ""}${issue.message}`
    })
    .join("\n")
}

function parseYaml(content: string): unknown {
  try {
    return yaml.parse(content)
  } catch {
    return null
  }
}

function toYamlString(obj: unknown): string {
  return yaml.stringify(obj, { lineWidth: 0, defaultStringType: "QUOTE_DOUBLE", defaultKeyType: "PLAIN" })
}

function sessionsDir(worktree: string): string {
  return path.join(worktree, ".opencode", "sessions")
}

function getWorkflowDir(worktree: string, workflowType: string): string {
  return path.join(sessionsDir(worktree), workflowType)
}

function generateSessionId(): string {
  const timestamp = Math.floor(Date.now() / 1000)
  const random = crypto.randomBytes(4).toString("hex")
  return `${timestamp}-${random}`
}

function getCurrentBranch(worktree: string): string | null {
  try {
    const dotGitPath = path.join(worktree, ".git")

    if (fs.existsSync(dotGitPath)) {
      const dotGitStat = fs.statSync(dotGitPath)

      if (dotGitStat.isFile()) {
        const gitdirContent = fs.readFileSync(dotGitPath, "utf-8").trim()
        const gitdirMatch = gitdirContent.match(/^gitdir: (.+)$/)
        if (gitdirMatch) {
          const gitdir = gitdirMatch[1]
          const headPath = path.join(gitdir, "HEAD")
          if (fs.existsSync(headPath)) {
            const headContent = fs.readFileSync(headPath, "utf-8").trim()
            const branchMatch = headContent.match(/^ref: refs\/heads\/(.+)$/)
            if (branchMatch) {
              return branchMatch[1]
            }
          }
        }
      } else if (dotGitStat.isDirectory()) {
        const headPath = path.join(dotGitPath, "HEAD")
        if (fs.existsSync(headPath)) {
          const headContent = fs.readFileSync(headPath, "utf-8").trim()
          const branchMatch = headContent.match(/^ref: refs\/heads\/(.+)$/)
          if (branchMatch) {
            return branchMatch[1]
          }
        }
      }
    }
  } catch {
    // Return null on error
  }
  return null
}

// --- Tools ---

export const session_start = tool({
  description:
    "Start a new session for tracking problems during an agent workflow. " +
    "Creates a new session file in .opencode/sessions/<workflow-type>/. " +
    "Returns the session ID which should be used for subsequent session operations.",
  args: {
    workflow_type: tool.schema
      .enum(["spec", "vibe", "fix", "document"])
      .describe("Type of workflow being run"),
  },
  async execute(args, context) {
    const sessionId = generateSessionId()
    const branch = getCurrentBranch(context.worktree)

    // Determine worktree path (relative to worktree root if possible)
    let worktreePath: string | null = null
    const worktreesDir = path.join(context.worktree, ".worktrees")
    if (fs.existsSync(worktreesDir)) {
      // Check if we're in a worktree by looking for .git file
      const gitPath = path.join(context.worktree, ".git")
      if (fs.existsSync(gitPath) && fs.statSync(gitPath).isFile()) {
        // We're in a worktree
        worktreePath = context.worktree
      }
    }

    const session: Session = {
      session_id: sessionId,
      workflow_type: args.workflow_type,
      started_at: new Date().toISOString(),
      branch: branch || undefined,
      worktree: worktreePath || undefined,
      problems: [],
    }

    // Create directory
    const workflowDir = getWorkflowDir(context.worktree, args.workflow_type)
    fs.mkdirSync(workflowDir, { recursive: true })

    // Write session file
    const filename = `${sessionId}.yaml`
    const filepath = path.join(workflowDir, filename)
    const yamlContent = toYamlString(session)
    fs.writeFileSync(filepath, yamlContent)

    return `Session started successfully.

Session ID: ${sessionId}
Workflow: ${args.workflow_type}
Branch: ${branch || "(main)"}
File: ${filepath}

Use this session_id (${sessionId}) for session_record_problem and session_end.`
  },
})

export const session_record_problem = tool({
  description:
    "Record a problem that occurred during the session. " +
    "Use this when something goes unexpectedly wrong - tool failures, misunderstandings, workflow friction, etc.",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID from session_start"),
    type: tool.schema
      .enum([
        "tool_failure",
        "misunderstanding",
        "workflow_friction",
        "assumption_wrong",
        "user_feedback_negative",
        "regression",
      ])
      .describe("Category of the problem"),
    description: tool.schema
      .string()
      .describe("What happened - brief description"),
    context: tool.schema
      .record(tool.schema.string(), tool.schema.unknown())
      .optional()
      .describe("Relevant context details (e.g., file, tool, expected vs actual)"),
    severity: tool.schema
      .enum(["critical", "high", "medium", "low"])
      .describe("Impact level"),
  },
  async execute(args, context) {
    // Find session file
    const sessionsBase = sessionsDir(context.worktree)

    // Search all workflow directories for this session
    let sessionFile: string | null = null
    const workflows = ["spec", "vibe", "fix", "document"]

    for (const wf of workflows) {
      const wfDir = path.join(sessionsBase, wf)
      if (fs.existsSync(wfDir)) {
        const candidate = path.join(wfDir, `${args.session_id}.yaml`)
        if (fs.existsSync(candidate)) {
          sessionFile = candidate
          break
        }
      }
    }

    if (!sessionFile) {
      return `ERROR: Session not found: ${args.session_id}

Use session_start first to create a session.`
    }

    // Read existing session
    const content = fs.readFileSync(sessionFile, "utf-8")
    let session: Session

    try {
      const parsed = parseYaml(content)
      if (!parsed) {
        return `ERROR: Failed to parse session file: ${sessionFile}`
      }
      session = parsed as Session
    } catch {
      return `ERROR: Failed to parse session file: ${sessionFile}`
    }
    // Validate
    const parseResult = SessionSchema.safeParse(session)
    if (!parseResult.success) {
      return `ERROR: Session file has validation issues:\n${formatValidationErrors(parseResult.error)}`
    }

    // Add problem
    const problem: Problem = {
      type: args.type,
      description: args.description,
      context: args.context || {},
      timestamp: new Date().toISOString(),
      severity: args.severity,
    }

    session.problems.push(problem)

    // Write back
    const yamlContent = toYamlString(session)
    fs.writeFileSync(sessionFile, yamlContent)

    return `Problem recorded successfully.

Session: ${args.session_id}
Problem: ${args.type} - ${args.description}
Severity: ${args.severity}

Total problems in session: ${session.problems.length}`
  },
})

export const session_end = tool({
  description:
    "End a session and add an optional summary. " +
    "Call this when the agent workflow completes (before committing).",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID from session_start"),
    summary: tool.schema
      .string()
      .optional()
      .describe("Optional summary of the session"),
  },
  async execute(args, context) {
    // Find session file
    const sessionsBase = sessionsDir(context.worktree)

    // Search all workflow directories for this session
    let sessionFile: string | null = null
    let workflowType: string | null = null
    const workflows = ["spec", "vibe", "fix", "document"]

    for (const wf of workflows) {
      const wfDir = path.join(sessionsBase, wf)
      if (fs.existsSync(wfDir)) {
        const candidate = path.join(wfDir, `${args.session_id}.yaml`)
        if (fs.existsSync(candidate)) {
          sessionFile = candidate
          workflowType = wf
          break
        }
      }
    }

    if (!sessionFile) {
      return `ERROR: Session not found: ${args.session_id}`
    }

    // Read existing session
    const content = fs.readFileSync(sessionFile, "utf-8")
    let session: Session

    try {
      const parsed = parseYaml(content)
      if (!parsed) {
        return `ERROR: Failed to parse session file: ${sessionFile}`
      }
      session = parsed as Session
    } catch {
      return `ERROR: Failed to parse session file: ${sessionFile}`
    }

    // Validate
    const parseResult = SessionSchema.safeParse(session)
    if (!parseResult.success) {
      return `ERROR: Session file has validation issues:\n${formatValidationErrors(parseResult.error)}`
    }

    // Update session
    session.ended_at = new Date().toISOString()
    if (args.summary) {
      session.summary = args.summary
    }

    // Write back
    const yamlContent = toYamlString(session)
    fs.writeFileSync(sessionFile, yamlContent)

    const problemSummary = {
      critical: session.problems.filter((p) => p.severity === "critical").length,
      high: session.problems.filter((p) => p.severity === "high").length,
      medium: session.problems.filter((p) => p.severity === "medium").length,
      low: session.problems.filter((p) => p.severity === "low").length,
    }

    return `Session ended successfully.

Session ID: ${args.session_id}
Workflow: ${workflowType}
Started: ${session.started_at}
Ended: ${session.ended_at}
Problems recorded: ${session.problems.length}
  - Critical: ${problemSummary.critical}
  - High: ${problemSummary.high}
  - Medium: ${problemSummary.medium}
  - Low: ${problemSummary.low}
${session.summary ? `\nSummary: ${session.summary}` : ""}

File: ${sessionFile}`
  },
})

export const session_mark_analysed = tool({
  description:
    "Mark problems in a session as analysed by the improve agent. " +
    "This prevents the same problems from being flagged repeatedly in future analysis.",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID to mark as analysed"),
    problem_indices: tool.schema
      .array(tool.schema.number())
      .describe("Indices of problems to mark as analysed (0-based)"),
  },
  async execute(args, context) {
    // Find session file
    const sessionsBase = sessionsDir(context.worktree)

    let sessionFile: string | null = null
    const workflows = ["spec", "vibe", "fix", "document"]

    for (const wf of workflows) {
      const wfDir = path.join(sessionsBase, wf)
      if (fs.existsSync(wfDir)) {
        const candidate = path.join(wfDir, `${args.session_id}.yaml`)
        if (fs.existsSync(candidate)) {
          sessionFile = candidate
          break
        }
      }
    }

    if (!sessionFile) {
      return `ERROR: Session not found: ${args.session_id}`
    }

    // Read existing session
    const content = fs.readFileSync(sessionFile, "utf-8")
    let session: Session

    try {
      const parsed = parseYaml(content)
      if (!parsed) {
        return `ERROR: Failed to parse session file: ${sessionFile}`
      }
      session = parsed as Session
    } catch {
      return `ERROR: Failed to parse session file: ${sessionFile}`
    }

    // Mark problems as analysed
    const updated: number[] = []
    const invalid: number[] = []

    for (const idx of args.problem_indices) {
      if (idx >= 0 && idx < session.problems.length) {
        session.problems[idx].analysed = true
        updated.push(idx)
      } else {
        invalid.push(idx)
      }
    }

    if (updated.length === 0) {
      return `ERROR: No valid problem indices. Valid range: 0-${session.problems.length - 1}`
    }

    // Write back
    const yamlContent = toYamlString(session)
    fs.writeFileSync(sessionFile, yamlContent)

    return `Marked ${updated.length} problem(s) as analysed.

Session: ${args.session_id}
Updated indices: ${updated.join(", ")}
${invalid.length > 0 ? `Invalid indices (skipped): ${invalid.join(", ")}` : ""}`
  },
})
