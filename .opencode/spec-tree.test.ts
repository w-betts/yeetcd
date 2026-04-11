/**
 * Spec-Tree Prompt Tests
 *
 * Unit tests for the spec-tree primary agent prompts.
 * Verifies the prompt files exist and contain required content.
 */

import { test, expect, describe } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';

const SPEC_TREE_PROMPTS_DIR = path.join(process.cwd(), 'prompts');

// ============================================================================
// Helper Functions
// ============================================================================

function readPrompt(filename: string): string | null {
  const promptPath = path.join(SPEC_TREE_PROMPTS_DIR, filename);
  try {
    return fs.readFileSync(promptPath, 'utf-8');
  } catch {
    return null;
  }
}

// ============================================================================
// Test Suite: Spec-Tree Orchestrator Prompt
// ============================================================================

describe('Spec-Tree Orchestrator Prompt', () => {
  const ORCHESTRATOR_FILE = 'spec-tree-orchestrator.md';

  test('Orchestrator prompt file exists', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
  });

  test('Orchestrator prompt is not empty', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBe('');
    expect(content?.length).toBeGreaterThan(100);
  });

  test('Describes orchestrator role', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/orchestrator/i);
  });

  test('Mentions breadth-first exploration', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/breadth[- ]?first/i);
  });

  test('Mentions question tool for user interaction', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/question tool/i);
  });

  test('References spec_tree_write tool', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/spec_tree_write/);
  });

  test('References spec_tree_register_node tool', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/spec_tree_register_node/);
  });

  test('References spec_tree_get_leaves tool', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/spec_tree_get_leaves/);
  });

  test('Mentions spawning spec-tree-planner subagent', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/spec-tree-planner|@spec-tree-planner/i);
  });

  test('Describes relay pattern for questions', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/relay/i);
  });

  test('Describes "What next?" question pattern', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/What next\?/i);
  });

  test('Does NOT implement code directly', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    const lowerContent = content.toLowerCase();
    expect(lowerContent).not.toMatch(/function\s+\w+\s*\(/);
    expect(lowerContent).not.toMatch(/class\s+\w+\s*{/);
  });
});

// ============================================================================
// Test Suite: Spec-Tree Planner Prompt
// ============================================================================

describe('Spec-Tree Planner Prompt', () => {
  const PLANNER_FILE = 'spec-tree-planner.md';

  test('Planner prompt file exists', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
  });

  test('Planner prompt is not empty', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBe('');
    expect(content?.length).toBeGreaterThan(100);
  });

  test('Describes planner role as subagent', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/planner/i);
  });

  test('Describes exploration process', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/explor/i);
  });

  test('Describes question accumulation', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/question/i);
  });

  test('Does NOT use question tool directly', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    // Should explicitly say NOT to use question tool
    expect(content).toMatch(/do.*not.*question.*tool|not.*use.*question/i);
  });

  test('Mentions returning questions to orchestrator', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/return.*orchestrator|orchestrator.*question/i);
  });

  test('References spec_tree_get_my_node tool', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/spec_tree_get_my_node/);
  });

  test('References spec_tree_update tool', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/spec_tree_update/);
  });

  test('Describes "What next?" options', () => {
    const content = readPrompt(PLANNER_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/What next\?/i);
    expect(content).toMatch(/break down|child nodes/i);
    expect(content).toMatch(/plan.*detail|leaf/);
    expect(content).toMatch(/explore more|Explore more/);
  });
});

// ============================================================================
// Test Suite: Tool Naming Convention
// ============================================================================

describe('Tool Naming Convention (spec_tree_*)', () => {
  test('Orchestrator uses snake_case spec_tree_ tools', () => {
    const content = readPrompt('spec-tree-orchestrator.md');
    expect(content).not.toBeNull();
    // Should use spec_tree_* not spectree_*
    expect(content).not.toMatch(/spectree_/);
    expect(content).toMatch(/spec_tree_/);
  });

  test('Planner uses snake_case spec_tree_ tools', () => {
    const content = readPrompt('spec-tree-planner.md');
    expect(content).not.toBeNull();
    // Should use spec_tree_* not spectree_*
    expect(content).not.toMatch(/spectree_/);
    expect(content).toMatch(/spec_tree_/);
  });
});

// ============================================================================
// Test Suite: Spec-Tree Plugin
// ============================================================================

describe('Spec-Tree Plugin', () => {
  const PLUGIN_FILE = path.join(process.cwd(), 'plugins', 'spec-tree-plugin.ts');

  test('Plugin file exists', () => {
    expect(fs.existsSync(PLUGIN_FILE)).toBe(true);
  });

  test('Plugin exports SpecTreePlugin', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/export.*SpecTreePlugin/);
  });

  test('Plugin defines spec_tree_* tools', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/spec_tree_write/);
    expect(content).toMatch(/spec_tree_read/);
    expect(content).toMatch(/spec_tree_register_node/);
    expect(content).toMatch(/spec_tree_update/);
    expect(content).toMatch(/spec_tree_get_my_node/);
    expect(content).toMatch(/spec_tree_get_leaves/);
  });
});