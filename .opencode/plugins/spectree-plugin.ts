/**
 * Spectree Plugin
 *
 * Provides tools for the spectree recursive decomposition workflow.
 * Implemented as a plugin (not standalone tools) to access the SDK client
 * for session parent lookup - enabling automatic node identity via sessionID.
 *
 * Tools:
 * - spectree_write: Create new spectree spec with root node (ID = sessionID)
 * - spectree_read: Read spec or specific node
 * - spectree_register_node: Child self-registers under parent (ID = sessionID, parent = parentID)
 * - spectree_update: Update own node (identity via sessionID)
 * - spectree_get_my_node: Get own node (identity via sessionID)
 * - spectree_get_leaves: Get leaf nodes in depth-first order
 */

import type { Plugin } from "@opencode-ai/plugin"
import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

// --- Schema ---

const NodeSchema: z.ZodType<Node> = z.object({
  id: z.string().describe("Unique node identifier (sessionID of owning agent)"),
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

const SpectreeSpecSchema = z.object({
  title: z.string().describe("Spec title"),
  version: z.number().describe("Spec version"),
  root: NodeSchema.describe("Root node of the spec tree"),
})

// --- Types ---

type Node = {
  id: string
  title: string
  description: string
  children: Node[]
  interaction_log: { question: string; answer: string; timestamp: string }[]
  tests: {
    description: string
    type: "unit" | "integration" | "e2e"
    target_component: string
    contracts: string[]
    given_when_then: string
  }[]
  test_status?: "pending" | "in_progress" | "passed" | "failed"
  file_changes: {
    action: "create" | "modify" | "delete"
    description: string
    is_test: boolean
    path: string
  }[]
  impl_status?: "pending" | "in_progress" | "completed"
  reviews?: {
    reviewer: string
    feedback: string
    status: "pending" | "passed" | "failed"
    timestamp: string
  }[]
}

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
  return yaml.stringify(obj, {
    lineWidth: 0,
    defaultStringType: "QUOTE_DOUBLE",
    defaultKeyType: "PLAIN",
  })
}

function spectreeDir(worktree: string): string {
  return path.join(worktree, "spectree")
}

function findNodeById(node: Node, id: string): Node | null {
  if (node.id === id) return node
  for (const child of node.children) {
    const found = findNodeById(child, id)
    if (found) return found
  }
  return null
}

function findParentOfNode(root: Node, childId: string): Node | null {
  for (const child of root.children) {
    if (child.id === childId) return root
    const found = findParentOfNode(child, childId)
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
    children: node.children.map((child) =>
      findAndUpdateNode(child, id, updates)
    ),
  }
}

function addChildToNode(root: Node, parentId: string, child: Node): Node {
  if (root.id === parentId) {
    return { ...root, children: [...root.children, child] }
  }
  return {
    ...root,
    children: root.children.map((c) => addChildToNode(c, parentId, child)),
  }
}

function getLeafNodes(node: Node): Node[] {
  if (node.children.length === 0) {
    return [node]
  }
  return node.children.flatMap(getLeafNodes)
}

function loadSpec(worktree: string): z.infer<typeof SpectreeSpecSchema> {
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
    throw new Error(
      `Invalid spectree spec: ${formatValidationErrors(specResult.error)}`
    )
  }

  return specResult.data
}

function saveSpec(
  worktree: string,
  spec: z.infer<typeof SpectreeSpecSchema>
): void {
  const specPath = path.join(spectreeDir(worktree), "spectree.yaml")

  const result = SpectreeSpecSchema.safeParse(spec)
  if (!result.success) {
    throw new Error(
      `Invalid spec after update: ${formatValidationErrors(result.error)}`
    )
  }

  fs.writeFileSync(specPath, toYamlString(spec))
}

function makeEmptyNode(id: string, title: string, description: string): Node {
  return {
    id,
    title,
    description,
    children: [],
    interaction_log: [],
    tests: [],
    file_changes: [],
  }
}

// --- Plugin ---

export const SpectreePlugin: Plugin = async (ctx) => {
  const { client } = ctx

  return {
    tool: {
      spectree_write: tool({
        description:
          "Create a new spectree spec file. " +
          "Initializes a spectree.yaml in the spectree/ directory. " +
          "The root node ID is automatically set to the current sessionID.",
        args: {
          title: tool.schema.string().describe("Title of the spectree spec"),
          root_node: tool.schema
            .object({
              title: tool.schema.string(),
              description: tool.schema.string(),
            })
            .describe("Root node data (ID is auto-assigned from sessionID)"),
        },
        async execute(args, context) {
          const { worktree, sessionID } = context
          const dir = spectreeDir(worktree)

          if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true })
          }

          const root = makeEmptyNode(
            sessionID,
            args.root_node.title,
            args.root_node.description
          )

          const spec = {
            title: args.title,
            version: 1,
            root,
          }

          const parsed = SpectreeSpecSchema.safeParse(spec)
          if (!parsed.success) {
            throw new Error(
              `Invalid spec: ${formatValidationErrors(parsed.error)}`
            )
          }

          const specPath = path.join(dir, "spectree.yaml")
          fs.writeFileSync(specPath, toYamlString(spec))

          return {
            success: true,
            title: args.title,
            root_node_id: sessionID,
          }
        },
      }),

      spectree_read: tool({
        description:
          "Read the spectree spec file. " +
          "Returns the full spec or a specific node by ID if provided.",
        args: {
          node_id: tool.schema
            .string()
            .optional()
            .describe("Optional node ID to read specific node"),
        },
        async execute(args, context) {
          const { worktree } = context
          const spec = loadSpec(worktree)

          if (args.node_id) {
            const node = findNodeById(spec.root, args.node_id)
            if (!node) {
              throw new Error(`Node with id "${args.node_id}" not found`)
            }
            return node
          }

          return spec
        },
      }),

      spectree_register_node: tool({
        description:
          "Register a new child node in the spectree spec. " +
          "The node ID is automatically set to the current sessionID. " +
          "The parent node is automatically determined from the session's parent session. " +
          "MUST be called as the first spectree action by any node subagent.",
        args: {
          title: tool.schema.string().describe("Title of this node"),
          description: tool.schema
            .string()
            .describe("Description of the sub-problem this node handles"),
        },
        async execute(args, context) {
          const { worktree, sessionID } = context

          // Look up parent session via SDK
          const sessionResult = await client.session.get({
            path: { id: sessionID },
          })

          const parentSessionID = sessionResult.data?.parentID
          if (!parentSessionID) {
            throw new Error(
              "Cannot register node: no parent session found. " +
                "This tool must be called from a subagent spawned by a spectree or node agent."
            )
          }

          const spec = loadSpec(worktree)

          // Check if already registered
          const existing = findNodeById(spec.root, sessionID)
          if (existing) {
            return {
              success: true,
              node_id: sessionID,
              parent_id: parentSessionID,
              already_registered: true,
            }
          }

          // Find parent node in spec
          const parentNode = findNodeById(spec.root, parentSessionID)
          if (!parentNode) {
            throw new Error(
              `Parent node with id "${parentSessionID}" not found in spec. ` +
                "The parent agent must have created the spec or registered itself first."
            )
          }

          // Create and add child node
          const childNode = makeEmptyNode(sessionID, args.title, args.description)
          const updatedRoot = addChildToNode(
            spec.root,
            parentSessionID,
            childNode
          )
          const updatedSpec = { ...spec, root: updatedRoot }

          saveSpec(worktree, updatedSpec)

          return {
            success: true,
            node_id: sessionID,
            parent_id: parentSessionID,
          }
        },
      }),

      spectree_update: tool({
        description:
          "Update a node in the spectree spec. " +
          "Can only update the current agent's own node (enforced via sessionID).",
        args: {
          updates: tool.schema
            .record(tool.schema.string(), tool.schema.unknown())
            .describe("Fields to update on the node"),
        },
        async execute(args, context) {
          const { worktree, sessionID } = context
          const spec = loadSpec(worktree)

          // Verify node exists and belongs to this session
          const node = findNodeById(spec.root, sessionID)
          if (!node) {
            throw new Error(
              `No node found for sessionID "${sessionID}". ` +
                "Call spectree_register_node first, or use spectree_write for the root."
            )
          }

          // Update the node
          const updatedRoot = findAndUpdateNode(spec.root, sessionID, args.updates)
          const updatedSpec = { ...spec, root: updatedRoot }

          saveSpec(worktree, updatedSpec)

          return { success: true, node_id: sessionID }
        },
      }),

      spectree_get_my_node: tool({
        description:
          "Get the current agent's node from the spectree spec. " +
          "Automatically identifies the node via the current sessionID.",
        args: {},
        async execute(args, context) {
          const { worktree, sessionID } = context
          const spec = loadSpec(worktree)
          const node = findNodeById(spec.root, sessionID)

          if (!node) {
            throw new Error(
              `No node found for sessionID "${sessionID}". ` +
                "Call spectree_register_node first, or use spectree_write for the root."
            )
          }

          return node
        },
      }),

      spectree_get_leaves: tool({
        description:
          "Get all leaf nodes from the spectree spec in depth-first order. " +
          "Returns array of leaf nodes ordered left-to-right.",
        args: {},
        async execute(args, context) {
          const { worktree } = context
          const spec = loadSpec(worktree)
          return getLeafNodes(spec.root)
        },
      }),
    },
  }
}
