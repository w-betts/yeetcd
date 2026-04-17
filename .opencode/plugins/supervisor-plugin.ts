/**
 * Supervisor Plugin
 *
 * Provides supervisor_log tool for agents to log decisions and difficulties.
 * Invokes supervisor LLM to analyze the log and detect misalignment/lack of progress.
 *
 * Also hooks into:
 * - tool.execute.after for question tool (captures user answers)
 * - chat.message (captures other user inputs)
 */

import type { Plugin } from "@opencode-ai/plugin"
import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

// --- Types ---

type LogEntry = {
  timestamp: string
  type: "user_input" | "decision" | "difficulty" | "supervisor_analysis"
  description: string
  context?: Record<string, unknown>
  // For supervisor analysis results
  status?: "proceed" | "intervene"
  question?: string
  analysis?: string
}

type DecisionLog = {
  session_id: string
  entries: LogEntry[]
}

// --- Helpers ---
function formatTimestamp(): string {
  return new Date().toISOString()
}

function debugLog(worktree: string, ...args: unknown[]): void {
  const logPath = path.join(worktree, ".opencode", "supervisor-debug.log")
  const msg = args.map(a => typeof a === 'object' ? JSON.stringify(a, null, 2) : String(a)).join(' ')
  const line = `[${formatTimestamp()}] ${msg}\n`
  fs.appendFileSync(logPath, line)
}

function decisionLogsDir(worktree: string): string {
  return path.join(worktree, ".opencode", "decision-logs")
}

function ensureDecisionLogsDir(worktree: string): string {
  const dir = decisionLogsDir(worktree)
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true })
  }
  return dir
}

function getLogPath(worktree: string, sessionID: string): string {
  return path.join(ensureDecisionLogsDir(worktree), `${sessionID}.yaml`)
}

function loadLog(worktree: string, sessionID: string): DecisionLog {
  const logPath = getLogPath(worktree, sessionID)
  if (!fs.existsSync(logPath)) {
    return { session_id: sessionID, entries: [] }
  }
  const content = fs.readFileSync(logPath, "utf-8")
  try {
    return yaml.parse(content) as DecisionLog
  } catch {
    return { session_id: sessionID, entries: [] }
  }
}

function saveLog(worktree: string, log: DecisionLog): void {
  const logPath = getLogPath(worktree, log.session_id)
  try {
    fs.writeFileSync(logPath, yaml.stringify(log, { lineWidth: 0 }))
  } catch (err) {
    debugLog(worktree, 'saveLog: ERROR writing file:', String(err))
  }
}

function addEntry(worktree: string, sessionID: string, entry: LogEntry): void {
  const log = loadLog(worktree, sessionID)
  log.entries.push(entry)
  saveLog(worktree, log)
}

// --- Plugin ---

// Store client globally for use in tool execution
let globalClient: any = null

export const SupervisorPlugin: Plugin = async (ctx) => {
  const { worktree, client } = ctx
  
  // Store the client from plugin context for use in tool execution
  globalClient = client

  return {
    event: async ({ event }) => {
      // No action needed on events currently
    },
    tool: {
      supervisor_log: tool({
        description:
          "Log a decision or difficulty for supervisor analysis. " +
          "The supervisor will analyze the decision log and return guidance. " +
          "Use after: (1) taking an approach that differs from agreed plan, " +
          "(2) making a choice where user intent was ambiguous, " +
          "(3) repeatedly failing to make progress, " +
          "(4) being unsure what the user wants.",
        args: {
          type: tool.schema
            .enum(["decision", "difficulty"])
            .describe("Type of entry: decision or difficulty"),
          description: tool.schema
            .string()
            .describe("What happened or what you're deciding"),
          context: tool.schema
            .optional(tool.schema.record(tool.schema.string(), tool.schema.unknown()))
            .describe("Relevant context (e.g., what was agreed, what failed)"),
        },
        async execute(args, input) {
          const sessionIDToUse = input.sessionID || "default"
          // Use global client from plugin context instead of input.client (which doesn't exist)
          const client = globalClient

          // Add entry to log
          const entry: LogEntry = {
            timestamp: formatTimestamp(),
            type: args.type,
            description: args.description,
            context: args.context,
          }
          addEntry(worktree, sessionIDToUse, entry)

          // Load full log for analysis
          const log = loadLog(worktree, sessionIDToUse)

          // Prepare log summary for supervisor
          const logSummary = log.entries
            .slice(-10) // Last 10 entries
            .map((e) => {
              let line = `[${e.type}] ${e.timestamp}: ${e.description}`
              if (e.context?.agreed_plan) {
                line += ` (agreed: ${e.context.agreed_plan})`
              }
              if (e.context?.attempt_count) {
                line += ` (attempt #${e.context.attempt_count})`
              }
              return line
            })
            .join("\n")

          // Create supervisor session for analysis
          const supervisorPrompt = `You are a supervisor for an AI coding agent. Analyze the agent's recent decisions and difficulties to detect issues.

## Your job

Given the decision log below, determine if the agent is:
1. MISALIGNED - doing something different from what the user asked or agreed
2. LACK OF PROGRESS - not making meaningful forward motion
3. STALLED - repeating failed approaches without success

## Decision Log (recent)

${logSummary}

## Guidance

- If the agent is proceeding correctly: return { status: "proceed" }
- If there's misalignment, lack of progress, or the agent is stalled: 
  return { status: "intervene", question: "a question for the user to clarify" }

Output your response as JSON with this format:
{ "status": "proceed" | "intervene", "question": "optional question for user", "analysis": "brief explanation" }`

          try {
            // Create new session for supervisor analysis
            const supervisorSession = await client.session.create({
              body: { parentID: sessionIDToUse },
            })

            // Get the session ID
            const supSessionID = supervisorSession.data?.id
            if (!supSessionID) {
              return JSON.stringify({
                status: "proceed",
                analysis: "Unable to create supervisor session",
              })
            }

            // Send prompt to supervisor - use parts array format
            await client.session.prompt({
              path: { id: supSessionID },
              body: {
                noReply: true,
                parts: [{ type: 'text', text: supervisorPrompt }],
              },
            })

            // Wait for supervisor response with polling
            const maxAttempts = 20
            const delayMs = 500
            let assistantResponse: any = null

            for (let attempt = 0; attempt < maxAttempts; attempt++) {
              // Get the supervisor's response
              const messages = await client.session.messages({
                path: { id: supSessionID },
                query: { limit: 10 },
              })

              // Look for assistant message
              const assistantMessages = messages.data?.filter(
                (m: any) => m.role === "assistant"
              )
              
              if (assistantMessages && assistantMessages.length > 0) {
                assistantResponse = assistantMessages[assistantMessages.length - 1]
                break
              }

              // Wait before next poll
              await new Promise(resolve => setTimeout(resolve, delayMs))
            }

            if (!assistantResponse) {
              return JSON.stringify({
                status: "proceed",
                analysis: "Supervisor did not respond in time",
              })
            }

            // Parse the supervisor's response
            const content = assistantResponse.parts
              ?.map((p: any) => p.text || "")
              .join("") || ""

            // Try to parse JSON response
            let result = { status: "proceed" as const, analysis: "" }
            try {
              const match = content.match(/\{[\s\S]*\}/)
              if (match) {
                result = JSON.parse(match[0])
              }
            } catch {
              result.analysis = content.substring(0, 200)
            }

            // Add supervisor analysis to log
            addEntry(worktree, sessionIDToUse, {
              timestamp: formatTimestamp(),
              type: "supervisor_analysis",
              description: result.analysis,
              status: result.status,
              question: result.question,
            })

            // Return result to agent
            return JSON.stringify({
              status: result.status,
              question: result.question,
              analysis: result.analysis,
            })
          } catch (err) {
            console.error("[supervisor-plugin] Error:", err)
            return JSON.stringify({
              status: "proceed",
              analysis: "Supervisor analysis failed",
            })
          }
        },
      }),
    },

    // Hook: tool.execute.after for question tool - capture user answers
    "tool.execute.after": async (input, output) => {
      // Add null check for output - can be undefined in some cases
      if (!output) return
      
      if (input.tool === "question" && input.sessionID) {
        const args = input.args as any
        const questions = args?.questions || []

        // Extract user answers from output.metadata.answers (primary) or output.output (fallback)
        let userAnswers: string[] | undefined
        
        // Primary: metadata.answers is an array of arrays (one per question)
        const metadataAnswers = output?.metadata?.answers
        if (metadataAnswers && Array.isArray(metadataAnswers) && metadataAnswers.length > 0) {
          userAnswers = metadataAnswers.map((a: any) => 
            Array.isArray(a) ? a.join("\n") : String(a)
          )
        }
        
        // Fallback: parse from output.output string
        if (!userAnswers || userAnswers.length === 0) {
          userAnswers = output?.output
            ?.split(/\n(?=Answer:)/)
            ?.map((s: string) => s.replace(/^Answer:\s*/i, "").trim())
        }

        // Log each user's answer
        for (let i = 0; i < questions.length; i++) {
          const q = questions[i]
          const answer = userAnswers?.[i] || userAnswers?.[0] || "no answer captured"
          addEntry(worktree, input.sessionID, {
            timestamp: formatTimestamp(),
            type: "user_input",
            description: `Answered question: ${q.header || q.question}`,
            context: { question: q.question, answer },
          })
        }
      }
    },

    // Hook: chat.message - messages sent to/from LLM
    // The message object doesn't have content - text is in the parts array
    "chat.message": async (input: any, { message, parts }: any) => {
      const sessionID = input?.sessionID || 'unknown'
      
      // Extract text from parts array (not message.content which doesn't exist)
      const messageText = (parts || [])
        .filter((p: any) => p.type === 'text')
        .map((p: any) => p.text || '')
        .join('')
        .substring(0, 500)
      
      if (!messageText) return

      // Log user messages
      if (message?.role === "user" && sessionID !== 'unknown') {
        addEntry(worktree, sessionID, {
          timestamp: formatTimestamp(),
          type: "user_input",
          description: `User message: ${messageText}`,
        })
      }
    },
  }
}