/**
 * Spec-Tree Plugin
 *
 * Provides tools for the spec-tree recursive decomposition workflow.
 * Implemented as a plugin (not standalone tools) to access the SDK client
 * for explicit parent ID handling - enabling robust node identity.
 *
 * Tools:
 * - spec_tree_write: Create new spec-tree spec with root node
 * - spec_tree_read: Read spec or specific node
 * - spec_tree_register_node: Register child node under explicit parent
 * - spec_tree_update: Update own node
 * - spec_tree_get_my_node: Get own node
 * - spec_tree_get_leaves: Get leaf nodes in depth-first order
 */

import type { Plugin } from "@opencode-ai/plugin"
import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

// --- Schema ---

const NodeSchema: z.ZodType<Node> = z.object({
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

const SpecTreeSpecSchema = z.object({
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

function specTreeDir(worktree: string): string {
  return path.join(worktree, "spec-tree")
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

function loadSpec(worktree: string): z.infer<typeof SpecTreeSpecSchema> {
  const specPath = path.join(specTreeDir(worktree), "spec-tree.yaml")

  if (!fs.existsSync(specPath)) {
    throw new Error("spec-tree.yaml not found. Use spec_tree_write to create one.")
  }

  const content = fs.readFileSync(specPath, "utf-8")
  const parsed = parseYaml(content)
  if (!parsed) {
    throw new Error("Invalid YAML in spec-tree.yaml")
  }

  const specResult = SpecTreeSpecSchema.safeParse(parsed)
  if (!specResult.success) {
    throw new Error(
      `Invalid spec-tree spec: ${formatValidationErrors(specResult.error)}`
    )
  }

  return specResult.data
}

function saveSpec(
  worktree: string,
  spec: z.infer<typeof SpecTreeSpecSchema>
): void {
  const specPath = path.join(specTreeDir(worktree), "spec-tree.yaml")

  const result = SpecTreeSpecSchema.safeParse(spec)
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

export const SpecTreePlugin: Plugin = async (ctx) => {
  return {
    tool: {
      spec_tree_write: tool({
        description:
          "Create a new spec-tree spec file. " +
          "Initializes a spec-tree.yaml in the spec-tree/ directory. " +
          "Returns the root node ID for reference.",
        args: {
          title: tool.schema.string().describe("Title of the spec-tree spec"),
          root_node: tool.schema
            .object({
              id: tool.schema.string().describe("Unique ID for root node"),
              title: tool.schema.string(),
              description: tool.schema.string(),
            })
            .describe("Root node data"),
        },
        async execute(args, context) {
          const { worktree } = context
          const dir = specTreeDir(worktree)

          if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true })
          }

          const root = makeEmptyNode(
            args.root_node.id,
            args.root_node.title,
            args.root_node.description
          )

          const spec = {
            title: args.title,
            version: 1,
            root,
          }

          const parsed = SpecTreeSpecSchema.safeParse(spec)
          if (!parsed.success) {
            throw new Error(
              `Invalid spec: ${formatValidationErrors(parsed.error)}`
            )
          }

          const specPath = path.join(dir, "spec-tree.yaml")
          fs.writeFileSync(specPath, toYamlString(spec))

          return {
            success: true,
            title: args.title,
            root_node_id: args.root_node.id,
          }
        },
      }),

      spec_tree_read: tool({
        description:
          "Read the spec-tree spec file. " +
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

      spec_tree_register_node: tool({
        description:
          "Register a new child node in the spec-tree spec. " +
          "The node ID is passed explicitly, and parent_id must be provided to link to the correct parent.",
        args: {
          id: tool.schema.string().describe("Unique ID for this node"),
          parent_id: tool.schema
            .string()
            .describe("ID of the parent node this child belongs to"),
          title: tool.schema.string().describe("Title of this node"),
          description: tool.schema
            .string()
            .describe("Description of the sub-problem this node handles"),
        },
        async execute(args, context) {
          const { worktree } = context

          const spec = loadSpec(worktree)

          // Check if already registered
          const existing = findNodeById(spec.root, args.id)
          if (existing) {
            return {
              success: true,
              node_id: args.id,
              parent_id: args.parent_id,
              already_registered: true,
            }
          }

          // Find parent node in spec
          const parentNode = findNodeById(spec.root, args.parent_id)
          if (!parentNode) {
            throw new Error(
              `Parent node with id "${args.parent_id}" not found in spec. ` +
                "The parent must exist in the spec before registering children."
            )
          }

          // Create and add child node
          const childNode = makeEmptyNode(args.id, args.title, args.description)
          const updatedRoot = addChildToNode(
            spec.root,
            args.parent_id,
            childNode
          )
          const updatedSpec = { ...spec, root: updatedRoot }

          saveSpec(worktree, updatedSpec)

          return {
            success: true,
            node_id: args.id,
            parent_id: args.parent_id,
          }
        },
      }),

      spec_tree_update: tool({
        description:
          "Update a node in the spec-tree spec. " +
          "Can only update a node if the node_id is provided.",
        args: {
          node_id: tool.schema
            .string()
            .describe("ID of the node to update"),
          updates: tool.schema
            .record(tool.schema.string(), tool.schema.unknown())
            .describe("Fields to update on the node"),
        },
        async execute(args, context) {
          const { worktree } = context
          const spec = loadSpec(worktree)

          // Verify node exists
          const node = findNodeById(spec.root, args.node_id)
          if (!node) {
            throw new Error(
              `No node found with id "${args.node_id}". ` +
                "Call spec_tree_register_node first, or use spec_tree_write for the root."
            )
          }

          // Update the node
          const updatedRoot = findAndUpdateNode(spec.root, args.node_id, args.updates)
          const updatedSpec = { ...spec, root: updatedRoot }

          saveSpec(worktree, updatedSpec)

          return { success: true, node_id: args.node_id }
        },
      }),

      spec_tree_get_my_node: tool({
        description:
          "Get a specific node from the spec-tree spec by ID.",
        args: {
          node_id: tool.schema
            .string()
            .describe("ID of the node to retrieve"),
        },
        async execute(args, context) {
          const { worktree } = context
          const spec = loadSpec(worktree)
          const node = findNodeById(spec.root, args.node_id)

          if (!node) {
            throw new Error(
              `No node found with id "${args.node_id}". ` +
                "Call spec_tree_register_node first, or use spec_tree_write for the root."
            )
          }

          return node
        },
      }),

      spec_tree_get_leaves: tool({
        description:
          "Get all leaf nodes from the spec-tree spec in depth-first order. " +
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