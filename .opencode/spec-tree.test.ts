/**
 * Spec-Tree Prompt Tests
 *
 * Unit tests for the spec-tree primary agent prompts.
 * Verifies the prompt contains required workflow patterns.
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

  test('Mentions working directly (no planners)', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/directly|work directly/i);
    // Should NOT mention spawning planners (positive mention like "spawn planner subagents")
    // But should mention NOT spawning them
    expect(content).toMatch(/not.*spawn.*planner|do NOT.*spawn.*planner/i);
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

  test('References spec_tree_render_ascii tool', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/spec_tree_render_ascii/);
  });

  test('Mentions Phase 5.5 Pre-Implementation Review', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/Phase 5\.5|Pre-Implementation Review/);
  });

  test('Mentions surfacing ambiguities early', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/ambiguit|early|immediately/i);
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

  test('Mentions three-way breakdown choice after exploring node', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    // Should present user with three options: best split, other way, leaf
    expect(content).toMatch(/best split|agent.*split|suggested.*split/i);
    expect(content).toMatch(/break down.*other way|different.*split|alternative.*split/i);
    expect(content).toMatch(/leaf node|mark as leaf/i);
  });

  test('Specifies agent should identify best split before asking', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    // Agent should proactively identify the best way to split
    expect(content).toMatch(/identify.*split|propose.*split|best.*decomposition/i);
  });

  test('Does NOT have separate Phase 1 (root node treated like any other node)', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    // Phase1 should not exist as separate phase
    expect(content).not.toMatch(/### Phase 1:/);
  });

  test('Does NOT have separate Phase 2 (spec-tree created at init)', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    // Phase2 should not exist as separate phase
    expect(content).not.toMatch(/### Phase 2:/);
  });

  test('Does NOT have separate Phase 4 (recursive exploration in Phase 3)', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    // Phase4 should not exist as separate phase
    expect(content).not.toMatch(/### Phase 4:/);
  });

  test('Phase 3 is recursive (explores root and children recursively)', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/RECURSIVE|recursively/i);
    expect(content).toMatch(/back to step 1|back to Phase 3/i);
  });

  test('Spec-tree created during session initialization', () => {
    const content = readPrompt(ORCHESTRATOR_FILE);
    expect(content).not.toBeNull();
    expect(content).toMatch(/Create the spec-tree with root node/i);
    expect(content).toMatch(/session init/i);
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
