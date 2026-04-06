/**
 * Node Subagent Prompt Tests
 *
 * Unit tests for the node subagent prompt.
 * Verifies the prompt file exists and contains required content for handling sub-problems.
 */

import { test, expect, describe } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';

const NODE_PROMPT_PATH = path.join(process.cwd(), '.opencode', 'prompts', 'node.md');

/**
 * Reads the node.md prompt file content.
 * Returns null if the file doesn't exist.
 */
function readNodePrompt(): string | null {
  try {
    return fs.readFileSync(NODE_PROMPT_PATH, 'utf-8');
  } catch {
    return null;
  }
}

/**
 * Reads and parses opencode.json to check agent configuration.
 */
function readOpencodeConfig(): any | null {
  try {
    const configPath = path.join(process.cwd(), 'opencode.json');
    const content = fs.readFileSync(configPath, 'utf-8');
    return JSON.parse(content);
  } catch {
    return null;
  }
}

// ============================================================================
// Test Suite: Node Prompt File Existence (Chunk 1)
// ============================================================================

describe('Node Prompt File Existence', () => {
  test('Node prompt file exists', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
  });

  test('Node prompt file is not empty', () => {
    const content = readNodePrompt();
    expect(content).not.toBe('');
    expect(content?.length).toBeGreaterThan(100);
  });
});

// ============================================================================
// Test Suite: Node Prompt Content - Sub-Problem Handling (Chunk 1)
// ============================================================================

describe('Node Prompt Content - Sub-Problem Handling', () => {
  test('Contains instructions for handling sub-problems', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should reference sub-problem or node handling
    expect(content).toMatch(/sub[ -]?problem|node/i);
  });

  test('Contains behavior for subagent that handles individual nodes', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should describe subagent role
    expect(content).toMatch(/subagent/i);
  });

  test('Is subagent-style (does NOT implement code directly)', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should NOT contain implementation details like function definitions
    const lowerContent = content.toLowerCase();
    expect(lowerContent).not.toMatch(/function\s+\w+\s*\(/);
    expect(lowerContent).not.toMatch(/class\s+\w+\s*{/);
  });
});

// ============================================================================
// Test Suite: Node Prompt Content - Interaction with User (Chunk 1)
// ============================================================================

describe('Node Prompt Content - Interaction with User', () => {
  test('Contains user interaction references', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should reference user communication
    expect(content).toMatch(/user|question|interact/i);
  });

  test('Contains question tool usage', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should reference question tool for user interaction
    expect(content).toMatch(/question/i);
  });
});

// ============================================================================
// Test Suite: Node Prompt Content - Decision Options (Chunk 1)
// ============================================================================

describe('Node Prompt Content - Decision Options', () => {
  test('Contains breakdown decision options', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should reference breakdown or decompose options
    expect(content).toMatch(/breakdown|decomp|split/i);
  });

  test('Contains "deep enough" or leaf decision concept', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should reference "deep enough" concept for leaf nodes
    expect(content).toMatch(/deep enough|leaf|terminal/i);
  });

  test('Contains custom answer option', () => {
    const content = readNodePrompt();
    expect(content).not.toBeNull();
    
    // Should reference custom answer option
    expect(content).toMatch(/custom|other|answer/i);
  });
});

// ============================================================================
// Test Suite: Node Agent Configuration in opencode.json (Chunk 2)
// ============================================================================

describe('Node Agent Configuration in opencode.json', () => {
  test('Node agent exists in opencode.json', () => {
    const config = readOpencodeConfig();
    expect(config).not.toBeNull();
    expect(config.agent).toBeDefined();
    expect(config.agent.node).toBeDefined();
  });

  test('Node agent has mode set to subagent', () => {
    const config = readOpencodeConfig();
    expect(config).not.toBeNull();
    expect(config.agent.node).toBeDefined();
    expect(config.agent.node.mode).toBe('subagent');
  });

  test('Node agent references node.md prompt', () => {
    const config = readOpencodeConfig();
    expect(config).not.toBeNull();
    expect(config.agent.node).toBeDefined();
    expect(config.agent.node.prompt).toContain('node.md');
  });

  test('Node agent is spawnable (not hidden)', () => {
    const config = readOpencodeConfig();
    expect(config).not.toBeNull();
    expect(config.agent.node).toBeDefined();
    // hidden should be false or undefined (undefined means not hidden by default)
    expect(config.agent.node.hidden).not.toBe(true);
  });
});

// ============================================================================
// Test Suite: Spectree Agent Can Spawn Node Subagent (Chunk 2)
// ============================================================================

describe('Spectree Agent Can Spawn Node Subagent', () => {
  test('Spectree agent permissions include node subagent', () => {
    const config = readOpencodeConfig();
    expect(config).not.toBeNull();
    expect(config.agent.spectree).toBeDefined();
    expect(config.agent.spectree.permission).toBeDefined();
    
    // Check that spectree can spawn node
    const taskPerm = config.agent.spectree.permission.task;
    if (taskPerm && typeof taskPerm === 'object') {
      expect(taskPerm.node).toBe('allow');
    } else {
      // If task is "*": "allow", node should be allowed
      expect(taskPerm).toBe('allow');
    }
  });
});