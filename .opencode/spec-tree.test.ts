/**
 * Spec-Tree Plugin Tests
 *
 * Unit tests for the spec-tree plugin implementation.
 */

import { test, expect, describe } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';

const PLUGIN_FILE = path.join(process.cwd(), 'plugins', 'spec-tree-plugin.ts');
const PROMPT_FILE = path.join(process.cwd(), 'prompts', 'spec-tree-orchestrator.md');

// ============================================================================
// Test Suite: Spec-Tree Plugin
// ============================================================================

describe('Spec-Tree Plugin', () => {
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

  test('Plugin defines spec_tree_list tool', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/spec_tree_list/);
  });

  test('Plugin defines spec_tree_use tool', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/spec_tree_use/);
  });

  test('Plugin uses .opencode/spec-trees directory', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/\.opencode\/spec-trees/);
  });

  test('Plugin has SPEC_TREES_DIR constant', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/SPEC_TREES_DIR/);
  });

  test('Plugin generates unique spec paths with branch and timestamp', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/generateSpecPath/);
    expect(content).toMatch(/getBranchName/);
    expect(content).toMatch(/timestamp/);
  });

  test('Plugin uses active spec pointer file', () => {
    const content = fs.readFileSync(PLUGIN_FILE, 'utf-8');
    expect(content).toMatch(/ACTIVE_SPEC_POINTER/);
    expect(content).toMatch(/getActiveSpecPath/);
    expect(content).toMatch(/setActiveSpecPath/);
  });
});

// ============================================================================
// Test Suite: Spec-Tree Orchestrator Prompt
// ============================================================================

describe('Spec-Tree Orchestrator Prompt', () => {
  let content: string;

  test('Prompt file exists', () => {
    expect(fs.existsSync(PROMPT_FILE)).toBe(true);
  });

  test('Has CONFIRM MUTUAL UNDERSTANDING step (step 5)', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/CONFIRM MUTUAL UNDERSTANDING.*MANDATORY.*NO EXCEPTIONS/);
  });

  test('Plays back mutual understanding summary', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/Play back the mutual understanding.*Summarize what was discussed/);
  });

  test('Asks confirmation question before breakdown decision', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/Are you happy with what we've agreed or do you want to discuss this node further before deciding whether to break it down\?/);
  });

  test('Repeats confirmation after further discussion', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/REPEAT this step.*play back updated understanding/);
  });

  test('Only proceeds to breakdown after user confirms satisfaction', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/Only proceed to step 6 when user explicitly confirms satisfaction/);
  });

  test('DECIDE DECOMPOSITION is step 6 (not step 5)', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/6\.\s*\*\*.*DECIDE DECOMPOSITION/);
  });

  test('Step 6 references steps 3-5 for context', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/Base your recommendation on the discussion in steps 3-5/);
  });

  test('Recursive workflow references steps 1-6', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/steps 1-6.*Explore.*Question.*Self-critique.*Confirm Understanding.*Decide Decomposition/);
  });

  test('Critical Rules includes confirm mutual understanding rule', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/CONFIRM MUTUAL UNDERSTANDING BEFORE BREAKDOWN/);
    expect(content).toMatch(/Are you happy with what we've agreed/);
  });

  test('Node Decomposition section includes confirm understanding step', () => {
    const content = fs.readFileSync(PROMPT_FILE, 'utf-8');
    expect(content).toMatch(/MUST confirm mutual understanding BEFORE asking about breakdown/);
  });
});
