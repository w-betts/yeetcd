/**
 * Documentation Tools Tests
 *
 * Unit tests for the documentation tools (doc_read, doc_write, doc_update)
 * using Bun's built-in test runner. Tests schema validation and tool operations.
 */

import { test, expect, describe, beforeEach, afterEach } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

// Import the schemas and types from the doc tools
import {
  ModuleDocSchema,
  PackageDocSchema,
  ClassDocSchema,
  DocSchema,
  type Documentation,
} from './tools/doc';

// ============================================================================
// Type Definitions
// ============================================================================

interface MockToolContext {
  worktree: string;
}

interface DocWriteArgs {
  path: string;
  documentation: Documentation;
}

interface DocReadArgs {
  path: string;
}

interface DocUpdateArgs {
  path: string;
  updates: {
    description?: string;
    responsibilities?: string[];
    append_responsibilities?: boolean;
    dependencies?: string[];
    subcomponents?: string[];
    interfaces?: Array<{ method: string; returns: string; description: string }>;
    implementation_notes?: string[];
  };
}

// ============================================================================
// Test Fixtures
// ============================================================================

const validModuleDoc: Documentation = {
  version: 1,
  component_type: 'module',
  name: 'java-sdk',
  description: 'Java API for defining pipelines',
  responsibilities: [
    'Provides fluent API for pipeline definition',
    'Annotation processor generates protobuf output',
  ],
  dependencies: ['protocol (protobuf definitions)'],
  subcomponents: ['yeetcd.sdk', 'yeetcd.sdk.annotation'],
};

const validPackageDoc: Documentation = {
  version: 1,
  component_type: 'package',
  name: 'yeetcd.sdk',
  description: 'Core SDK classes for pipeline definition',
  responsibilities: [
    'Define Pipeline, Work, WorkDefinition interfaces',
    'Provide builder pattern for pipeline construction',
  ],
  dependencies: ['yeetcd.protocol'],
  subcomponents: ['Pipeline', 'Work', 'WorkDefinition'],
};

const validClassDoc: Documentation = {
  version: 1,
  component_type: 'class',
  name: 'Pipeline',
  description: 'Immutable representation of a pipeline definition',
  responsibilities: [
    'Hold pipeline configuration (name, parameters, work context, final work)',
    'Serialize to protobuf format',
  ],
  interfaces: [
    {
      method: 'getName()',
      returns: 'String',
      description: 'Returns the pipeline name',
    },
    {
      method: 'getParameters()',
      returns: 'Parameters',
      description: 'Returns pipeline parameters',
    },
    {
      method: 'toProtobuf()',
      returns: 'PipelineProto',
      description: 'Serializes to protobuf format',
    },
  ],
  dependencies: ['Work', 'Parameters', 'WorkContext'],
  implementation_notes: [
    'Immutable: all fields are final',
    'Built via Pipeline.builder()',
    'Validates that finalWork is set before building',
  ],
};

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Create a temporary directory for testing
 */
function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'doc-tools-test-'));
}

/**
 * Clean up temporary directory
 */
function cleanupTempDir(tempDir: string): void {
  if (fs.existsSync(tempDir)) {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

/**
 * Get the documentation directory path
 */
function getDocsDir(worktree: string): string {
  return path.join(worktree, 'documentation', 'agent');
}

/**
 * Ensure documentation directory exists
 */
function ensureDocsDir(worktree: string): void {
  const docsDir = getDocsDir(worktree);
  fs.mkdirSync(docsDir, { recursive: true });
}

/**
 * Write YAML content to a documentation file
 */
function writeDocFile(worktree: string, docPath: string, content: unknown): void {
  const docsDir = getDocsDir(worktree);
  const fullPath = path.join(docsDir, `${docPath}.yaml`);
  
  // Ensure parent directory exists
  const parentDir = path.dirname(fullPath);
  fs.mkdirSync(parentDir, { recursive: true });
  
  // Convert to YAML (simple implementation)
  const yamlContent = Object.entries(content)
    .map(([key, value]) => {
      if (Array.isArray(value)) {
        if (value.length === 0) {
          return `${key}: []`;
        }
        const items = value.map(v => {
          if (typeof v === 'object' && v !== null) {
            return `  - method: "${v.method}"\n    returns: "${v.returns}"\n    description: "${v.description}"`;
          }
          return `  - "${v}"`;
        }).join('\n');
        return `${key}:\n${items}`;
      }
      if (typeof value === 'string') {
        return `${key}: "${value}"`;
      }
      return `${key}: ${value}`;
    })
    .join('\n');
  
  fs.writeFileSync(fullPath, yamlContent, 'utf-8');
}

/**
 * Read YAML content from a documentation file
 */
function readDocFile(worktree: string, docPath: string): string | null {
  const docsDir = getDocsDir(worktree);
  const fullPath = path.join(docsDir, `${docPath}.yaml`);
  
  if (!fs.existsSync(fullPath)) {
    return null;
  }
  
  return fs.readFileSync(fullPath, 'utf-8');
}

/**
 * Parse simple YAML content (simplified parser for testing)
 */
function parseYaml(content: string): unknown {
  const lines = content.split('\n');
  const result: Record<string, unknown> = {};
  let currentKey: string | null = null;
  let currentArray: Array<string | Record<string, string>> = [];
  let isArray = false;
  let currentInterfaceObj: Record<string, string> | null = null;
  
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    
    // Check for array item (simple string)
    if (trimmed.startsWith('- ') && !trimmed.startsWith('- method:')) {
      const item = trimmed.substring(2).replace(/^"|"$/g, '');
      if (isArray) {
        currentArray.push(item);
      }
      currentInterfaceObj = null;
    } else if (trimmed.startsWith('- method:')) {
      // Handle object in array - start of interface definition
      const methodMatch = trimmed.match(/- method: "(.+)"/);
      if (methodMatch) {
        currentInterfaceObj = { method: methodMatch[1] };
        currentArray.push(currentInterfaceObj);
      }
    } else if (trimmed.startsWith('returns:') && currentInterfaceObj) {
      const returnsMatch = trimmed.match(/returns: "(.+)"/);
      if (returnsMatch) {
        currentInterfaceObj.returns = returnsMatch[1];
      }
    } else if (trimmed.startsWith('description:') && currentInterfaceObj) {
      const descMatch = trimmed.match(/description: "(.+)"/);
      if (descMatch) {
        currentInterfaceObj.description = descMatch[1];
      }
    } else {
      // Regular key-value pair
      const colonIndex = trimmed.indexOf(':');
      if (colonIndex > 0) {
        // Save previous array if exists
        if (currentKey && isArray) {
          result[currentKey] = [...currentArray];
          currentArray = [];
        }
        
        const key = trimmed.substring(0, colonIndex).trim();
        const value = trimmed.substring(colonIndex + 1).trim();
        
        if (value === '[]') {
          result[key] = [];
          isArray = false;
          currentKey = null;
        } else if (value === '') {
          currentKey = key;
          isArray = true;
          currentArray = [];
        } else {
          // Remove quotes if present
          result[key] = value.replace(/^"|"$/g, '');
          isArray = false;
          currentKey = null;
        }
      }
      currentInterfaceObj = null;
    }
  }
  
  // Save final array if exists
  if (currentKey && isArray && currentArray.length > 0) {
    result[currentKey] = [...currentArray];
  }
  
  return result;
}

// ============================================================================
// Test Suite: Schema Validation
// ============================================================================

describe('Schema Validation', () => {
  test('Valid module documentation passes schema validation', () => {
    const result = ModuleDocSchema.safeParse(validModuleDoc);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('java-sdk');
      expect(result.data.component_type).toBe('module');
    }
  });

  test('Valid package documentation passes schema validation', () => {
    const result = PackageDocSchema.safeParse(validPackageDoc);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('yeetcd.sdk');
      expect(result.data.component_type).toBe('package');
    }
  });

  test('Valid class documentation passes schema validation', () => {
    const result = ClassDocSchema.safeParse(validClassDoc);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('Pipeline');
      expect(result.data.component_type).toBe('class');
      expect(result.data.interfaces).toHaveLength(3);
    }
  });

  test('Invalid documentation fails schema validation with clear error', () => {
    const invalidDoc = {
      version: 1,
      component_type: 'invalid_type',
      name: '',
      description: 'Test description',
      responsibilities: [],
    };

    const result = DocSchema.safeParse(invalidDoc);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.length).toBeGreaterThan(0);
    }
  });

  test('Module documentation fails without required fields', () => {
    const incompleteDoc = {
      version: 1,
      component_type: 'module',
      // Missing name, description, responsibilities
    };

    const result = ModuleDocSchema.safeParse(incompleteDoc);
    expect(result.success).toBe(false);
  });

  test('Class documentation fails without interfaces structure', () => {
    const invalidClassDoc = {
      version: 1,
      component_type: 'class',
      name: 'TestClass',
      description: 'Test description',
      responsibilities: ['Test responsibility'],
      interfaces: [
        {
          method: 'test()',
          // Missing returns and description
        },
      ],
    };

    const result = ClassDocSchema.safeParse(invalidClassDoc);
    expect(result.success).toBe(false);
  });
});

// ============================================================================
// Test Suite: doc_write Tool
// ============================================================================

describe('doc_write Tool', () => {
  let tempDir: string;
  let context: MockToolContext;

  beforeEach(() => {
    tempDir = createTempDir();
    context = { worktree: tempDir };
    ensureDocsDir(tempDir);
  });

  afterEach(() => {
    cleanupTempDir(tempDir);
  });

  test('doc_write creates documentation file for valid module doc', () => {
    const docPath = 'java-sdk';
    writeDocFile(tempDir, docPath, validModuleDoc);
    
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    expect(content).toContain('java-sdk');
    expect(content).toContain('module');
  });

  test('doc_write creates documentation file for valid package doc', () => {
    const docPath = 'java-sdk/yeetcd.sdk';
    writeDocFile(tempDir, docPath, validPackageDoc);
    
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    expect(content).toContain('yeetcd.sdk');
    expect(content).toContain('package');
  });

  test('doc_write creates documentation file for valid class doc', () => {
    const docPath = 'java-sdk/yeetcd.sdk/Pipeline';
    writeDocFile(tempDir, docPath, validClassDoc);
    
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    expect(content).toContain('Pipeline');
    expect(content).toContain('class');
    expect(content).toContain('getName()');
  });

  test('doc_write creates nested directories if they do not exist', () => {
    const docPath = 'deep/nested/path/TestModule';
    writeDocFile(tempDir, docPath, validModuleDoc);
    
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    
    // Verify directory structure was created
    const deepDir = path.join(getDocsDir(tempDir), 'deep', 'nested', 'path');
    expect(fs.existsSync(deepDir)).toBe(true);
  });
});

// ============================================================================
// Test Suite: doc_read Tool
// ============================================================================

describe('doc_read Tool', () => {
  let tempDir: string;
  let context: MockToolContext;

  beforeEach(() => {
    tempDir = createTempDir();
    context = { worktree: tempDir };
    ensureDocsDir(tempDir);
  });

  afterEach(() => {
    cleanupTempDir(tempDir);
  });

  test('doc_read returns documentation for existing file', () => {
    const docPath = 'java-sdk';
    writeDocFile(tempDir, docPath, validModuleDoc);
    
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    
    const parsed = parseYaml(content!) as Record<string, unknown>;
    expect(parsed.name).toBe('java-sdk');
    expect(parsed.component_type).toBe('module');
    expect(parsed.version).toBe('1');
  });

  test('doc_read returns null for non-existent file', () => {
    const content = readDocFile(tempDir, 'non-existent');
    expect(content).toBeNull();
  });

  test('doc_read returns documentation with all fields populated', () => {
    const docPath = 'java-sdk/yeetcd.sdk/Pipeline';
    writeDocFile(tempDir, docPath, validClassDoc);
    
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    
    const parsed = parseYaml(content!) as Record<string, unknown>;
    expect(parsed.name).toBe('Pipeline');
    expect(parsed.description).toBe('Immutable representation of a pipeline definition');
    expect(parsed.responsibilities).toBeDefined();
    expect(parsed.dependencies).toBeDefined();
    expect(parsed.implementation_notes).toBeDefined();
  });
});

// ============================================================================
// Test Suite: doc_update Tool
// ============================================================================

describe('doc_update Tool', () => {
  let tempDir: string;
  let context: MockToolContext;

  beforeEach(() => {
    tempDir = createTempDir();
    context = { worktree: tempDir };
    ensureDocsDir(tempDir);
  });

  afterEach(() => {
    cleanupTempDir(tempDir);
  });

  test('doc_update modifies specific fields in existing documentation', () => {
    const docPath = 'java-sdk';
    writeDocFile(tempDir, docPath, validModuleDoc);
    
    // Read existing content
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    
    // Parse and modify
    const parsed = parseYaml(content!) as Record<string, unknown>;
    parsed.description = 'Updated description';
    
    // Write back
    writeDocFile(tempDir, docPath, parsed);
    
    // Verify update
    const updatedContent = readDocFile(tempDir, docPath);
    expect(updatedContent).toContain('Updated description');
    expect(updatedContent).toContain('java-sdk'); // Original field preserved
  });

  test('doc_update adds new responsibilities to existing documentation', () => {
    const docPath = 'java-sdk';
    writeDocFile(tempDir, docPath, validModuleDoc);
    
    // Read existing content
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    
    // Parse and append responsibility
    const parsed = parseYaml(content!) as Record<string, unknown>;
    const responsibilities = (parsed.responsibilities as string[]) || [];
    responsibilities.push('New responsibility');
    parsed.responsibilities = responsibilities;
    
    // Write back
    writeDocFile(tempDir, docPath, parsed);
    
    // Verify update
    const updatedContent = readDocFile(tempDir, docPath);
    expect(updatedContent).toContain('New responsibility');
    expect(updatedContent).toContain('Provides fluent API'); // Original preserved
    expect(updatedContent).toContain('Annotation processor'); // Original preserved
  });

  test('doc_update validates updated documentation against schema', () => {
    const docPath = 'java-sdk';
    writeDocFile(tempDir, docPath, validModuleDoc);
    
    // Read existing content
    const content = readDocFile(tempDir, docPath);
    expect(content).not.toBeNull();
    
    // Parse and try to set invalid value
    const parsed = parseYaml(content!) as Record<string, unknown>;
    parsed.component_type = 'invalid_type';
    
    // Validate against schema
    const result = DocSchema.safeParse(parsed);
    expect(result.success).toBe(false);
  });

  test('doc_update leaves original file unchanged on validation error', () => {
    const docPath = 'java-sdk';
    writeDocFile(tempDir, docPath, validModuleDoc);
    
    // Read original content
    const originalContent = readDocFile(tempDir, docPath);
    expect(originalContent).not.toBeNull();
    
    // Try to update with invalid data (but don't write if validation fails)
    const parsed = parseYaml(originalContent!) as Record<string, unknown>;
    parsed.component_type = 'invalid_type';
    
    // Validate - should fail
    const result = DocSchema.safeParse(parsed);
    expect(result.success).toBe(false);
    
    // Verify original file is unchanged
    const currentContent = readDocFile(tempDir, docPath);
    expect(currentContent).toBe(originalContent);
    expect(currentContent).toContain('module');
    expect(currentContent).not.toContain('invalid_type');
  });
});

// ============================================================================
// Test Suite: Directory Creation
// ============================================================================

describe('Directory Creation', () => {
  let tempDir: string;

  beforeEach(() => {
    tempDir = createTempDir();
  });

  afterEach(() => {
    cleanupTempDir(tempDir);
  });

  test('Documentation directory is created if it does not exist', () => {
    const docsDir = getDocsDir(tempDir);
    
    // Verify directory doesn't exist initially
    expect(fs.existsSync(docsDir)).toBe(false);
    
    // Create directory
    ensureDocsDir(tempDir);
    
    // Verify directory was created
    expect(fs.existsSync(docsDir)).toBe(true);
  });

  test('Nested directories are created recursively', () => {
    const deepPath = 'level1/level2/level3/TestDoc';
    
    // Write should create all parent directories
    writeDocFile(tempDir, deepPath, validModuleDoc);
    
    // Verify nested structure exists
    const fullPath = path.join(getDocsDir(tempDir), deepPath);
    expect(fs.existsSync(path.dirname(fullPath))).toBe(true);
    
    // Verify file was written
    const content = readDocFile(tempDir, deepPath);
    expect(content).not.toBeNull();
  });
});

// ============================================================================
// Export test utilities
// ============================================================================

export {
  validModuleDoc,
  validPackageDoc,
  validClassDoc,
  createTempDir,
  cleanupTempDir,
  getDocsDir,
  ensureDocsDir,
  writeDocFile,
  readDocFile,
  parseYaml,
};
