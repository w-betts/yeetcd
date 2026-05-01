import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

// --- Schema ---

const DecisionSchema = z.object({
  timestamp: z.string().describe("ISO timestamp when the decision was made"),
  agent_type: z.string().describe("Type of agent that made the decision (e.g., spec, vibe, fix, implementer)"),
  decision: z.string().min(1).describe("The decision that was made"),
  alternatives_considered: z.array(z.string()).optional().describe("Alternative options that were considered before making the decision"),
  rationale: z.string().optional().describe("Why this decision was made"),
})

const DecisionLogSchema = z.object({
  version: z.literal(1).describe("Decision log schema version"),
  session_id: z.string().describe("Associated session ID"),
  decisions: z.array(DecisionSchema).describe("Array of decisions made during the session"),
  created_at: z.string().describe("ISO timestamp when decision log was created"),
})

type DecisionLog = z.infer<typeof DecisionLogSchema>
type Decision = z.infer<typeof DecisionSchema>

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

function getDecisionLogPath(worktree: string, sessionId: string): string {
  return path.join(worktree, ".opencode", "decision-logs", `${sessionId}.yaml`)
}

function loadOrCreateDecisionLog(worktree: string, sessionId: string): DecisionLog {
  const decisionLogPath = getDecisionLogPath(worktree, sessionId)
  
  if (fs.existsSync(decisionLogPath)) {
    const content = fs.readFileSync(decisionLogPath, "utf-8")
    const parsed = parseYaml(content)
    if (parsed) {
      const result = DecisionLogSchema.safeParse(parsed)
      if (result.success) {
        return result.data
      }
    }
  }
  
  return {
    version: 1,
    session_id: sessionId,
    decisions: [],
    created_at: new Date().toISOString(),
  }
}

function saveDecisionLog(worktree: string, decisionLog: DecisionLog): void {
  const decisionLogPath = getDecisionLogPath(worktree, decisionLog.session_id)
  
  // Ensure directory exists
  const dir = path.dirname(decisionLogPath)
  fs.mkdirSync(dir, { recursive: true })
  
  const yamlContent = toYamlString(decisionLog)
  fs.writeFileSync(decisionLogPath, yamlContent)
}

function formatDecisionLog(decisionLog: DecisionLog): string {
  let output = `DECISION LOG (Session: ${decisionLog.session_id})\n`
  output += `Created: ${decisionLog.created_at}\n`
  output += `Total Decisions: ${decisionLog.decisions.length}\n\n`
  
  if (decisionLog.decisions.length === 0) {
    output += `No decisions recorded yet.\n`
  } else {
    for (let i = 0; i < decisionLog.decisions.length; i++) {
      const decision = decisionLog.decisions[i]
      output += `[${i + 1}] ${decision.decision}\n`
      output += `    Agent: ${decision.agent_type}\n`
      output += `    Time: ${decision.timestamp}\n`
      if (decision.alternatives_considered && decision.alternatives_considered.length > 0) {
        output += `    Alternatives: ${decision.alternatives_considered.join(", ")}\n`
      }
      if (decision.rationale) {
        output += `    Rationale: ${decision.rationale}\n`
      }
      output += `\n`
    }
  }
  
  return output
}

// --- Tools ---

export const decision_log = tool({
  description:
    "Log a decision made during an agent workflow. " +
    "Records the decision, alternatives considered, and rationale to a YAML file " +
    "stored at .opencode/decision-logs/<session-id>.yaml. " +
    "Appends to existing decision log if one exists for the session.",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID from session_start"),
    agent_type: tool.schema
      .string()
      .describe("Type of agent making the decision (e.g., spec, vibe, fix, implementer, planner)"),
    decision: tool.schema
      .string()
      .min(1)
      .describe("The decision that was made"),
    alternatives_considered: tool.schema
      .array(tool.schema.string())
      .optional()
      .describe("Alternative options that were considered before making the decision"),
    rationale: tool.schema
      .string()
      .optional()
      .describe("Why this decision was made"),
  },
  async execute(args, context) {
    const decisionLog = loadOrCreateDecisionLog(context.worktree, args.session_id)
    
    const decision: Decision = {
      timestamp: new Date().toISOString(),
      agent_type: args.agent_type,
      decision: args.decision,
    }
    
    if (args.alternatives_considered) {
      decision.alternatives_considered = args.alternatives_considered
    }
    
    if (args.rationale) {
      decision.rationale = args.rationale
    }
    
    decisionLog.decisions.push(decision)
    saveDecisionLog(context.worktree, decisionLog)
    
    return `Decision logged successfully.

Session: ${args.session_id}
Agent: ${args.agent_type}
Decision: ${args.decision}
${args.alternatives_considered ? `Alternatives: ${args.alternatives_considered.join(", ")}` : ""}
${args.rationale ? `Rationale: ${args.rationale}` : ""}

Total decisions in log: ${decisionLog.decisions.length}

File: ${getDecisionLogPath(context.worktree, args.session_id)}`
  },
})

export const decision_read = tool({
  description:
    "Read the decision log for a session. " +
    "Returns all decisions recorded for the given session ID, " +
    "or a message if no decision log exists.",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID to read decision log for"),
  },
  async execute(args, context) {
    const decisionLogPath = getDecisionLogPath(context.worktree, args.session_id)
    
    if (!fs.existsSync(decisionLogPath)) {
      return `No decision log found for session: ${args.session_id}

Expected path: ${decisionLogPath}

Use decision_log to start recording decisions for this session.`
    }
    
    const content = fs.readFileSync(decisionLogPath, "utf-8")
    const parsed = parseYaml(content)
    
    if (!parsed) {
      return `ERROR: Failed to parse decision log file: ${decisionLogPath}`
    }
    
    const result = DecisionLogSchema.safeParse(parsed)
    if (!result.success) {
      return `ERROR: Decision log file has validation issues:\n${formatValidationErrors(result.error)}`
    }
    
    return formatDecisionLog(result.data)
  },
})
