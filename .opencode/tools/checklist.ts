import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

// --- Schema ---

const ChecklistItemTypeSchema = z.enum(["question", "task"])
const ChecklistItemSchema = z.object({
  id: z.number().describe("Unique identifier for this item"),
  type: ChecklistItemTypeSchema.describe("Type of item"),
  description: z.string().min(1).describe("What needs to be addressed"),
  created_at: z.string().describe("ISO timestamp when item was created"),
  resolved: z.boolean().describe("Whether this item has been resolved"),
  resolved_at: z.string().optional().describe("ISO timestamp when item was resolved"),
  resolution_note: z.string().optional().describe("Optional note about how it was resolved"),
})

const ChecklistSchema = z.object({
  version: z.literal(1).describe("Checklist schema version"),
  session_id: z.string().describe("Associated session ID"),
  workflow_type: z.enum(["spec", "vibe", "fix", "document"]).describe("Type of workflow"),
  items: z.array(ChecklistItemSchema).describe("Checklist items"),
  created_at: z.string().describe("ISO timestamp when checklist was created"),
})

type Checklist = z.infer<typeof ChecklistSchema>
type ChecklistItem = z.infer<typeof ChecklistItemSchema>

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

function getChecklistPath(worktree: string, sessionId: string): string {
  return path.join(worktree, ".opencode", "sessions", `${sessionId}-checklist.yaml`)
}

function getSessionPath(worktree: string, sessionId: string): string | null {
  const workflows = ["spec", "vibe", "fix", "document"]
  for (const wf of workflows) {
    const candidate = path.join(worktree, ".opencode", "sessions", wf, `${sessionId}.yaml`)
    if (fs.existsSync(candidate)) {
      return candidate
    }
  }
  return null
}

function loadOrCreateChecklist(worktree: string, sessionId: string): Checklist {
  const checklistPath = getChecklistPath(worktree, sessionId)
  
  if (fs.existsSync(checklistPath)) {
    const content = fs.readFileSync(checklistPath, "utf-8")
    const parsed = parseYaml(content)
    if (parsed) {
      return parsed as Checklist
    }
  }
  
  // Determine workflow type from session
  const sessionPath = getSessionPath(worktree, sessionId)
  let workflowType: Checklist["workflow_type"] = "vibe" // default
  
  if (sessionPath) {
    const sessionContent = fs.readFileSync(sessionPath, "utf-8")
    const sessionParsed = parseYaml(sessionContent)
    if (sessionParsed && typeof sessionParsed === "object" && "workflow_type" in sessionParsed) {
      workflowType = (sessionParsed as { workflow_type: string }).workflow_type as Checklist["workflow_type"]
    }
  }
  
  return {
    version: 1,
    session_id: sessionId,
    workflow_type: workflowType,
    items: [],
    created_at: new Date().toISOString(),
  }
}

function saveChecklist(worktree: string, checklist: Checklist): void {
  const checklistPath = getChecklistPath(worktree, checklist.session_id)
  
  // Ensure directory exists
  const dir = path.dirname(checklistPath)
  fs.mkdirSync(dir, { recursive: true })
  
  const yamlContent = toYamlString(checklist)
  fs.writeFileSync(checklistPath, yamlContent)
}

function formatChecklist(checklist: Checklist, showResolved = false): string {
  const pending = checklist.items.filter((i) => !i.resolved)
  const resolved = checklist.items.filter((i) => i.resolved)
  
  let output = `CHECKLIST (Session: ${checklist.session_id})\n`
  output += `Workflow: ${checklist.workflow_type}\n`
  output += `Created: ${checklist.created_at}\n\n`
  
  if (pending.length === 0) {
    output += `✓ No pending items\n`
  } else {
    output += `PENDING ITEMS (${pending.length}):\n`
    for (const item of pending) {
      output += `[${item.id}] [${item.type.toUpperCase()}] ${item.description}\n`
    }
  }
  
  if (showResolved && resolved.length > 0) {
    output += `\nRESOLVED (${resolved.length}):\n`
    for (const item of resolved) {
      output += `[${item.id}] [${item.type.toUpperCase()}] ${item.description}\n`
      if (item.resolution_note) {
        output += `    → ${item.resolution_note}\n`
      }
    }
  }
  
  return output
}

// --- Tools ---

export const checklist_tick = tool({
  description:
    "Add an item to the session's checklist. " +
    "Use this to track questions that need answering or tasks discovered during implementation. " +
    "Items must be resolved via checklist_complete before proceeding past a phase boundary.",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID from session_start"),
    type: tool.schema
      .enum(["question", "task"])
      .describe("Type of item to track"),
    description: tool.schema
      .string()
      .min(1)
      .describe("What needs to be addressed - be specific about what question needs answering or what task needs doing"),
  },
  async execute(args, context) {
    const checklist = loadOrCreateChecklist(context.worktree, args.session_id)
    
    const nextId = checklist.items.length > 0 
      ? Math.max(...checklist.items.map((i) => i.id)) + 1 
      : 1
    
    const item: ChecklistItem = {
      id: nextId,
      type: args.type,
      description: args.description,
      created_at: new Date().toISOString(),
      resolved: false,
    }
    
    checklist.items.push(item)
    saveChecklist(context.worktree, checklist)
    
    const pendingCount = checklist.items.filter((i) => !i.resolved).length
    
    return `Item added to checklist.

Session: ${args.session_id}
Item: [${item.id}] [${item.type.toUpperCase()}] ${item.description}

Total pending items: ${pendingCount}

Use checklist_complete to resolve this item, or checklist_status to view all items.`
  },
})

export const checklist_complete = tool({
  description:
    "Mark a checklist item as resolved/completed. " +
    "Provide a resolution_note to explain how the item was addressed " +
    "(e.g., 'decided to use Redis', 'deferred to separate issue #123', 'not needed - clarified with user').",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID from session_start"),
    item_id: tool.schema
      .number()
      .describe("ID of the item to resolve (from checklist_status output)"),
    resolution_note: tool.schema
      .string()
      .optional()
      .describe("Optional note explaining how this item was addressed"),
  },
  async execute(args, context) {
    const checklist = loadOrCreateChecklist(context.worktree, args.session_id)
    
    const itemIndex = checklist.items.findIndex((i) => i.id === args.item_id)
    
    if (itemIndex === -1) {
      const validIds = checklist.items.map((i) => i.id).join(", ")
      return `ERROR: Item ${args.item_id} not found.
      
Valid item IDs: ${validIds || "(no items)"}

Use checklist_status to see all items.`
    }
    
    const item = checklist.items[itemIndex]
    item.resolved = true
    item.resolved_at = new Date().toISOString()
    if (args.resolution_note) {
      item.resolution_note = args.resolution_note
    }
    
    saveChecklist(context.worktree, checklist)
    
    const pendingCount = checklist.items.filter((i) => !i.resolved).length
    
    return `Item resolved.

Session: ${args.session_id}
Item: [${item.id}] [${item.type.toUpperCase()}] ${item.description}
Resolved: ${item.resolved_at}
${item.resolution_note ? `Note: ${item.resolution_note}` : ""}

Total pending items: ${pendingCount}`
  },
})

export const checklist_status = tool({
  description:
    "Check the current status of all checklist items for a session. " +
    "Use this to verify all items are resolved before proceeding past a phase boundary. " +
    "Returns a summary showing pending items and optionally resolved items.",
  args: {
    session_id: tool.schema
      .string()
      .describe("Session ID from session_start"),
    show_resolved: tool.schema
      .boolean()
      .optional()
      .default(false)
      .describe("Whether to show resolved items (default: false)"),
  },
  async execute(args, context) {
    const checklist = loadOrCreateChecklist(context.worktree, args.session_id)
    
    const pending = checklist.items.filter((i) => !i.resolved)
    const resolved = checklist.items.filter((i) => i.resolved)
    
    let output = formatChecklist(checklist, args.show_resolved || false)
    
    if (pending.length > 0) {
      output += `\n⚠️  ${pending.length} item(s) must be resolved before proceeding to next phase.`
      output += `\nUse checklist_complete <item_id> resolution_note="..." to resolve items.`
    } else {
      output += `\n✓ All items resolved. Safe to proceed to next phase.`
    }
    
    return output
  },
})
