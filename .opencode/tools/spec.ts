import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

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
  contracts: z.array(z.string()).describe("Interfaces/classes/APIs under test (e.g., 'PipelinePvcManager.createPvc()', 'KubernetesExecutionEngine.runJob()')"),
  given_when_then: z.string().describe("Pseudo test structure: GIVEN initial state/context, WHEN action is performed, THEN expected outcome"),
})

const FileChangeSchema = z.object({
  path: z.string().describe("File path relative to project root"),
  action: z.enum(["create", "modify", "delete"]).describe("What action to take"),
  description: z.string().describe("What this file change accomplishes"),
  is_test: z.boolean().describe("Whether this is a test file"),
})

const ChunkTestCaseSchema = TestCaseSchema // Alias for clarity when in chunk context

const ChunkSchema = z.object({
  name: z.string().describe("Chunk name (e.g., 'Chunk 1: Core parser logic')"),
  description: z.string().describe("What this chunk accomplishes"),
  status: z.enum(["pending", "in_progress", "completed"]).describe("Chunk status"),
  file_changes: z.array(FileChangeSchema).describe("File changes for this chunk"),
  test_cases: z.array(ChunkTestCaseSchema).describe("Test cases for this chunk"),
})

const PhaseSchema = z.object({
  name: z.string().describe("Phase name (e.g. 'Phase 1: Parser Implementation')"),
  description: z.string().describe("What this phase accomplishes"),
  status: z.enum(["pending", "in_progress", "completed", "released"]).describe("Phase status"),
  is_release_boundary: z.boolean().describe("Whether this phase marks a release boundary"),
  file_changes: z.array(FileChangeSchema).describe("File changes for this phase (filled by planner) - deprecated: use chunks instead"),
  test_cases: z.array(TestCaseSchema).describe("Test cases for this phase (filled by planner) - deprecated: use chunks instead"),
  chunks: z.array(ChunkSchema).describe("Implementation chunks within this phase - each chunk is independently verifiable"),
})

const AddressedIssueSchema = z.object({
  issue: z.string().describe("Description of the issue that was raised"),
  resolution: z.enum(["fixed", "ignored", "deferred", "clarified"]).describe("How the issue was resolved"),
  resolution_note: z.string().optional().describe("Optional note about how it was resolved or why it was ignored"),
  timestamp: z.string().describe("When this issue was addressed"),
})

const ReviewSchema = z.object({
  status: z.enum(["pending", "passed", "failed"]).describe("Review status"),
  reviewer: z.string().describe("Reviewer agent identifier"),
  timestamp: z.string().describe("ISO timestamp of review"),
  feedback: z.string().optional().describe("Review feedback (required if failed)"),
})

const SpecSchema = z.object({
  version: z.literal(2).describe("Spec schema version"),
  problem_statement: z.string().min(1).describe("Clear description of the problem being solved"),
  goals: z.array(z.string().min(1)).min(1).describe("List of goals this spec achieves"),
  constraints: z.array(z.string().min(1)).describe("Constraints and limitations to respect"),
  tech_choices: z.array(TechChoiceSchema).min(1).describe("Technology choices with rationale"),
  architecture: z.object({
    description: z.string().min(1).describe("High-level architecture description"),
    components: z.array(ComponentSchema).min(1).describe("System components"),
  }),
  test_strategy: z.object({
    approach: z.string().min(1).describe("Overall testing approach"),
    test_patterns: z.array(TestPatternSchema).min(1).describe("Test file patterns per language"),
  }),
  phases: z.array(PhaseSchema).min(1).describe("Implementation phases"),
  review: ReviewSchema.optional().describe("Adversarial review of the spec"),
  addressed_issues: z.array(AddressedIssueSchema).optional().describe("Issues raised in reviews and how they were resolved"),
  status: z.enum(["draft", "planned", "reviewed", "approved", "in_progress", "completed"]).describe("Spec status"),
})

type Spec = z.infer<typeof SpecSchema>

// --- Helpers ---

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60)
}

function specsDir(worktree: string): string {
  // Get the current branch name
  const branch = getCurrentBranch(worktree)
  // Use branch name as subdirectory (e.g., work/spec-feature-auth)
  const branchDir = branch || "unknown"
  return path.join(worktree, ".opencode", "specs", branchDir)
}

function getCurrentBranch(worktree: string): string | null {
  try {
    // Read the HEAD file from the worktree's .git directory
    // For worktrees, .git is a file containing the path to the git directory
    const dotGitPath = path.join(worktree, ".git")
    
    if (fs.existsSync(dotGitPath)) {
      const dotGitStat = fs.statSync(dotGitPath)
      
      if (dotGitStat.isFile()) {
        // This is a worktree - .git file contains path to gitdir
        const gitdirContent = fs.readFileSync(dotGitPath, "utf-8").trim()
        const gitdirMatch = gitdirContent.match(/^gitdir: (.+)$/)
        if (gitdirMatch) {
          const gitdir = gitdirMatch[1]
          // Read HEAD from the git directory
          const headPath = path.join(gitdir, "HEAD")
          if (fs.existsSync(headPath)) {
            const headContent = fs.readFileSync(headPath, "utf-8").trim()
            // HEAD can be either a branch ref or a commit hash
            const branchMatch = headContent.match(/^ref: refs\/heads\/(.+)$/)
            if (branchMatch) {
              return branchMatch[1]
            }
          }
        }
      } else if (dotGitStat.isDirectory()) {
        // This is the main repo - read HEAD directly
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
  } catch (error) {
    // If we can't determine the branch, return null
  }
  return null
}

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
    // Fallback to simple parsing if yaml library fails
    return null
  }
}

function toYamlString(obj: unknown): string {
  return yaml.stringify(obj, { lineWidth: 0, defaultStringType: "QUOTE_DOUBLE", defaultKeyType: "PLAIN" })
}

// --- Tools ---

export const spec_write = tool({
  description:
    "Write a development spec to a YAML file. Validates all mandatory fields against the spec schema. " +
    "The spec file is saved to .opencode/specs/<timestamp>-<slug>.yaml. " +
    "All fields are mandatory. Returns the file path on success, or validation errors on failure.",
  args: {
    title: tool.schema
      .string()
      .describe("Short descriptive title for the spec (used in filename, e.g. 'csv-parser' or 'auth-system')"),
    spec: tool.schema.object({
      problem_statement: tool.schema.string().describe("Clear description of the problem being solved"),
      goals: tool.schema.array(tool.schema.string()).describe("List of goals this spec achieves"),
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
      }),
      phases: tool.schema
        .array(
          tool.schema.object({
            name: tool.schema.string(),
            description: tool.schema.string(),
            status: tool.schema.enum(["pending", "in_progress", "completed", "released"]),
            is_release_boundary: tool.schema.boolean(),
            file_changes: tool.schema.array(
              tool.schema.object({
                path: tool.schema.string(),
                action: tool.schema.enum(["create", "modify", "delete"]),
                description: tool.schema.string(),
                is_test: tool.schema.boolean(),
              })
            ).optional(),
            test_cases: tool.schema.array(
              tool.schema.object({
                description: tool.schema.string(),
                type: tool.schema.enum(["unit", "integration", "e2e"]),
                target_component: tool.schema.string(),
                contracts: tool.schema.array(tool.schema.string()),
                given_when_then: tool.schema.string(),
              })
            ).optional(),
            chunks: tool.schema
              .array(
                tool.schema.object({
                  name: tool.schema.string(),
                  description: tool.schema.string(),
                  status: tool.schema.enum(["pending", "in_progress", "completed"]),
                  file_changes: tool.schema.array(
                    tool.schema.object({
                      path: tool.schema.string(),
                      action: tool.schema.enum(["create", "modify", "delete"]),
                      description: tool.schema.string(),
                      is_test: tool.schema.boolean(),
                    })
                  ),
                  test_cases: tool.schema.array(
                    tool.schema.object({
                      description: tool.schema.string(),
                      type: tool.schema.enum(["unit", "integration", "e2e"]),
                      target_component: tool.schema.string(),
                      contracts: tool.schema.array(tool.schema.string()),
                      given_when_then: tool.schema.string(),
                    })
                  ),
                })
              )
              .optional(),
          })
        )
        .describe("Implementation phases"),
      review: tool.schema
        .object({
          status: tool.schema.enum(["pending", "passed", "failed"]),
          reviewer: tool.schema.string(),
          timestamp: tool.schema.string(),
          feedback: tool.schema.string().optional(),
        })
        .optional()
        .describe("Adversarial review of the spec"),
      addressed_issues: tool.schema
        .array(
          tool.schema.object({
            issue: tool.schema.string(),
            resolution: tool.schema.enum(["fixed", "ignored", "deferred", "clarified"]),
            resolution_note: tool.schema.string().optional(),
            timestamp: tool.schema.string(),
          })
        )
        .optional()
        .describe("Issues raised in reviews and how they were resolved"),
      status: tool.schema
        .enum(["draft", "planned", "reviewed", "approved", "in_progress", "completed"])
        .describe("Spec status"),
    }),
  },
  async execute(args, context) {
    const fullSpec = { version: 2 as const, ...args.spec }

    // Validate
    const result = SpecSchema.safeParse(fullSpec)
    if (!result.success) {
      return `VALIDATION FAILED:\n${formatValidationErrors(result.error)}\n\nPlease fix the above issues and try again.`
    }

    // Generate filename
    const epoch = Math.floor(Date.now() / 1000)
    const slug = slugify(args.title)
    const filename = `${epoch}-${slug}.yaml`
    const dir = specsDir(context.worktree)

    // Ensure directory exists
    fs.mkdirSync(dir, { recursive: true })

    // Write YAML only
    const yamlPath = path.join(dir, filename)
    const header = `# Spec: ${args.title}\n# Generated: ${new Date().toISOString()}\n\n`
    const yamlContent = header + toYamlString(fullSpec)
    fs.writeFileSync(yamlPath, yamlContent)

    const phaseSummary = fullSpec.phases.map((p, i) => {
      const boundary = p.is_release_boundary ? " [RELEASE BOUNDARY]" : ""
      return `${i + 1}. ${p.name} (${p.status})${boundary}`
    }).join("\n")

    return `Spec written successfully.\n\nFile: ${yamlPath}\nStatus: ${fullSpec.status}\nTitle: ${args.title}\n\nThe spec contains:\n- ${fullSpec.goals.length} goals\n- ${fullSpec.constraints.length} constraints\n- ${fullSpec.tech_choices.length} tech choices\n- ${fullSpec.architecture.components.length} components\n- ${fullSpec.test_strategy.test_patterns.length} test patterns\n- ${fullSpec.phases.length} phases:\n${phaseSummary}`
  },
})

export const spec_read = tool({
  description:
    "Read a development spec. If no path is specified, reads the most recent spec file. " +
    "Returns the full spec content including all fields. " +
    "Use this to understand what to implement or test.",
  args: {
    path: tool.schema
      .string()
      .optional()
      .describe("Optional: specific spec file path. If omitted, reads the most recent spec."),
  },
  async execute(args, context) {
    const dir = specsDir(context.worktree)

    if (!fs.existsSync(dir)) {
      return "ERROR: No specs directory found. No specs have been created yet."
    }

    let yamlPath: string

    if (args.path) {
      yamlPath = args.path.endsWith(".yaml") ? args.path : `${args.path}.yaml`
      if (!path.isAbsolute(yamlPath)) {
        yamlPath = path.join(context.worktree, yamlPath)
      }
    } else {
      // Find most recent spec
      const files = fs.readdirSync(dir).filter((f) => f.endsWith(".yaml"))
      if (files.length === 0) {
        return "ERROR: No spec files found. Use spec_write to create a spec first."
      }
      // Sort by name (timestamp prefix ensures chronological order)
      files.sort()
      const latest = files[files.length - 1]
      yamlPath = path.join(dir, latest)
    }

    if (!fs.existsSync(yamlPath)) {
      return `ERROR: Spec file not found: ${yamlPath}`
    }

    const content = fs.readFileSync(yamlPath, "utf-8")
    let spec: Spec

    try {
      const parsed = parseYaml(content)
      if (!parsed) {
        return `ERROR: Failed to parse spec file: ${yamlPath}`
      }
      spec = parsed as Spec
    } catch {
      return `ERROR: Failed to parse spec file: ${yamlPath}`
    }

    // Validate
    const result = SpecSchema.safeParse(spec)
    if (!result.success) {
      return `WARNING: Spec file has validation issues:\n${formatValidationErrors(result.error)}\n\nRaw content:\n${content}`
    }

    // Format output
    let output = `SPEC: ${yamlPath}\n`
    output += `STATUS: ${spec.status}\n`
    output += `\n--- PROBLEM ---\n${spec.problem_statement}\n`
    output += `\n--- GOALS ---\n${spec.goals.map((g, i) => `${i + 1}. ${g}`).join("\n")}\n`
    output += `\n--- CONSTRAINTS ---\n${spec.constraints.map((c, i) => `${i + 1}. ${c}`).join("\n")}\n`
    output += `\n--- TECH CHOICES ---\n${spec.tech_choices.map((t) => `- ${t.area}: ${t.choice} (${t.rationale})`).join("\n")}\n`
    output += `\n--- ARCHITECTURE ---\n${spec.architecture.description}\n`
    output += `\nComponents:\n${spec.architecture.components.map((c) => `- ${c.name}: ${c.responsibility}\n  Interfaces: ${c.interfaces.join(", ")}`).join("\n")}\n`
    output += `\n--- TEST STRATEGY ---\n${spec.test_strategy.approach}\n`
    output += `\nTest patterns:\n${spec.test_strategy.test_patterns.map((p) => `- ${p.language}: ${p.pattern}`).join("\n")}\n`
    output += `\n--- PHASES ---\n`
    output += `(Note: Use 0-based index for spec_update, e.g., phase_index=0 for Phase 1)\n`
    spec.phases.forEach((phase, i) => {
      const boundary = phase.is_release_boundary ? " [RELEASE BOUNDARY]" : ""
      output += `\nPhase ${i + 1} (index ${i}): ${phase.name} (${phase.status})${boundary}\n`
      output += `   ${phase.description}\n`
      
      // Show chunks if present
      if (phase.chunks && phase.chunks.length > 0) {
        output += `   Chunks:\n`
        phase.chunks.forEach((chunk, ci) => {
          output += `     Chunk ${ci + 1} (index ${ci}): ${chunk.name} (${chunk.status})\n`
          output += `       ${chunk.description}\n`
          if (chunk.file_changes.length > 0) {
            output += `       File changes:\n${chunk.file_changes.map((f) => `         - ${f.action} ${f.path}${f.is_test ? " [TEST]" : ""}: ${f.description}`).join("\n")}\n`
          }
          if (chunk.test_cases.length > 0) {
            output += `       Test cases:\n${chunk.test_cases.map((t) => `         - [${t.type}] ${t.description} (targets: ${t.target_component})\n           Contracts: ${t.contracts.join(", ")}\n           ${t.given_when_then}`).join("\n")}\n`
          }
        })
      } else {
        // Fallback to deprecated file_changes and test_cases
        if (phase.file_changes && phase.file_changes.length > 0) {
          output += `   File changes:\n${phase.file_changes.map((f) => `     - ${f.action} ${f.path}${f.is_test ? " [TEST]" : ""}: ${f.description}`).join("\n")}\n`
        }
        if (phase.test_cases && phase.test_cases.length > 0) {
          output += `   Test cases:\n${phase.test_cases.map((t) => `     - [${t.type}] ${t.description} (targets: ${t.target_component})\n       Contracts: ${t.contracts.join(", ")}\n       ${t.given_when_then}`).join("\n")}\n`
        }
      }
    })
    if (spec.review) {
      output += `\n--- REVIEW ---\n`
      output += `Status: ${spec.review.status}\n`
      output += `Reviewer: ${spec.review.reviewer}\n`
      output += `Timestamp: ${spec.review.timestamp}\n`
      if (spec.review.feedback) {
        output += `Feedback: ${spec.review.feedback}\n`
      }
    }
    if (spec.addressed_issues && spec.addressed_issues.length > 0) {
      output += `\n--- ADDRESSED ISSUES ---\n`
      output += `Issues that were raised in reviews and resolved:\n`
      spec.addressed_issues.forEach((ai, i) => {
        output += `\n${i + 1}. [${ai.resolution.toUpperCase()}] ${ai.issue}\n`
        if (ai.resolution_note) {
          output += `   Note: ${ai.resolution_note}\n`
        }
        output += `   Resolved: ${ai.timestamp}\n`
      })
    }

    return output
  },
})

export const spec_update = tool({
  description:
    "Update portions of a development spec. Supports updating the overall status, " +
    "a specific phase's status, file_changes, or test_cases, or the review field. " +
    "Use this to mark phases as in_progress/completed, add low-level details from the planner, " +
    "or record review results from the reviewer.",
  args: {
    path: tool.schema
      .string()
      .optional()
      .describe("Optional: specific spec file path. If omitted, updates the most recent spec."),
    status: tool.schema
      .enum(["draft", "planned", "reviewed", "approved", "in_progress", "completed"])
      .optional()
      .describe("Update the overall spec status"),
    phase_index: tool.schema
      .number()
      .optional()
      .describe("Index of the phase to update (0-based: 0=Phase 1, 1=Phase 2, etc.)"),
    phase_status: tool.schema
      .enum(["pending", "in_progress", "completed", "released"])
      .optional()
      .describe("Update the phase status"),
    phase_file_changes: tool.schema
      .array(
        tool.schema.object({
          path: tool.schema.string(),
          action: tool.schema.enum(["create", "modify", "delete"]),
          description: tool.schema.string(),
          is_test: tool.schema.boolean(),
        })
      )
      .optional()
      .describe("Replace the phase's file changes"),
    phase_test_cases: tool.schema
      .array(
        tool.schema.object({
          description: tool.schema.string(),
          type: tool.schema.enum(["unit", "integration", "e2e"]),
          target_component: tool.schema.string(),
          contracts: tool.schema.array(tool.schema.string()),
          given_when_then: tool.schema.string(),
        })
      )
      .optional()
      .describe("Replace the phase's test cases (deprecated: use chunk_level updates instead)"),
    chunk_index: tool.schema
      .number()
      .optional()
      .describe("Index of the chunk to update within the phase (0-based)"),
    chunk_status: tool.schema
      .enum(["pending", "in_progress", "completed"])
      .optional()
      .describe("Update the chunk status"),
    chunk_file_changes: tool.schema
      .array(
        tool.schema.object({
          path: tool.schema.string(),
          action: tool.schema.enum(["create", "modify", "delete"]),
          description: tool.schema.string(),
          is_test: tool.schema.boolean(),
        })
      )
      .optional()
      .describe("Replace the chunk's file changes"),
    chunk_test_cases: tool.schema
      .array(
        tool.schema.object({
          description: tool.schema.string(),
          type: tool.schema.enum(["unit", "integration", "e2e"]),
          target_component: tool.schema.string(),
          contracts: tool.schema.array(tool.schema.string()),
          given_when_then: tool.schema.string(),
        })
      )
      .optional()
      .describe("Replace the chunk's test cases"),
    review_status: tool.schema
      .enum(["pending", "passed", "failed"])
      .optional()
      .describe("Update the review status (only reviewer agent should use this)"),
    review_feedback: tool.schema
      .string()
      .optional()
      .describe("Review feedback (required if review_status is 'failed')"),
    review_reviewer: tool.schema
      .string()
      .optional()
      .describe("Reviewer agent identifier"),
    add_addressed_issue: tool.schema
      .object({
        issue: tool.schema.string().describe("Description of the issue that was raised"),
        resolution: tool.schema.enum(["fixed", "ignored", "deferred", "clarified"]).describe("How the issue was resolved"),
        resolution_note: tool.schema.string().optional().describe("Optional note about how it was resolved or why it was ignored"),
      })
      .optional()
      .describe("Add an issue that was addressed (used by spec agent when user decides how to handle a review issue)"),
  },
  async execute(args, context) {
    const dir = specsDir(context.worktree)

    if (!fs.existsSync(dir)) {
      return "ERROR: No specs directory found. No specs have been created yet."
    }

    let yamlPath: string

    if (args.path) {
      yamlPath = args.path.endsWith(".yaml") ? args.path : `${args.path}.yaml`
      if (!path.isAbsolute(yamlPath)) {
        yamlPath = path.join(context.worktree, yamlPath)
      }
    } else {
      // Find most recent spec
      const files = fs.readdirSync(dir).filter((f) => f.endsWith(".yaml"))
      if (files.length === 0) {
        return "ERROR: No spec files found. Use spec_write to create a spec first."
      }
      files.sort()
      const latest = files[files.length - 1]
      yamlPath = path.join(dir, latest)
    }

    if (!fs.existsSync(yamlPath)) {
      return `ERROR: Spec file not found: ${yamlPath}`
    }

    const content = fs.readFileSync(yamlPath, "utf-8")
    let spec: Spec

    try {
      const parsed = parseYaml(content)
      if (!parsed) {
        return `ERROR: Failed to parse spec file: ${yamlPath}`
      }
      spec = parsed as Spec
    } catch {
      return `ERROR: Failed to parse spec file: ${yamlPath}`
    }

    // Apply updates
    const updates: string[] = []

    if (args.status !== undefined) {
      spec.status = args.status
      updates.push(`Overall status → ${args.status}`)
    }

    if (args.phase_index !== undefined) {
      if (args.phase_index < 0 || args.phase_index >= spec.phases.length) {
        return `ERROR: Invalid phase_index ${args.phase_index}. Valid range: 0-${spec.phases.length - 1}`
      }

      const phase = spec.phases[args.phase_index]

      if (args.phase_status !== undefined) {
        phase.status = args.phase_status
        updates.push(`Phase ${args.phase_index} ("${phase.name}") status → ${args.phase_status}`)
      }

      if (args.phase_file_changes !== undefined) {
        phase.file_changes = args.phase_file_changes
        updates.push(`Phase ${args.phase_index} ("${phase.name}") file_changes → ${args.phase_file_changes.length} changes`)
      }

      if (args.phase_test_cases !== undefined) {
        phase.test_cases = args.phase_test_cases
        updates.push(`Phase ${args.phase_index} ("${phase.name}") test_cases → ${args.phase_test_cases.length} cases`)
      }

      // Handle chunk-level updates
      if (args.chunk_index !== undefined) {
        if (!phase.chunks) {
          return `ERROR: Phase ${args.phase_index} has no chunks. Add chunks via spec_write or planner first.`
        }
        if (args.chunk_index < 0 || args.chunk_index >= phase.chunks.length) {
          return `ERROR: Invalid chunk_index ${args.chunk_index}. Valid range: 0-${phase.chunks.length - 1}`
        }
        const chunk = phase.chunks[args.chunk_index]
        if (args.chunk_status !== undefined) {
          chunk.status = args.chunk_status
          updates.push(`Phase ${args.phase_index} chunk ${args.chunk_index} ("${chunk.name}") status → ${args.chunk_status}`)
        }
        if (args.chunk_file_changes !== undefined) {
          chunk.file_changes = args.chunk_file_changes
          updates.push(`Phase ${args.phase_index} chunk ${args.chunk_index} ("${chunk.name}") file_changes → ${args.chunk_file_changes.length} changes`)
        }
        if (args.chunk_test_cases !== undefined) {
          chunk.test_cases = args.chunk_test_cases
          updates.push(`Phase ${args.phase_index} chunk ${args.chunk_index} ("${chunk.name}") test_cases → ${args.chunk_test_cases.length} cases`)
        }
      }
    }

    // Handle review updates
    if (args.review_status !== undefined) {
      if (args.review_status === "failed" && !args.review_feedback) {
        return "ERROR: review_feedback is required when review_status is 'failed'"
      }
      if (!args.review_reviewer) {
        return "ERROR: review_reviewer is required when updating review status"
      }
      spec.review = {
        status: args.review_status,
        reviewer: args.review_reviewer,
        timestamp: new Date().toISOString(),
        feedback: args.review_feedback,
      }
      updates.push(`Review status → ${args.review_status} (by ${args.review_reviewer})`)
    }

    // Handle addressed issue additions
    if (args.add_addressed_issue !== undefined) {
      if (!spec.addressed_issues) {
        spec.addressed_issues = []
      }
      spec.addressed_issues.push({
        issue: args.add_addressed_issue.issue,
        resolution: args.add_addressed_issue.resolution,
        resolution_note: args.add_addressed_issue.resolution_note,
        timestamp: new Date().toISOString(),
      })
      updates.push(`Addressed issue added: "${args.add_addressed_issue.issue.substring(0, 50)}..." → ${args.add_addressed_issue.resolution}`)
    }

    if (updates.length === 0) {
      return "ERROR: No updates specified. Provide at least one field to update."
    }

    // Validate updated spec
    const result = SpecSchema.safeParse(spec)
    if (!result.success) {
      return `VALIDATION FAILED after update:\n${formatValidationErrors(result.error)}\n\nUpdates were not applied.`
    }

    // Write back
    const header = `# Spec (updated)\n# Updated: ${new Date().toISOString()}\n\n`
    const yamlContent = header + toYamlString(spec)
    fs.writeFileSync(yamlPath, yamlContent)

    return `Spec updated successfully.\n\nFile: ${yamlPath}\nUpdates:\n${updates.map((u) => `- ${u}`).join("\n")}`
  },
})
