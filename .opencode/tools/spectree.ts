import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

// --- Schema ---

// Node schema for spectree spec structure
const NodeSchema = z.object({
  id: z.string().describe("Unique node identifier"),
  title: z.string().describe("Node title"),
  description: z.string().describe("Node description"),
  children: z.array(z.lazy(() => NodeSchema)).describe("Child nodes"),
  interaction_log: z
    .array(
      z.object({
        question: z.string(),
        answer: z.string(),
        timestamp: z.string(),
      })
    )
    .describe("Log of user interactions"),
  tests: z
    .array(
      z.object({
        description: z.string(),
        type: z.enum(["unit", "integration", "e2e"]),
        target_component: z.string(),
        contracts: z.array(z.string()),
        given_when_then: z.string(),
      })
    )
    .describe("Test cases for this node"),
  test_status: z
    .enum(["pending", "in_progress", "passed", "failed"])
    .optional()
    .describe("Status of test execution"),
  file_changes: z
    .array(
      z.object({
        action: z.enum(["create", "modify", "delete"]),
        description: z.string(),
        is_test: z.boolean(),
        path: z.string(),
      })
    )
    .describe("File changes associated with this node"),
  impl_status: z
    .enum(["pending", "in_progress", "completed"])
    .optional()
    .describe("Implementation status"),
  reviews: z
    .array(
      z.object({
        reviewer: z.string(),
        feedback: z.string(),
        status: z.enum(["pending", "passed", "failed"]),
        timestamp: z.string(),
      })
    )
    .optional()
    .describe("Code reviews for this node"),
})

// Full spectree spec schema
const SpectreeSpecSchema = z.object({
  title: z.string().describe("Spec title"),
  version: z.number().describe("Spec version"),
  root: NodeSchema.describe("Root node of the spec tree"),
})

// --- Helpers ---

function formatValidationErrors(error: z.ZodError): string {
  return error.issues
    .map((issue) => {
      const p = issue.path.join(".")
      return `  - ${p ? p + ": " : ""}${issue.message}`
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

function spectreeDir(worktree: string): string {
  return path.join(worktree, "spectree")
}

function getAgentId(worktree: string): string | null {
  const agentIdPath = path.join(worktree, ".opencode", "agent_id")
  if (fs.existsSync(agentIdPath)) {
    return fs.readFileSync(agentIdPath, "utf-8").trim()
  }
  return null
}

type Node = z.infer<typeof NodeSchema>

function findNodeById(node: Node, id: string): Node | null {
  if (node.id === id) return node
  for (const child of node.children) {
    const found = findNodeById(child, id)
    if (found) return found
  }
  return null
}

function findAndUpdateNode(
  node: Node,
  id: string,
  updates: Record<string, unknown>
): Node {
  if (node.id === id) {
    return { ...node, ...updates }
  }
  return {
    ...node,
    children: node.children.map((child) => findAndUpdateNode(child, id, updates)),
  }
}

function getLeafNodes(node: Node): Node[] {
  if (node.children.length === 0) {
    return [node]
  }
  return node.children.flatMap(getLeafNodes)
}

// --- Tools ---

export const spectree_write = tool({
  description:
    "Create a new spectree spec file. " +
    "Initializes a spectree.yaml in the spectree/ directory with the given title and root node structure.",
  args: {
    title: tool.schema.string().describe("Title of the spectree spec"),
    root_node: tool.schema
      .object({
        id: tool.schema.string(),
        title: tool.schema.string(),
        description: tool.schema.string(),
      })
      .describe("Root node data"),
  },
  async execute(args, context) {
    const { worktree } = context
    const dir = spectreeDir(worktree)

    // Create directory if it doesn't exist
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true })
    }

    // Create the root node with default empty arrays
    const root = {
      id: args.root_node.id,
      title: args.root_node.title,
      description: args.root_node.description,
      children: [],
      interaction_log: [],
      tests: [],
      file_changes: [],
    }

    const spec = {
      title: args.title,
      version: 1,
      root,
    }

    // Validate
    const parsed = SpectreeSpecSchema.safeParse(spec)
    if (!parsed.success) {
      throw new Error(`Invalid spec: ${formatValidationErrors(parsed.error)}`)
    }

    // Write to file
    const specPath = path.join(dir, "spectree.yaml")
    fs.writeFileSync(specPath, toYamlString(spec))

    return { success: true, title: args.title }
  },
})

export const spectree_read = tool({
  description:
    "Read the spectree spec file. " +
    "Returns the full spec or a specific node by ID if provided.",
  args: {
    node_id: tool.schema.string().optional().describe("Optional node ID to read specific node"),
  },
  async execute(args, context) {
    const { worktree } = context
    const specPath = path.join(spectreeDir(worktree), "spectree.yaml")

    if (!fs.existsSync(specPath)) {
      throw new Error("spectree.yaml not found. Use spectree_write to create one.")
    }

    const content = fs.readFileSync(specPath, "utf-8")
    const parsed = parseYaml(content)
    if (!parsed) {
      throw new Error("Invalid YAML in spectree.yaml")
    }

    const specResult = SpectreeSpecSchema.safeParse(parsed)
    if (!specResult.success) {
      throw new Error(`Invalid spectree spec: ${formatValidationErrors(specResult.error)}`)
    }

    const spec = specResult.data

    if (args.node_id) {
      const node = findNodeById(spec.root, args.node_id)
      if (!node) {
        throw new Error(`Node with id "${args.node_id}" not found`)
      }
      return node
    }

    return spec
  },
})

export const spectree_update = tool({
  description:
    "Update a node in the spectree spec. " +
    "Can only update the current agent's own node (enforced by tool).",
  args: {
    node_id: tool.schema.string().describe("ID of node to update"),
    updates: tool.schema
      .record(tool.schema.string(), tool.schema.unknown())
      .describe("Fields to update"),
  },
  async execute(args, context) {
    const { worktree } = context

    // Get agent ID and enforce ownership
    const agentId = getAgentId(worktree)
    if (!agentId) {
      throw new Error("No agent_id found. Run in agent context.")
    }

    if (args.node_id !== agentId) {
      throw new Error(
        `Agent can only update their own node. Agent ID: ${agentId}, requested: ${args.node_id}`
      )
    }

    const specPath = path.join(spectreeDir(worktree), "spectree.yaml")

    if (!fs.existsSync(specPath)) {
      throw new Error("spectree.yaml not found. Use spectree_write to create one.")
    }

    const content = fs.readFileSync(specPath, "utf-8")
    const parsed = parseYaml(content)
    if (!parsed) {
      throw new Error("Invalid YAML in spectree.yaml")
    }

    const specResult = SpectreeSpecSchema.safeParse(parsed)
    if (!specResult.success) {
      throw new Error(`Invalid spectree spec: ${formatValidationErrors(specResult.error)}`)
    }

    const spec = specResult.data

    // Check node exists
    const node = findNodeById(spec.root, args.node_id)
    if (!node) {
      throw new Error(`Node with id "${args.node_id}" not found`)
    }

    // Update the node
    const updatedRoot = findAndUpdateNode(spec.root, args.node_id, args.updates)
    const updatedSpec = { ...spec, root: updatedRoot }

    // Validate
    const newSpecResult = SpectreeSpecSchema.safeParse(updatedSpec)
    if (!newSpecResult.success) {
      throw new Error(
        `Invalid spec after update: ${formatValidationErrors(newSpecResult.error)}`
      )
    }

    // Write back
    fs.writeFileSync(specPath, toYamlString(updatedSpec))

    return { success: true, node_id: args.node_id }
  },
})

export const spectree_get_my_node = tool({
  description:
    "Get the current agent's node from the spectree spec. " +
    "Reads agent_id from .opencode/agent_id file and returns the matching node.",
  args: {},
  async execute(args, context) {
    const { worktree } = context

    // Get agent ID
    const agentId = getAgentId(worktree)
    if (!agentId) {
      throw new Error("No agent_id found. Run in agent context.")
    }

    const specPath = path.join(spectreeDir(worktree), "spectree.yaml")

    if (!fs.existsSync(specPath)) {
      throw new Error("spectree.yaml not found. Use spectree_write to create one.")
    }

    const content = fs.readFileSync(specPath, "utf-8")
    const parsed = parseYaml(content)
    if (!parsed) {
      throw new Error("Invalid YAML in spectree.yaml")
    }

    const specResult = SpectreeSpecSchema.safeParse(parsed)
    if (!specResult.success) {
      throw new Error(`Invalid spectree spec: ${formatValidationErrors(specResult.error)}`)
    }

    const spec = specResult.data
    const node = findNodeById(spec.root, agentId)

    if (!node) {
      throw new Error(`Node with id "${agentId}" not found`)
    }

    return node
  },
})

export const spectree_get_leaves = tool({
  description:
    "Get all leaf nodes from the spectree spec in depth-first order. " +
    "Returns array of leaf nodes ordered left-to-right.",
  args: {},
  async execute(args, context) {
    const { worktree } = context
    const specPath = path.join(spectreeDir(worktree), "spectree.yaml")

    if (!fs.existsSync(specPath)) {
      throw new Error("spectree.yaml not found. Use spectree_write to create one.")
    }

    const content = fs.readFileSync(specPath, "utf-8")
    const parsed = parseYaml(content)
    if (!parsed) {
      throw new Error("Invalid YAML in spectree.yaml")
    }

    const specResult = SpectreeSpecSchema.safeParse(parsed)
    if (!specResult.success) {
      throw new Error(`Invalid spectree spec: ${formatValidationErrors(specResult.error)}`)
    }

    const spec = specResult.data
    return getLeafNodes(spec.root)
  },
})
