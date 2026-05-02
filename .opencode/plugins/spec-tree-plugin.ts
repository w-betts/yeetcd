/**
 * Spec-Tree Plugin
 *
 * Provides tools for the spec-tree recursive decomposition workflow.
 * Implemented as a plugin (not standalone tools) to access the SDK client
 * for explicit parent ID handling - enabling robust node identity.
 *
 * Tools:
 * - spec_tree_write: Create new spec-tree spec with root node (saved to .opencode/spec-trees/)
 * - spec_tree_read: Read spec or specific node
 * - spec_tree_register_node: Register child node under explicit parent
 * - spec_tree_update: Update own node
 * - spec_tree_get_my_node: Get own node
 * - spec_tree_get_leaves: Get leaf nodes in depth-first order
 * - spec_tree_render_ascii: Render spec-tree as ASCII visualization
 * - spec_tree_list: List all spec-tree files in the repository
 * - spec_tree_use: Switch the active spec-tree for the current worktree
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
  node_type: z.enum(["unexpanded", "branch", "leaf"]).optional().describe("Explicit node type: unexpanded (not yet explored), branch (has children, decomposed), leaf (implementation unit)"),
  depends_on: z.array(z.string()).optional().describe("Node IDs this node depends on for implementation"),
  planning_status: z.enum(["pending", "exploring", "ready"]).optional().describe("Planning status: pending (not started), exploring (in progress), ready (decomposition complete)"),
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
  node_type?: "unexpanded" | "branch" | "leaf"
  depends_on?: string[]
  planning_status?: "pending" | "exploring" | "ready"
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

const SPEC_TREES_DIR = ".opencode/spec-trees"
const ACTIVE_SPEC_POINTER = ".active-spec"

function formatValidationErrors(error: z.ZodError): string {
  return error.issues
    .map((issue) => {
      const p = issue.path.join(".")
      return `  - ${p ? p + ": " : ""}${issue.message}`
    })
    .join("\n")
}

function getBranchName(worktree: string): string {
  try {
    const result = require("child_process").execSync(
      "git -C " + worktree + " rev-parse --abbrev-ref HEAD",
      { encoding: "utf-8" }
    )
    return result.trim()
  } catch {
    return "unknown-branch"
  }
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
}

function generateSpecPath(worktree: string, title: string): string {
  const branch = getBranchName(worktree)
  const timestamp = Math.floor(Date.now() / 1000)
  const titleSlug = slugify(title)
  const filename = `${branch}-${timestamp}-${titleSlug}.yaml`
  return path.join(worktree, SPEC_TREES_DIR, filename)
}

function getActiveSpecPointerPath(worktree: string): string {
  return path.join(specTreeDir(worktree), ACTIVE_SPEC_POINTER)
}

function getActiveSpecPath(worktree: string): string | null {
  const pointerPath = getActiveSpecPointerPath(worktree)
  if (!fs.existsSync(pointerPath)) {
    return null
  }
  return fs.readFileSync(pointerPath, "utf-8").trim()
}

function setActiveSpecPath(worktree: string, specPath: string): void {
  const dir = specTreeDir(worktree)
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true })
  }
  const pointerPath = getActiveSpecPointerPath(worktree)
  fs.writeFileSync(pointerPath, specPath)
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
  return path.join(worktree, SPEC_TREES_DIR)
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
  // Only return nodes explicitly marked as "leaf"
  if (node.node_type === "leaf") {
    return [node]
  }
  // If node is "branch" or "unexpanded" with children, recurse
  if (node.children.length > 0) {
    return node.children.flatMap(getLeafNodes)
  }
  // Node has no children but isn't marked as leaf - it's unexpanded, not a leaf
  return []
}

function topologicalSort(nodes: Node[], nodeMap: Map<string, Node>): Node[] {
  const visited = new Set<string>()
  const result: Node[] = []

  function visit(node: Node) {
    if (visited.has(node.id)) return
    visited.add(node.id)

    // Visit dependencies first
    if (node.depends_on) {
      for (const depId of node.depends_on) {
        const dep = nodeMap.get(depId)
        if (dep) visit(dep)
      }
    }

    result.push(node)
  }

  for (const node of nodes) {
    visit(node)
  }

  return result
}

function loadSpec(worktree: string): z.infer<typeof SpecTreeSpecSchema> {
  const specPath = getActiveSpecPath(worktree)

  if (!specPath) {
    throw new Error("No active spec-tree found. Use spec_tree_write to create one.")
  }

  if (!fs.existsSync(specPath)) {
    throw new Error(`Spec-tree file not found at ${specPath}. It may have been moved or deleted.`)
  }

  const content = fs.readFileSync(specPath, "utf-8")
  const parsed = parseYaml(content)
  if (!parsed) {
    throw new Error(`Invalid YAML in spec-tree file at ${specPath}`)
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
  const specPath = getActiveSpecPath(worktree)

  if (!specPath) {
    throw new Error("No active spec-tree found. Use spec_tree_write to create one.")
  }

  const result = SpecTreeSpecSchema.safeParse(spec)
  if (!result.success) {
    throw new Error(
      `Invalid spec after update: ${formatValidationErrors(result.error)}`
    )
  }

  // Ensure directory exists
  const dir = path.dirname(specPath)
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true })
  }

  fs.writeFileSync(specPath, toYamlString(spec))
}

function makeEmptyNode(id: string, title: string, description: string): Node {
  return {
    id,
    title,
    description,
    node_type: "unexpanded",
    depends_on: [],
    planning_status: "pending",
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
          "Initializes a uniquely-named YAML file in .opencode/spec-trees/ directory. " +
          "The filename includes the branch name, timestamp, and title slug. " +
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

          // Generate unique spec path with branch, timestamp, and title
          const specPath = generateSpecPath(worktree, args.title)

          // Ensure directory exists
          const dir = path.dirname(specPath)
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

          // Write the spec file
          fs.writeFileSync(specPath, toYamlString(spec))

          // Set this as the active spec for this worktree
          setActiveSpecPath(worktree, specPath)

          return `Spec-tree created successfully.

Title: ${args.title}
Root Node ID: ${args.root_node.id}
Root Node Title: ${args.root_node.title}
File: ${specPath}

Use spec_tree_register_node to add child nodes.`
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
            return `Node: ${node.id}
Title: ${node.title}
Description: ${node.description}
Type: ${node.node_type || "unexpanded"}
Planning status: ${node.planning_status || "pending"}
Depends on: ${node.depends_on?.length ? node.depends_on.join(", ") : "none"}
Children: ${node.children.length}
Test cases: ${node.tests.length}
File changes: ${node.file_changes.length}
Impl status: ${node.impl_status || "pending"}
Test status: ${node.test_status || "pending"}`
          }

          // Return full spec formatted
          let output = `Spec-Tree: ${spec.title}
Root: ${spec.root.id} - ${spec.root.title}

`
          const formatNode = (n: Node, indent: number = 0): string => {
            const prefix = "  ".repeat(indent)
            let s = `${prefix}${n.id}: ${n.title}\n`
            s += `${prefix}  Description: ${n.description}\n`
            if (n.children.length > 0) {
              s += `${prefix}  Children: ${n.children.length}\n`
              for (const child of n.children) {
                s += formatNode(child, indent + 1)
              }
            } else {
              s += `${prefix}  Status: impl=${n.impl_status || "pending"}, test=${n.test_status || "pending"}\n`
            }
            return s
          }

          return output + formatNode(spec.root)
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
            return `Node already registered.

Node ID: ${args.id}
Parent ID: ${args.parent_id}
Title: ${args.title}
Description: ${args.description}

Use spec_tree_update to modify this node.`
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

          return `Node registered successfully.

Node ID: ${args.id}
Parent ID: ${args.parent_id}
Title: ${args.title}
Description: ${args.description}

Use spec_tree_update to add children or modify status.`
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

          // Format updated fields
          const updatedFields = Object.keys(args.updates).join(", ")

          return `Node updated successfully.

Node ID: ${args.node_id}
Updated fields: ${updatedFields}

Use spec_tree_read to verify the changes.`
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

          return `Node: ${node.id}
Title: ${node.title}
Description: ${node.description}
Type: ${node.node_type || "unexpanded"}
Planning status: ${node.planning_status || "pending"}
Depends on: ${node.depends_on?.length ? node.depends_on.join(", ") : "none"}
Children: ${node.children.length}
Test cases: ${node.tests.length}
File changes: ${node.file_changes.length}
Impl status: ${node.impl_status || "pending"}
Test status: ${node.test_status || "pending"}`
        },
      }),

      spec_tree_get_leaves: tool({
        description:
          "Get all leaf nodes from the spec-tree spec in dependency order. " +
          "Returns array of leaf nodes sorted by dependencies (topological sort).",
        args: {},
        async execute(args, context) {
          const { worktree } = context
          const spec = loadSpec(worktree)
          const leaves = getLeafNodes(spec.root)

          if (leaves.length === 0) {
            return `No leaf nodes found. The spec-tree may only have a root node.`
          }

          // Build node map for dependency resolution
          const nodeMap = new Map<string, Node>()
          const addToMap = (node: Node) => {
            nodeMap.set(node.id, node)
            for (const child of node.children) {
              addToMap(child)
            }
          }
          addToMap(spec.root)

          // Sort by dependencies
          const sorted = topologicalSort(leaves, nodeMap)

          let output = `Leaf nodes found: ${sorted.length} (sorted by dependencies)\n\n`
          for (let i = 0; i < sorted.length; i++) {
            const leaf = sorted[i]
            output += `${i + 1}. ${leaf.id}: ${leaf.title}\n`
            output += `   Description: ${leaf.description}\n`
            output += `   Type: ${leaf.node_type || "unexpanded"}\n`
            output += `   Depends on: ${leaf.depends_on?.length ? leaf.depends_on.join(", ") : "none"}\n`
            output += `   Impl status: ${leaf.impl_status || "pending"}\n`
            output += `   Test status: ${leaf.test_status || "pending"}\n\n`
          }

          return output
        },
      }),

      spec_tree_render_ascii: tool({
        description:
          "Render the spec-tree as an ASCII tree visualization. " +
          "Optionally highlight a specific node (e.g., current leaf being reviewed). " +
          "Shows node types, status, and tree structure using tree characters.",
        args: {
          highlight_node_id: tool.schema
            .string()
            .optional()
            .describe("Optional node ID to highlight with asterisk (*)"),
        },
        async execute(args, context) {
          const { worktree } = context
          const spec = loadSpec(worktree)

          let output = `Spec-Tree: ${spec.title}\n\n`

          const formatNode = (
            node: Node,
            prefix: string = "",
            isLast: boolean = true,
            isHighlighted: boolean = false
          ): string => {
            const connector = isLast ? "└── " : "├── "
            const highlight = isHighlighted ? " *" : ""
            const nodeType = node.node_type || "unexpanded"
            const implStatus = node.impl_status || "pending"
            const testStatus = node.test_status || "pending"

            let line = `${prefix}${connector}${node.id}: ${node.title}${highlight}\n`
            line += `${prefix}${isLast ? "    " : "│   "}Type: ${nodeType}, Impl: ${implStatus}, Test: ${testStatus}\n`

            // Add children
            for (let i = 0; i < node.children.length; i++) {
              const child = node.children[i]
              const childIsLast = i === node.children.length - 1
              const childIsHighlighted = child.id === args.highlight_node_id
              line += formatNode(
                child,
                prefix + (isLast ? "    " : "│   "),
                childIsLast,
                childIsHighlighted
              )
            }

            return line
          }

          output += formatNode(spec.root, "", true, args.highlight_node_id === spec.root.id)

          if (args.highlight_node_id) {
            output += `\n* = Currently highlighted node\n`
          }

          return output
        },
      }),

      spec_tree_list: tool({
        description:
          "List all spec-tree files in the repository. " +
          "Shows the active spec for the current worktree (marked with *). " +
          "Use spec_tree_use to switch the active spec.",
        args: {},
        async execute(args, context) {
          const { worktree } = context
          const specTreesDir = specTreeDir(worktree)

          if (!fs.existsSync(specTreesDir)) {
            return "No spec-trees directory found. Use spec_tree_write to create one."
          }

          const files = fs.readdirSync(specTreesDir)
            .filter(f => f.endsWith(".yaml") && f !== ACTIVE_SPEC_POINTER)
            .sort()

          if (files.length === 0) {
            return "No spec-tree files found. Use spec_tree_write to create one."
          }

          const activeSpecPath = getActiveSpecPath(worktree)
          const activeFilename = activeSpecPath ? path.basename(activeSpecPath) : null

          let output = `Spec-tree files found: ${files.length}\n\n`

          for (const file of files) {
            const isActive = file === activeFilename
            const filePath = path.join(specTreesDir, file)
            const stats = fs.statSync(filePath)

            // Try to read the title from the file
            let title = "Unknown"
            try {
              const content = fs.readFileSync(filePath, "utf-8")
              const parsed = parseYaml(content)
              if (parsed && typeof parsed === 'object' && 'title' in parsed) {
                title = (parsed as any).title
              }
            } catch {
              // Ignore parse errors
            }

            const prefix = isActive ? "* " : "  "
            output += `${prefix}${file}\n`
            output += `    Title: ${title}\n`
            output += `    Created: ${stats.birthtime.toISOString()}\n\n`
          }

          output += "* = Active spec for this worktree\n"
          output += "\nUse spec_tree_use to switch the active spec."

          return output
        },
      }),

      spec_tree_use: tool({
        description:
          "Switch the active spec-tree for the current worktree. " +
          "Provide the filename (e.g., 'main-1777677963-my-title.yaml') " +
          "or the full path relative to the spec-trees directory.",
        args: {
          spec_file: tool.schema
            .string()
            .describe("Filename or path of the spec-tree YAML file to use as active"),
        },
        async execute(args, context) {
          const { worktree } = context
          const specTreesDir = specTreeDir(worktree)

          // Resolve the path
          let specPath: string
          if (path.isAbsolute(args.spec_file)) {
            specPath = args.spec_file
          } else if (args.spec_file.startsWith(SPEC_TREES_DIR)) {
            specPath = path.join(worktree, args.spec_file)
          } else {
            specPath = path.join(specTreesDir, args.spec_file)
          }

          // Validate the file exists and is a valid spec
          if (!fs.existsSync(specPath)) {
            throw new Error(`Spec file not found: ${specPath}`)
          }

          // Try to parse and validate
          const content = fs.readFileSync(specPath, "utf-8")
          const parsed = parseYaml(content)
          if (!parsed) {
            throw new Error(`Invalid YAML in spec file: ${specPath}`)
          }

          const specResult = SpecTreeSpecSchema.safeParse(parsed)
          if (!specResult.success) {
            throw new Error(
              `Invalid spec-tree spec: ${formatValidationErrors(specResult.error)}`
            )
          }

          // Set as active
          setActiveSpecPath(worktree, specPath)

          return `Active spec-tree switched successfully.

File: ${specPath}
Title: ${specResult.data.title}
Root Node: ${specResult.data.root.id} - ${specResult.data.root.title}

Use spec_tree_read to view the spec.`
        },
      }),
    },
  }
}