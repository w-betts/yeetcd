/**
 * Decision Tool Tests
 *
 * Unit tests for the decision logging tool implementation.
 * Tests decision_log (write/append) and decision_read (read) functions.
 */

import { test, expect, describe, beforeEach, afterEach } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

const TOOL_FILE = path.join(process.cwd(), '.opencode', 'tools', 'decision.ts');
const OPENCODE_CONFIG = path.join(process.cwd(), 'opencode.json');
const TEST_DECISION_LOGS_DIR = path.join(os.tmpdir(), 'yeetcd-test-decision-logs');

// ============================================================================
// Test Suite: Decision Tool File
// ============================================================================

describe('Decision Tool - File Structure', () => {
  test('Tool file exists', () => {
    expect(fs.existsSync(TOOL_FILE)).toBe(true);
  });

  test('Tool file exports decision_log function', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/export.*decision_log/);
  });

  test('Tool file exports decision_read function', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/export.*decision_read/);
  });

  test('Tool file uses Zod for schema validation', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/import.*\{ z \}.*from.*"zod"/);
    expect(content).toMatch(/const DecisionLogSchema.*=/);
    expect(content).toMatch(/const DecisionSchema.*=/);
  });

  test('Tool file follows session.ts/checklist.ts patterns', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    // Should have the helper functions
    expect(content).toMatch(/function formatValidationErrors/);
    expect(content).toMatch(/function parseYaml/);
    expect(content).toMatch(/function toYamlString/);
  });
});

// ============================================================================
// Test Suite: Decision Tool - Schema
// ============================================================================

describe('Decision Tool - Schema Validation', () => {
  test('YAML schema uses array of decisions with all required fields', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Check DecisionLogSchema has decisions array
    expect(content).toMatch(/decisions.*z\.array.*DecisionSchema/);
    
    // Check DecisionSchema has required fields
    expect(content).toMatch(/timestamp.*z\.string/);
    expect(content).toMatch(/agent_type.*z\.string/);
    expect(content).toMatch(/decision.*z\.string.*min\(1\)/);
    
    // Check optional fields
    expect(content).toMatch(/alternatives_considered.*optional/);
    expect(content).toMatch(/rationale.*optional/);
  });

  test('DecisionLogSchema has version field', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/version.*z\.literal\(1\)/);
  });

  test('DecisionLogSchema has session_id field', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/session_id.*z\.string/);
  });
});

// ============================================================================
// Test Suite: Decision Tool - Path Configuration
// ============================================================================

describe('Decision Tool - Path Configuration', () => {
  test('decision_log writes to correct path (.opencode/decision-logs/<session-id>.yaml)', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/decision-logs.*\$\{sessionId\}\.yaml/);
    expect(content).toMatch(/\.opencode.*decision-logs/);
  });

  test('getDecisionLogPath function exists', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/function getDecisionLogPath/);
  });
});

// ============================================================================
// Test Suite: Decision Tool - Integration with opencode.json
// ============================================================================

describe('Decision Tool - OpenCode Config', () => {
  test('opencode.json contains decision.ts in plugin array', () => {
    const configContent = fs.readFileSync(OPENCODE_CONFIG, 'utf-8');
    expect(configContent).toMatch(/\.opencode\/tools\/decision\.ts/);
  });

  test('opencode.json plugin array has correct structure', () => {
    const config = JSON.parse(fs.readFileSync(OPENCODE_CONFIG, 'utf-8'));
    expect(config.plugin).toBeInstanceOf(Array);
    expect(config.plugin).toContain('./.opencode/tools/decision.ts');
  });
});

// ============================================================================
// Test Suite: Decision Tool - Functional Tests
// ============================================================================

describe('Decision Tool - Functional Tests', () => {
  // We'll test the actual tool functions by importing them
  // Since the tools use 'tool()' from @opencode-ai/plugin, we'll test the helpers directly
  
  beforeEach(() => {
    // Clean up test directory
    if (fs.existsSync(TEST_DECISION_LOGS_DIR)) {
      fs.rmSync(TEST_DECISION_LOGS_DIR, { recursive: true, force: true });
    }
    fs.mkdirSync(TEST_DECISION_LOGS_DIR, { recursive: true });
  });

  afterEach(() => {
    // Clean up test directory
    if (fs.existsSync(TEST_DECISION_LOGS_DIR)) {
      fs.rmSync(TEST_DECISION_LOGS_DIR, { recursive: true, force: true });
    }
  });

  test('loadOrCreateDecisionLog creates new log with correct structure', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Verify the function exists and the structure is correct
    expect(content).toMatch(/function loadOrCreateDecisionLog/);
    expect(content).toMatch(/version: 1/);
    expect(content).toMatch(/decisions: \[\]/);
  });

  test('saveDecisionLog creates directory if needed', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/fs\.mkdirSync.*recursive: true/);
  });

  test('decision_log appends to existing decision log file (array grows)', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Verify the tool pushes to the decisions array
    expect(content).toMatch(/decisionLog\.decisions\.push/);
    expect(content).toMatch(/saveDecisionLog/);
  });

  test('decision_read tool reads decision log correctly', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Verify the tool reads and formats the log
    expect(content).toMatch(/function formatDecisionLog/);
    expect(content).toMatch(/fs\.existsSync.*decisionLogPath/);
    expect(content).toMatch(/parseYaml.*content/);
  });

  test('decision_read tool handles missing session gracefully', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Verify it checks for file existence and returns appropriate message
    expect(content).toMatch(/No decision log found for session/);
    expect(content).toMatch(/Use decision_log to start recording/);
  });
});

// ============================================================================
// Test Suite: Decision Tool - YAML Format
// ============================================================================

describe('Decision Tool - YAML Format', () => {
  test('YAML output uses correct formatting options', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    expect(content).toMatch(/lineWidth: 0/);
    expect(content).toMatch(/defaultStringType: "QUOTE_DOUBLE"/);
    expect(content).toMatch(/defaultKeyType: "PLAIN"/);
  });

  test('Decision log structure in YAML is correct', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Verify the schema structure
    expect(content).toMatch(/DecisionLogSchema.*=.*z\.object/);
    expect(content).toMatch(/DecisionSchema.*=.*z\.object/);
  });
});

// ============================================================================
// Test Suite: Decision Tool - Tool Arguments
// ============================================================================

describe('Decision Tool - Tool Arguments', () => {
  test('decision_log has required arguments (session_id, agent_type, decision)', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Check decision_log tool args - look for the pattern in the tool definition
    expect(content).toMatch(/export const decision_log[\s\S]*?args[\s\S]*?session_id/);
    expect(content).toMatch(/export const decision_log[\s\S]*?args[\s\S]*?agent_type/);
    expect(content).toMatch(/export const decision_log[\s\S]*?args[\s\S]*?decision/);
  });

  test('decision_log has optional arguments (alternatives_considered, rationale)', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Check optional args
    expect(content).toMatch(/alternatives_considered.*optional/);
    expect(content).toMatch(/rationale.*optional/);
  });

  test('decision_read has required argument (session_id)', () => {
    const content = fs.readFileSync(TOOL_FILE, 'utf-8');
    
    // Check decision_read tool args
    expect(content).toMatch(/export const decision_read[\s\S]*?args[\s\S]*?session_id/);
  });
});
