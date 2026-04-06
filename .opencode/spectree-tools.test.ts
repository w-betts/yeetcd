/**
 * Spectree Plugin Tests
 *
 * Unit tests for the spectree plugin.
 * Verifies the spectree-plugin.ts file exists, exports the plugin with required tools,
 * and is registered in opencode.json.
 */

import { describe, it, expect } from "bun:test"
import path from "path"
import fs from "fs"

const PLUGINS_DIR = path.join(process.cwd(), ".opencode", "plugins")
const SPECTREE_PLUGIN_PATH = path.join(PLUGINS_DIR, "spectree-plugin.ts")
const OPENCODE_JSON_PATH = path.join(process.cwd(), "opencode.json")

describe("Spectree plugin file exists and exports required tools", () => {
  it("spectree-plugin.ts file exists", () => {
    expect(fs.existsSync(SPECTREE_PLUGIN_PATH)).toBe(true)
  })

  it("exports SpectreePlugin", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("export const SpectreePlugin")
  })

  it("contains spectree_write tool", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("spectree_write")
  })

  it("contains spectree_read tool", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("spectree_read")
  })

  it("contains spectree_register_node tool", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("spectree_register_node")
  })

  it("contains spectree_update tool", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("spectree_update")
  })

  it("contains spectree_get_my_node tool", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("spectree_get_my_node")
  })

  it("contains spectree_get_leaves tool", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("spectree_get_leaves")
  })
})

describe("Spectree plugin contains schema definitions", () => {
  it("NodeSchema is defined", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("NodeSchema")
  })

  it("SpectreeSpecSchema is defined", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("SpectreeSpecSchema")
  })
})

describe("Spectree plugin is valid TypeScript", () => {
  it("file can be parsed without syntax errors", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content.length).toBeGreaterThan(0)
    expect(content).toContain("import")
    expect(content).toContain("@opencode-ai/plugin")
  })

  it("implements Plugin interface", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("Plugin")
    expect(content).toContain("async (ctx)")
  })
})

describe("Spectree plugin uses sessionID for identity", () => {
  it("spectree_write uses context.sessionID for root node ID", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    // Should destructure sessionID from context
    expect(content).toContain("sessionID")
  })

  it("spectree_register_node uses SDK client for parent lookup", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).toContain("client.session.get")
    expect(content).toContain("parentID")
  })

  it("spectree_update uses sessionID for ownership check", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    // spectree_update should use sessionID, not agent_id file
    expect(content).not.toContain("agent_id")
    expect(content).not.toContain("getAgentId")
  })

  it("does not use file-based agent_id mechanism", () => {
    const content = fs.readFileSync(SPECTREE_PLUGIN_PATH, "utf-8")
    expect(content).not.toContain(".opencode/agent_id")
    expect(content).not.toContain("getAgentId")
  })
})

describe("Spectree plugin is registered in opencode.json", () => {
  it("spectree-plugin.ts is listed in opencode.json plugins array", () => {
    const opencodeContent = fs.readFileSync(OPENCODE_JSON_PATH, "utf-8")
    const opencode = JSON.parse(opencodeContent)

    expect(opencode).toHaveProperty("plugin")
    expect(Array.isArray(opencode.plugin)).toBe(true)

    const hasSpectreePlugin = opencode.plugin.some((p: string) =>
      p.includes("spectree-plugin.ts")
    )
    expect(hasSpectreePlugin).toBe(true)
  })

  it("plugin path matches expected location", () => {
    const opencodeContent = fs.readFileSync(OPENCODE_JSON_PATH, "utf-8")
    const opencode = JSON.parse(opencodeContent)

    const spectreePlugin = opencode.plugin.find((p: string) =>
      p.includes("spectree-plugin.ts")
    )
    expect(spectreePlugin).toBe("./.opencode/plugins/spectree-plugin.ts")
  })
})
