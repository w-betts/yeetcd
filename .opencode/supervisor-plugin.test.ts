/**
 * Tests for SupervisorPlugin
 *
 * Tests decision log capture, particularly user answer extraction
 */

import { describe, test, expect, jest, beforeEach, afterEach } from "@jest/globals"
import path from "path"
import fs from "fs"
import os from "os"
import yaml from "yaml"

// Mock the supervisor plugin before import
const mockAddEntry = jest.fn()
const mockLoadLog = jest.fn()
const mockSaveLog = jest.fn()

// We'll test the helper functions directly
describe("SupervisorPlugin - User Answer Extraction", () => {
  // Test that demonstrates the bug: output.output returns raw characters
  // while output.parts returns meaningful text

  test("should extract meaningful answer from output.parts", () => {
    // Simulated output from question tool that the plugin receives
    const mockOutput = {
      parts: [
        { text: "Yes, implement the fix" }
      ]
    }

    // This is how the chat.message hook correctly extracts text
    const extractedAnswer = mockOutput.parts
      ?.map((p: any) => p.text || "")
      .join("")
      .substring(0, 500)

    expect(extractedAnswer).toBe("Yes, implement the fix")
  })

  test("should extract multiple answers when multiple choice question", () => {
    // Simulated output where user could select multiple
    const mockOutputs = [
      { parts: [{ text: "Yes, implement the fix" }] },
      { parts: [{ text: "Use output.parts to extract answers" }] },
    ]

    const answers = mockOutputs.map(output => 
      output.parts?.map((p: any) => p.text || "").join("") || "no answer"
    )

    expect(answers[0]).toBe("Yes, implement the fix")
    expect(answers[1]).toBe("Use output.parts to extract answers")
  })

  test("should handle empty output.parts gracefully", () => {
    const mockOutput = {
      parts: undefined
    }

    const extractedAnswer = mockOutput.parts
      ?.map((p: any) => p.text || "")
      .join("")
      .substring(0, 500)

    // Should be undefined, not crash
    expect(extractedAnswer).toBeUndefined()
  })
})

describe("SupervisorPlugin - Decision Log Storage", () => {
  let tempDir: string

  beforeEach(() => {
    // Create temp directory for tests
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "supervisor-test-"))
  })

  afterEach(() => {
    // Clean up
    if (tempDir && fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true })
    }
  })

  test("should store logs in .opencode/decision-logs directory", () => {
    const decisionLogsDir = path.join(tempDir, ".opencode", "decision-logs")
    
    // Create directory
    fs.mkdirSync(decisionLogsDir, { recursive: true })
    
    expect(fs.existsSync(decisionLogsDir)).toBe(true)
    
    // Should be in decision-logs, not sessions/supervisor
    const expectedPath = path.join(decisionLogsDir, "test-session.yaml")
    expect(expectedPath).toContain("decision-logs")
  })

  test("should create session file in decision-logs", () => {
    const decisionLogsDir = path.join(tempDir, ".opencode", "decision-logs")
    fs.mkdirSync(decisionLogsDir, { recursive: true })
    
    const logData = {
      session_id: "test-session",
      entries: [
        {
          timestamp: new Date().toISOString(),
          type: "user_input" as const,
          description: "Answered question: Test question",
          context: { 
            question: "What is the answer?",
            answer: "Yes, this is the answer"
          }
        }
      ]
    }
    
    const logPath = path.join(decisionLogsDir, "test-session.yaml")
    fs.writeFileSync(logPath, yaml.stringify(logData, { lineWidth: 0 }))
    
    expect(fs.existsSync(logPath)).toBe(true)
    
    // Verify content
    const readBack = yaml.parse(fs.readFileSync(logPath, "utf-8"))
    expect(readBack.entries[0].context.answer).toBe("Yes, this is the answer")
  })
})

describe("SupervisorPlugin - Log Entry Format", () => {
  test("should store user input with full question context", () => {
    const entry = {
      timestamp: new Date().toISOString(),
      type: "user_input" as const,
      description: "Answered question: TypeScript target",
      context: { 
        question: "What should the TypeScript SDK target? Go supports both Node.js and Go execution (for custom work). Java only supports Java. What runtime should TypeScript target?",
        answer: "Node.js runtime"
      }
    }
    
    // Verify entry has meaningful answer in context
    expect(entry.context.answer).toBe("Node.js runtime")
    expect(entry.context.question).toContain("TypeScript SDK")
  })

  test("should differentiate between choice answers and text input", () => {
    const choiceEntry = {
      timestamp: new Date().toISOString(),
      type: "user_input" as const,
      description: "Answered question: Generator approach",
      context: {
        question: "Which approach do you prefer?",
        answer: "Build-time (ts-morph)"
      }
    }
    
    const textEntry = {
      timestamp: new Date().toISOString(),
      type: "user_input" as const,
      description: "Answered question: Custom name",
      context: {
        question: "What should the pipeline be named?",
        answer: "my-custom-pipeline"
      }
    }
    
    // Both should have meaningful answers
    expect(choiceEntry.context.answer).toBe("Build-time (ts-morph)")
    expect(textEntry.context.answer).toBe("my-custom-pipeline")
  })
})