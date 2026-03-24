import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"

// --- Schema ---

const TechChoiceSchema = z.object({
  area: z.string().describe("Domain area (e.g. 'database', 'http-framework', 'testing')"),
  choice: z.string().describe("The specific technology chosen"),
  rationale: z.string().describe("Why this choice was made"),
})

const ComponentSchema = z.object({
  name: z.string().describe("Component name"),
  responsibility: z.string().describe("What this component does"),
  interfaces: z.array(z.string()).describe("Public interfaces/APIs this component exposes"),
})

const TestPatternSchema = z.object({
  language: z.string().describe("Programming language (e.g. 'go', 'typescript', 'python')"),
  pattern: z.string().describe("Glob pattern for test files (e.g. '*_test.go', '*.test.ts')"),
})

const TestCaseSchema = z.object({
  description: z.string().describe("What this test verifies"),
  type: z.enum(["unit", "integration", "e2e"]).describe("Test type"),
  target_component: z.string().describe("Which component this test targets"),
})

const FileChangeSchema = z.object({
  path: z.string().describe("File path relative to project root"),
  action: z.enum(["create", "modify", "delete"]).describe("What action to take"),
  description: z.string().describe("What this file change accomplishes"),
  is_test: z.boolean().describe("Whether this is a test file"),
})

const PlanSchema = z.object({
  version: z.literal(1).describe("Plan schema version"),
  problem_statement: z.string().min(1).describe("Clear description of the problem being solved"),
  goals: z.array(z.string().min(1)).min(1).describe("List of goals this plan achieves"),
  constraints: z.array(z.string().min(1)).describe("Constraints and limitations to respect"),
  tech_choices: z.array(TechChoiceSchema).min(1).describe("Technology choices with rationale"),
  architecture: z.object({
    description: z.string().min(1).describe("High-level architecture description"),
    components: z.array(ComponentSchema).min(1).describe("System components"),
  }),
  test_strategy: z.object({
    approach: z.string().min(1).describe("Overall testing approach"),
    test_patterns: z.array(TestPatternSchema).min(1).describe("Test file patterns per language"),
    test_cases: z.array(TestCaseSchema).min(1).describe("Specific test cases to implement"),
  }),
  file_changes: z.array(FileChangeSchema).min(1).describe("List of file changes to make"),
  status: z.enum(["draft", "approved"]).describe("Plan status - must be 'approved' before implementation begins"),
})

type Plan = z.infer<typeof PlanSchema>

// --- Helpers ---

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60)
}

function plansDir(worktree: string): string {
  return path.join(worktree, ".opencode", "plans")
}

function formatValidationErrors(error: z.ZodError): string {
  return error.issues
    .map((issue) => {
      const path = issue.path.join(".")
      return `  - ${path ? path + ": " : ""}${issue.message}`
    })
    .join("\n")
}

function toYaml(obj: unknown, indent = 0): string {
  const pad = "  ".repeat(indent)

  if (obj === null || obj === undefined) return `${pad}null`
  if (typeof obj === "boolean") return `${pad}${obj}`
  if (typeof obj === "number") return `${pad}${obj}`
  if (typeof obj === "string") {
    if (obj.includes("\n") || obj.includes(": ") || obj.includes("#") || obj.startsWith("- ")) {
      const lines = obj.split("\n")
      return `${pad}|\n${lines.map((l) => `${pad}  ${l}`).join("\n")}`
    }
    if (/[:{}\[\],&*?|<>=!%@`]/.test(obj) || obj === "" || obj === "true" || obj === "false") {
      return `${pad}"${obj.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`
    }
    return `${pad}${obj}`
  }

  if (Array.isArray(obj)) {
    if (obj.length === 0) return `${pad}[]`
    return obj
      .map((item) => {
        if (typeof item === "object" && item !== null && !Array.isArray(item)) {
          const entries = Object.entries(item)
          const first = entries[0]
          const rest = entries.slice(1)
          let result = `${pad}- ${first[0]}: ${toYaml(first[1], 0).trim()}`
          for (const [key, val] of rest) {
            result += `\n${pad}  ${key}: ${toYaml(val, 0).trim()}`
          }
          return result
        }
        return `${pad}- ${toYaml(item, 0).trim()}`
      })
      .join("\n")
  }

  if (typeof obj === "object") {
    const entries = Object.entries(obj as Record<string, unknown>)
    if (entries.length === 0) return `${pad}{}`
    return entries
      .map(([key, val]) => {
        if (typeof val === "object" && val !== null) {
          return `${pad}${key}:\n${toYaml(val, indent + 1)}`
        }
        return `${pad}${key}: ${toYaml(val, 0).trim()}`
      })
      .join("\n")
  }

  return `${pad}${String(obj)}`
}

// --- Tools ---

export const write = tool({
  description:
    "Write a development plan to a YAML file. Validates all mandatory fields against the plan schema. " +
    "The plan file is saved to .opencode/plans/<timestamp>-<slug>.yaml. " +
    "All fields are mandatory. Returns the file path on success, or validation errors on failure.",
  args: {
    title: tool.schema
      .string()
      .describe("Short descriptive title for the plan (used in filename, e.g. 'csv-parser' or 'auth-system')"),
    plan: tool.schema.object({
      problem_statement: tool.schema.string().describe("Clear description of the problem being solved"),
      goals: tool.schema.array(tool.schema.string()).describe("List of goals this plan achieves"),
      constraints: tool.schema.array(tool.schema.string()).describe("Constraints and limitations to respect"),
      tech_choices: tool.schema
        .array(
          tool.schema.object({
            area: tool.schema.string(),
            choice: tool.schema.string(),
            rationale: tool.schema.string(),
          })
        )
        .describe("Technology choices with rationale"),
      architecture: tool.schema.object({
        description: tool.schema.string().describe("High-level architecture description"),
        components: tool.schema
          .array(
            tool.schema.object({
              name: tool.schema.string(),
              responsibility: tool.schema.string(),
              interfaces: tool.schema.array(tool.schema.string()),
            })
          )
          .describe("System components"),
      }),
      test_strategy: tool.schema.object({
        approach: tool.schema.string().describe("Overall testing approach"),
        test_patterns: tool.schema
          .array(
            tool.schema.object({
              language: tool.schema.string(),
              pattern: tool.schema.string(),
            })
          )
          .describe("Test file patterns per language"),
        test_cases: tool.schema
          .array(
            tool.schema.object({
              description: tool.schema.string(),
              type: tool.schema.enum(["unit", "integration", "e2e"]),
              target_component: tool.schema.string(),
            })
          )
          .describe("Specific test cases to implement"),
      }),
      file_changes: tool.schema
        .array(
          tool.schema.object({
            path: tool.schema.string(),
            action: tool.schema.enum(["create", "modify", "delete"]),
            description: tool.schema.string(),
            is_test: tool.schema.boolean(),
          })
        )
        .describe("List of file changes to make"),
      status: tool.schema.enum(["draft", "approved"]).describe("Plan status"),
    }),
  },
  async execute(args, context) {
    const fullPlan = { version: 1 as const, ...args.plan }

    // Validate
    const result = PlanSchema.safeParse(fullPlan)
    if (!result.success) {
      return `VALIDATION FAILED:\n${formatValidationErrors(result.error)}\n\nPlease fix the above issues and try again.`
    }

    // Generate filename
    const epoch = Math.floor(Date.now() / 1000)
    const slug = slugify(args.title)
    const filename = `${epoch}-${slug}.yaml`
    const dir = plansDir(context.worktree)

    // Ensure directory exists
    fs.mkdirSync(dir, { recursive: true })

    // Write YAML for human readability
    const yamlPath = path.join(dir, filename)
    const yamlContent = `# Plan: ${args.title}\n# Generated: ${new Date().toISOString()}\n\n${toYaml(fullPlan)}\n`
    fs.writeFileSync(yamlPath, yamlContent)

    // Write JSON sidecar for reliable parsing
    const jsonPath = path.join(dir, filename.replace(".yaml", ".json"))
    fs.writeFileSync(jsonPath, JSON.stringify(fullPlan, null, 2))

    return `Plan written successfully.\n\nFile: ${yamlPath}\nStatus: ${fullPlan.status}\nTitle: ${args.title}\n\nThe plan contains:\n- ${fullPlan.goals.length} goals\n- ${fullPlan.constraints.length} constraints\n- ${fullPlan.tech_choices.length} tech choices\n- ${fullPlan.architecture.components.length} components\n- ${fullPlan.test_strategy.test_cases.length} test cases\n- ${fullPlan.test_strategy.test_patterns.length} test patterns\n- ${fullPlan.file_changes.length} file changes`
  },
})

export const read = tool({
  description:
    "Read a development plan. If no path is specified, reads the most recent plan file. " +
    "Returns the full plan content including all fields. " +
    "Use this to understand what to implement or test.",
  args: {
    path: tool.schema
      .string()
      .optional()
      .describe("Optional: specific plan file path. If omitted, reads the most recent plan."),
  },
  async execute(args, context) {
    const dir = plansDir(context.worktree)

    if (!fs.existsSync(dir)) {
      return "ERROR: No plans directory found. No plans have been created yet."
    }

    let jsonPath: string

    if (args.path) {
      // If they gave a .yaml path, find the .json sidecar
      jsonPath = args.path.endsWith(".yaml") ? args.path.replace(".yaml", ".json") : args.path
      if (!path.isAbsolute(jsonPath)) {
        jsonPath = path.join(context.worktree, jsonPath)
      }
    } else {
      // Find most recent plan
      const files = fs.readdirSync(dir).filter((f) => f.endsWith(".json"))
      if (files.length === 0) {
        return "ERROR: No plan files found. Use plan_write to create a plan first."
      }
      // Sort by name (timestamp prefix ensures chronological order)
      files.sort()
      const latest = files[files.length - 1]
      jsonPath = path.join(dir, latest)
    }

    if (!fs.existsSync(jsonPath)) {
      return `ERROR: Plan file not found: ${jsonPath}`
    }

    const content = fs.readFileSync(jsonPath, "utf-8")
    let plan: Plan

    try {
      plan = JSON.parse(content)
    } catch {
      return `ERROR: Failed to parse plan file: ${jsonPath}`
    }

    // Validate
    const result = PlanSchema.safeParse(plan)
    if (!result.success) {
      return `WARNING: Plan file has validation issues:\n${formatValidationErrors(result.error)}\n\nRaw content:\n${content}`
    }

    // Also return the yaml path for reference
    const yamlPath = jsonPath.replace(".json", ".yaml")

    // Format output
    let output = `PLAN: ${yamlPath}\n`
    output += `STATUS: ${plan.status}\n`
    output += `\n--- PROBLEM ---\n${plan.problem_statement}\n`
    output += `\n--- GOALS ---\n${plan.goals.map((g, i) => `${i + 1}. ${g}`).join("\n")}\n`
    output += `\n--- CONSTRAINTS ---\n${plan.constraints.map((c, i) => `${i + 1}. ${c}`).join("\n")}\n`
    output += `\n--- TECH CHOICES ---\n${plan.tech_choices.map((t) => `- ${t.area}: ${t.choice} (${t.rationale})`).join("\n")}\n`
    output += `\n--- ARCHITECTURE ---\n${plan.architecture.description}\n`
    output += `\nComponents:\n${plan.architecture.components.map((c) => `- ${c.name}: ${c.responsibility}\n  Interfaces: ${c.interfaces.join(", ")}`).join("\n")}\n`
    output += `\n--- TEST STRATEGY ---\n${plan.test_strategy.approach}\n`
    output += `\nTest patterns:\n${plan.test_strategy.test_patterns.map((p) => `- ${p.language}: ${p.pattern}`).join("\n")}\n`
    output += `\nTest cases:\n${plan.test_strategy.test_cases.map((t) => `- [${t.type}] ${t.description} (targets: ${t.target_component})`).join("\n")}\n`
    output += `\n--- FILE CHANGES ---\n${plan.file_changes.map((f) => `- ${f.action} ${f.path}${f.is_test ? " [TEST]" : ""}: ${f.description}`).join("\n")}\n`

    return output
  },
})
