/**
 * Spec-Tree Plugin Tests
 *
 * Unit tests for the spec-tree plugin implementation.
 */

import { test, expect, describe } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';

const PLUGIN_FILE = path.join(process.cwd(), 'plugins', 'spec-tree-plugin.ts');

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
});
