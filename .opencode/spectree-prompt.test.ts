/**
 * Spectree Prompt Tests
 *
 * Unit tests for the spectree primary agent prompt.
 * Verifies the prompt file exists and contains required content.
 */

import { test, expect, describe } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';

const SPECTREE_PROMPT_PATH = path.join(process.cwd(), '.opencode', 'prompts', 'spectree.md');

/**
 * Reads the spectree.md prompt file content.
 * Returns null if the file doesn't exist.
 */
function readSpectreePrompt(): string | null {
  try {
    return fs.readFileSync(SPECTREE_PROMPT_PATH, 'utf-8');
  } catch {
    return null;
  }
}

// ============================================================================
// Test Suite: Spectree Prompt File Existence
// ============================================================================

describe('Spectree Prompt File Existence', () => {
  test('Spectree prompt file exists', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
  });

  test('Spectree prompt file is not empty', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBe('');
    expect(content?.length).toBeGreaterThan(100);
  });
});

// ============================================================================
// Test Suite: Spectree Prompt Content - Orchestrator Style
// ============================================================================

describe('Spectree Prompt Content - Orchestrator Style', () => {
  test('Contains instructions for recursive problem decomposition', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    // Check for recursive/problem decomposition concepts
    expect(content).toMatch(/recursive/i);
    expect(content).toMatch(/problem/i);
    expect(content).toMatch(/decomp/i);
  });

  test('Contains orchestrator-style role definition', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    // Orchestrator agents typically delegate to subagents
    expect(content).toMatch(/subagent|delegat/i);
  });

  test('Is orchestrator-style (does NOT implement code directly)', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    // Should NOT contain implementation details like function definitions
    // Should focus on workflow and delegation
    const lowerContent = content.toLowerCase();
    expect(lowerContent).not.toMatch(/function\s+\w+\s*\(/);
    expect(lowerContent).not.toMatch(/class\s+\w+\s*{/);
  });
});

// ============================================================================
// Test Suite: Spectree Prompt Content - Breadth-First Exploration
// ============================================================================

describe('Spectree Prompt Content - Breadth-First Exploration', () => {
  test('Describes breadth-first problem exploration', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/breadth[- ]?first/i);
  });

  test('Describes level-based exploration (all branches at level N before N+1)', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    // Should mention levels or exploration order
    expect(content).toMatch(/level/i);
  });

  test('Mentions subagent spawning behavior', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/subagent/i);
  });
});

// ============================================================================
// Test Suite: Spectree Prompt Content - Tool References
// ============================================================================

describe('Spectree Prompt Content - Tool References', () => {
  test('References spectree_write tool', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/spectree_write/);
  });

  test('References spectree_read tool', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/spectree_read/);
  });

  test('References spectree_update tool', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/spectree_update/);
  });

  test('References spectree_get_my_node tool', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/spectree_get_my_node/);
  });

  test('References spectree_get_leaves tool', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/spectree_get_leaves/);
  });

  test('Contains spectree tool family references (plural)', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    // Should reference multiple spectree tools (tool family)
    const spectreeToolMatches = content.match(/spectree_\w+/g);
    expect(spectreeToolMatches).not.toBeNull();
    expect(spectreeToolMatches?.length).toBeGreaterThanOrEqual(3);
  });
});

// ============================================================================
// Test Suite: Spectree Prompt Content - Workflow Structure
// ============================================================================

describe('Spectree Prompt Content - Workflow Structure', () => {
  test('Contains workflow phases or steps', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    // Should describe phases or steps in the workflow
    expect(content).toMatch(/phase|step|workflow/i);
  });

  test('Mentions interaction logging', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    expect(content).toMatch(/interaction|log/i);
  });

  test('Mentions depth-first implementation order for leaves', () => {
    const content = readSpectreePrompt();
    expect(content).not.toBeNull();
    
    // Should mention depth-first or left-to-right order
    expect(content).toMatch(/depth[- ]?first|left[- ]?to[- ]?right/i);
  });
});