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

  test('Mentions working directly (no planners)', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/directly|work directly/i);
    // Should NOT mention spawning planners
    expect(content).not.toMatch(/spawn.*planner|planner.*subagent/i);
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

  test('Mentions surfacing ambiguities early', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/ambiguit|early|immediately/i);
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

  test('Mentions batch questions for efficiency', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/batch|efficient/i);
  });

  test('References node_type field', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/node_type/i);
  });

  test('References depends_on field', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/depends_on|dependency/i);
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

  test('Plugin schema includes node_type field', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/node_type/);
    expect(content).toMatch(/unexpanded|branch|leaf/);
  });

  test('Plugin schema includes depends_on field', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/depends_on/);
  });

  test('Plugin schema includes planning_status field', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/planning_status/);
  });

  test('getLeafNodes only returns nodes with node_type="leaf"', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/node\.node_type === "leaf"/);
    // Should NOT return nodes just because children.length === 0
    expect(content).not.toMatch(/if \(node\.children\.length === 0\) \{\s*return \[node\]/);
  });

  test('spec_tree_get_leaves uses topological sort', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/topologicalSort/);
  });
});
