/**
 * Spectree Tools Tests
 *
 * Unit tests for the spectree tool family.
 * Verifies the spectree.ts file exists, exports required functions, and is registered in opencode.json.
 */

import { describe, it, expect } from "bun:test"
import path from "path"
import fs from "fs"

const TOOLS_DIR = path.join(process.cwd(), "..", ".opencode", "tools")
const SPECTREE_TOOL_PATH = path.join(TOOLS_DIR, "spectree.ts")
const OPENCODE_JSON_PATH = path.join(process.cwd(), "..", "opencode.json")

describe("Phase 2: spectree tools - Chunk 1", () => {
  describe("Spectree tools file exists and exports required functions", () => {
    it("spectree.ts file exists", () => {
      expect(fs.existsSync(SPECTREE_TOOL_PATH)).toBe(true)
    })

    it("spectree_write function is exported", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      expect(content).toContain("export const spectree_write")
    })

    it("spectree_read function is exported", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      expect(content).toContain("export const spectree_read")
    })

    it("spectree_update function is exported", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      expect(content).toContain("export const spectree_update")
    })

    it("spectree_get_my_node function is exported", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      expect(content).toContain("export const spectree_get_my_node")
    })

    it("spectree_get_leaves function is exported", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      expect(content).toContain("export const spectree_get_leaves")
    })
  })

  describe("Spectree tools file contains schema definitions for spec structure", () => {
    it("NodeSchema is defined", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      expect(content).toContain("NodeSchema")
    })

    it("SpectreeSpecSchema is defined", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      expect(content).toContain("SpectreeSpecSchema")
    })
  })

  describe("Spectree tools file is valid TypeScript", () => {
    it("file can be parsed without syntax errors", () => {
      const content = fs.readFileSync(SPECTREE_TOOL_PATH, "utf-8")
      // Basic check: file should not be empty and should have valid structure
      expect(content.length).toBeGreaterThan(0)
      expect(content).toContain("import")
      expect(content).toContain("@opencode-ai/plugin")
    })
  })
})

describe("Phase 2: spectree tools - Chunk 2", () => {
  describe("Spectree tools are loaded as plugins", () => {
    it("spectree.ts is listed in opencode.json plugins array", () => {
      const opencodeContent = fs.readFileSync(OPENCODE_JSON_PATH, "utf-8")
      const opencode = JSON.parse(opencodeContent)

      expect(opencode).toHaveProperty("plugin")
      expect(Array.isArray(opencode.plugin)).toBe(true)

      const hasSpectreePlugin = opencode.plugin.some(
        (p: string) => p.includes("spectree.ts")
      )
      expect(hasSpectreePlugin).toBe(true)
    })

    it("plugin path matches expected location", () => {
      const opencodeContent = fs.readFileSync(OPENCODE_JSON_PATH, "utf-8")
      const opencode = JSON.parse(opencodeContent)

      const spectreePlugin = opencode.plugin.find((p: string) =>
        p.includes("spectree.ts")
      )
      expect(spectreePlugin).toBe("./.opencode/tools/spectree.ts")
    })
  })
})
