import { tool } from "@opencode-ai/plugin"
import { z } from "zod"
import path from "path"
import fs from "fs"
import yaml from "yaml"

// --- Schema ---

const InterfaceSchema = z.object({
  method: z.string().describe("Method signature"),
  returns: z.string().describe("Return type"),
  description: z.string().describe("What this method does"),
})

const ModuleDocSchema = z.object({
  version: z.literal(1).describe("Documentation schema version"),
  component_type: z.literal("module").describe("Component type identifier"),
  name: z.string().min(1).describe("Module name"),
  description: z.string().min(1).describe("Module description"),
  responsibilities: z.array(z.string().min(1)).min(1).describe("List of responsibilities"),
  dependencies: z.array(z.string()).optional().describe("Dependencies on other modules/packages"),
  subcomponents: z.array(z.string()).optional().describe("Sub-packages or classes within this module"),
})

const PackageDocSchema = z.object({
  version: z.literal(1).describe("Documentation schema version"),
  component_type: z.literal("package").describe("Component type identifier"),
  name: z.string().min(1).describe("Package name"),
  description: z.string().min(1).describe("Package description"),
  responsibilities: z.array(z.string().min(1)).min(1).describe("List of responsibilities"),
  dependencies: z.array(z.string()).optional().describe("Dependencies on other packages/modules"),
  subcomponents: z.array(z.string()).optional().describe("Classes or sub-packages within this package"),
})

const ClassDocSchema = z.object({
  version: z.literal(1).describe("Documentation schema version"),
  component_type: z.literal("class").describe("Component type identifier"),
  name: z.string().min(1).describe("Class name"),
  description: z.string().min(1).describe("Class description"),
  responsibilities: z.array(z.string().min(1)).min(1).describe("List of responsibilities"),
  interfaces: z.array(InterfaceSchema).optional().describe("Public methods and their signatures"),
  dependencies: z.array(z.string()).optional().describe("Dependencies on other classes/packages"),
  implementation_notes: z.array(z.string()).optional().describe("Implementation details and notes"),
})

const DocSchema = z.union([ModuleDocSchema, PackageDocSchema, ClassDocSchema])

type Documentation = z.infer<typeof DocSchema>

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

function docsDir(worktree: string): string {
  return path.join(worktree, "documentation", "agent")
}

// --- Tools ---

export const doc_write = tool({
  description:
    "Write documentation to a YAML file. Validates against the documentation schema. " +
    "The documentation file is saved to documentation/agent/<path>.yaml. " +
    "Supports module, package, and class level documentation. " +
    "Returns success message or validation errors.",
  args: {
    path: tool.schema
      .string()
      .describe("Path relative to documentation/agent/ (e.g., 'java-sdk' or 'java-sdk/yeetcd.sdk/Pipeline')"),
    documentation: tool.schema.object({
      version: tool.schema.literal(1).describe("Schema version (must be 1)"),
      component_type: tool.schema.enum(["module", "package", "class"]).describe("Type of component"),
      name: tool.schema.string().describe("Component name"),
      description: tool.schema.string().describe("Component description"),
      responsibilities: tool.schema.array(tool.schema.string()).describe("List of responsibilities"),
      dependencies: tool.schema.array(tool.schema.string()).optional().describe("Dependencies"),
      subcomponents: tool.schema.array(tool.schema.string()).optional().describe("Subcomponents (for module/package)"),
      interfaces: tool.schema
        .array(
          tool.schema.object({
            method: tool.schema.string(),
            returns: tool.schema.string(),
            description: tool.schema.string(),
          })
        )
        .optional()
        .describe("Interfaces (for class only)"),
      implementation_notes: tool.schema.array(tool.schema.string()).optional().describe("Implementation notes (for class only)"),
    }),
  },
  async execute(args, context) {
    // Validate against schema
    const result = DocSchema.safeParse(args.documentation)
    if (!result.success) {
      return `VALIDATION FAILED:\n${formatValidationErrors(result.error)}\n\nPlease fix the above issues and try again.`
    }

    // Build full path
    const dir = docsDir(context.worktree)
    const yamlPath = path.join(dir, `${args.path}.yaml`)

    // Ensure directory exists
    fs.mkdirSync(path.dirname(yamlPath), { recursive: true })

    // Write YAML
    const header = `# Documentation: ${args.documentation.name}\n# Generated: ${new Date().toISOString()}\n\n`
    const yamlContent = header + toYamlString(args.documentation)
    fs.writeFileSync(yamlPath, yamlContent)

    const docType = args.documentation.component_type
    const hasInterfaces = args.documentation.interfaces && args.documentation.interfaces.length > 0
    const hasNotes = args.documentation.implementation_notes && args.documentation.implementation_notes.length > 0

    return `Documentation written successfully.\n\nFile: ${yamlPath}\nType: ${docType}\nName: ${args.documentation.name}\n\nThe documentation contains:\n- ${args.documentation.responsibilities.length} responsibilities${args.documentation.dependencies ? `\n- ${args.documentation.dependencies.length} dependencies` : ""}${args.documentation.subcomponents ? `\n- ${args.documentation.subcomponents.length} subcomponents` : ""}${hasInterfaces ? `\n- ${args.documentation.interfaces!.length} interfaces` : ""}${hasNotes ? `\n- ${args.documentation.implementation_notes!.length} implementation notes` : ""}`
  },
})

export const doc_read = tool({
  description:
    "Read documentation from a YAML file. Returns the parsed documentation object " +
    "or an error if the file doesn't exist or is invalid.",
  args: {
    path: tool.schema
      .string()
      .describe("Path relative to documentation/agent/ (e.g., 'java-sdk' or 'java-sdk/yeetcd.sdk/Pipeline')"),
  },
  async execute(args, context) {
    const dir = docsDir(context.worktree)
    const yamlPath = path.join(dir, `${args.path}.yaml`)

    if (!fs.existsSync(yamlPath)) {
      return `ERROR: Documentation file not found: ${yamlPath}`
    }

    const content = fs.readFileSync(yamlPath, "utf-8")
    let doc: Documentation

    try {
      const parsed = parseYaml(content)
      if (!parsed) {
        return `ERROR: Failed to parse documentation file: ${yamlPath}`
      }
      doc = parsed as Documentation
    } catch {
      return `ERROR: Failed to parse documentation file: ${yamlPath}`
    }

    // Validate
    const result = DocSchema.safeParse(doc)
    if (!result.success) {
      return `WARNING: Documentation file has validation issues:\n${formatValidationErrors(result.error)}\n\nRaw content:\n${content}`
    }

    // Format output
    let output = `DOCUMENTATION: ${yamlPath}\n`
    output += `Type: ${doc.component_type}\n`
    output += `Name: ${doc.name}\n`
    output += `\n--- DESCRIPTION ---\n${doc.description}\n`
    output += `\n--- RESPONSIBILITIES ---\n${doc.responsibilities.map((r, i) => `${i + 1}. ${r}`).join("\n")}\n`
    
    if (doc.dependencies && doc.dependencies.length > 0) {
      output += `\n--- DEPENDENCIES ---\n${doc.dependencies.map((d) => `- ${d}`).join("\n")}\n`
    }
    
    if (doc.subcomponents && doc.subcomponents.length > 0) {
      output += `\n--- SUBCOMPONENTS ---\n${doc.subcomponents.map((s) => `- ${s}`).join("\n")}\n`
    }
    
    if (doc.interfaces && doc.interfaces.length > 0) {
      output += `\n--- INTERFACES ---\n${doc.interfaces.map((iface) => `- ${iface.method} → ${iface.returns}\n  ${iface.description}`).join("\n")}\n`
    }
    
    if (doc.implementation_notes && doc.implementation_notes.length > 0) {
      output += `\n--- IMPLEMENTATION NOTES ---\n${doc.implementation_notes.map((n) => `- ${n}`).join("\n")}\n`
    }

    return output
  },
})

export const doc_update = tool({
  description:
    "Update specific fields in existing documentation. Validates the updated " +
    "documentation against the schema. Only updates the specified fields, " +
    "leaving other fields unchanged. Supports appending to arrays like responsibilities.",
  args: {
    path: tool.schema
      .string()
      .describe("Path relative to documentation/agent/ (e.g., 'java-sdk' or 'java-sdk/yeetcd.sdk/Pipeline')"),
    updates: tool.schema.object({
      description: tool.schema.string().optional().describe("New description"),
      responsibilities: tool.schema.array(tool.schema.string()).optional().describe("New responsibilities (replaces existing unless append=true)"),
      append_responsibilities: tool.schema.boolean().optional().describe("If true, append to existing responsibilities instead of replacing"),
      dependencies: tool.schema.array(tool.schema.string()).optional().describe("New dependencies (replaces existing)"),
      subcomponents: tool.schema.array(tool.schema.string()).optional().describe("New subcomponents (replaces existing)"),
      interfaces: tool.schema
        .array(
          tool.schema.object({
            method: tool.schema.string(),
            returns: tool.schema.string(),
            description: tool.schema.string(),
          })
        )
        .optional()
        .describe("New interfaces (replaces existing, for class only)"),
      implementation_notes: tool.schema.array(tool.schema.string()).optional().describe("New implementation notes (replaces existing, for class only)"),
    }),
  },
  async execute(args, context) {
    const dir = docsDir(context.worktree)
    const yamlPath = path.join(dir, `${args.path}.yaml`)

    if (!fs.existsSync(yamlPath)) {
      return `ERROR: Documentation file not found: ${yamlPath}`
    }

    const content = fs.readFileSync(yamlPath, "utf-8")
    let doc: Documentation

    try {
      const parsed = parseYaml(content)
      if (!parsed) {
        return `ERROR: Failed to parse documentation file: ${yamlPath}`
      }
      doc = parsed as Documentation
    } catch {
      return `ERROR: Failed to parse documentation file: ${yamlPath}`
    }

    // Apply updates
    const appliedUpdates: string[] = []

    if (args.updates.description !== undefined) {
      doc.description = args.updates.description
      appliedUpdates.push("description")
    }

    if (args.updates.responsibilities !== undefined) {
      if (args.updates.append_responsibilities && doc.responsibilities) {
        doc.responsibilities = [...doc.responsibilities, ...args.updates.responsibilities]
        appliedUpdates.push(`responsibilities (appended ${args.updates.responsibilities.length} items)`)
      } else {
        doc.responsibilities = args.updates.responsibilities
        appliedUpdates.push(`responsibilities (replaced with ${args.updates.responsibilities.length} items)`)
      }
    }

    if (args.updates.dependencies !== undefined) {
      doc.dependencies = args.updates.dependencies
      appliedUpdates.push(`dependencies (${args.updates.dependencies.length} items)`)
    }

    if (args.updates.subcomponents !== undefined) {
      doc.subcomponents = args.updates.subcomponents
      appliedUpdates.push(`subcomponents (${args.updates.subcomponents.length} items)`)
    }

    if (args.updates.interfaces !== undefined) {
      doc.interfaces = args.updates.interfaces
      appliedUpdates.push(`interfaces (${args.updates.interfaces.length} items)`)
    }

    if (args.updates.implementation_notes !== undefined) {
      doc.implementation_notes = args.updates.implementation_notes
      appliedUpdates.push(`implementation_notes (${args.updates.implementation_notes.length} items)`)
    }

    if (appliedUpdates.length === 0) {
      return "ERROR: No updates specified. Provide at least one field to update."
    }

    // Validate updated doc
    const result = DocSchema.safeParse(doc)
    if (!result.success) {
      return `VALIDATION FAILED after update:\n${formatValidationErrors(result.error)}\n\nUpdates were not applied.`
    }

    // Write back
    const header = `# Documentation (updated)\n# Updated: ${new Date().toISOString()}\n\n`
    const yamlContent = header + toYamlString(doc)
    fs.writeFileSync(yamlPath, yamlContent)

    return `Documentation updated successfully.\n\nFile: ${yamlPath}\nUpdates:\n${appliedUpdates.map((u) => `- ${u}`).join("\n")}`
  },
})

// Export schemas for testing
export { ModuleDocSchema, PackageDocSchema, ClassDocSchema, DocSchema, type Documentation }
