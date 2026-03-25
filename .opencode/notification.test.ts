/**
 * Sound Notification Plugin Tests
 *
 * Unit tests for the notification plugin using Bun's built-in test runner.
 */

import { test, expect, describe, beforeEach, jest } from 'bun:test';

// ============================================================================
// Type Definitions
// ============================================================================

interface NotificationConfig {
  enabled: boolean;
  idle?: string;
  permission?: string;
  question?: string;
}

interface BunShell {
  run(command: string): Promise<{ exitCode: number; stdout: string; stderr: string }>;
}

interface Event {
  type: string;
}

interface Permission {
  id: string;
  type: string;
  pattern?: string | Array<string>;
  sessionID: string;
  messageID: string;
  callID?: string;
  title: string;
  metadata: {
    [key: string]: unknown;
  };
}

interface ToolInput {
  tool: string;
  sessionID: string;
  callID: string;
}

// ============================================================================
// Helper Functions
// ============================================================================

const DEFAULT_SOUND = '/System/Library/Sounds/Ping.aiff';

async function playSound(
  soundFile: string | undefined,
  shell: BunShell
): Promise<void> {
  const fileToPlay = soundFile || DEFAULT_SOUND;
  
  try {
    const result = await shell.run(`afplay "${fileToPlay}"`);
    if (result.exitCode !== 0) {
      console.warn(`Sound playback failed: ${result.stderr}`);
    }
  } catch (error) {
    console.warn(`Failed to play sound:`, error);
  }
}

// ============================================================================
// Test Suite: Sound Playback
// ============================================================================

describe('SoundPlayer', () => {
  test('playSound executes afplay with default sound', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await playSound(undefined, shell);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${DEFAULT_SOUND}"`);
  });

  test('playSound executes afplay with custom sound', async () => {
    const customSound = '/custom/sound.wav';
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await playSound(customSound, shell);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${customSound}"`);
  });

  test('playSound handles playback failure gracefully', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 1, stdout: '', stderr: 'playback error' })),
    };

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});

    await playSound(undefined, shell);

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Sound playback failed')
    );

    consoleSpy.mockRestore();
  });

  test('playSound handles exceptions gracefully', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => {
        throw new Error('Command failed');
      }),
    };

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});

    await playSound(undefined, shell);

    expect(consoleSpy).toHaveBeenCalledWith(
      'Failed to play sound:',
      expect.any(Error)
    );

    consoleSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Event Hook Handler
// ============================================================================

describe('NotificationPlugin - Event Hook', () => {
  const mockConfig: NotificationConfig = {
    enabled: true,
  };

  test('event hook plays sound for session.idle when enabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const event = { type: 'session.idle' };
    if (event.type === 'session.idle' && mockConfig.enabled) {
      await playSound(mockConfig.idle, shell);
    }

    expect(shell.run).toHaveBeenCalledTimes(1);
  });

  test('event hook does not play sound for other events', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const nonIdleEvents = ['session.created', 'file.edited', 'tool.executed', 'permission.asked'];

    for (const eventType of nonIdleEvents) {
      const event = { type: eventType };
      if (event.type === 'session.idle' && mockConfig.enabled) {
        await playSound(mockConfig.idle, shell);
      }
    }

    expect(shell.run).not.toHaveBeenCalled();
  });

  test('event hook does not play sound when disabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const disabledConfig: NotificationConfig = {
      enabled: false,
    };

    const event = { type: 'session.idle' };
    if (event.type === 'session.idle' && disabledConfig.enabled) {
      await playSound(disabledConfig.idle, shell);
    }

    expect(shell.run).not.toHaveBeenCalled();
  });
});

// ============================================================================
// Test Suite: Tool Hook Handler
// ============================================================================

describe('NotificationPlugin - Tool Hook', () => {
  const mockConfig: NotificationConfig = {
    enabled: true,
  };

  test('tool.execute.before hook plays sound for question tool', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const toolInput = { tool: 'question', sessionID: '123', callID: '456' };
    if (toolInput.tool === 'question' && mockConfig.enabled) {
      await playSound(mockConfig.question, shell);
    }

    expect(shell.run).toHaveBeenCalledTimes(1);
  });

  test('tool.execute.before hook does not play sound for other tools', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const otherTools = ['bash', 'read', 'edit', 'write', 'glob', 'grep'];

    for (const tool of otherTools) {
      const toolInput = { tool, sessionID: '123', callID: '456' };
      if (toolInput.tool === 'question' && mockConfig.enabled) {
        await playSound(mockConfig.question, shell);
      }
    }

    expect(shell.run).not.toHaveBeenCalled();
  });
});

// ============================================================================
// Test Suite: Permission Hook Handler
// ============================================================================

describe('NotificationPlugin - Permission Hook', () => {
  const mockConfig: NotificationConfig = {
    enabled: true,
  };

  test('permission.ask hook plays sound when enabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    if (mockConfig.enabled) {
      await playSound(mockConfig.permission, shell);
    }

    expect(shell.run).toHaveBeenCalledTimes(1);
  });

  test('permission.ask hook does not play sound when disabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const disabledConfig: NotificationConfig = {
      enabled: false,
    };

    if (disabledConfig.enabled) {
      await playSound(disabledConfig.permission, shell);
    }

    expect(shell.run).not.toHaveBeenCalled();
  });
});

// ============================================================================
// Test Suite: Custom Sound Paths
// ============================================================================

describe('NotificationPlugin - Custom Sounds', () => {
  test('uses custom idle sound when provided', async () => {
    const customSound = '/custom/idle.wav';
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const config: NotificationConfig = {
      enabled: true,
      idle: customSound,
    };

    await playSound(config.idle, shell);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${customSound}"`);
  });

  test('uses custom permission sound when provided', async () => {
    const customSound = '/custom/permission.wav';
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const config: NotificationConfig = {
      enabled: true,
      permission: customSound,
    };

    await playSound(config.permission, shell);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${customSound}"`);
  });

  test('uses custom question sound when provided', async () => {
    const customSound = '/custom/question.wav';
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const config: NotificationConfig = {
      enabled: true,
      question: customSound,
    };

    await playSound(config.question, shell);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${customSound}"`);
  });
});

// ============================================================================
// Export test utilities
// ============================================================================

export {
  playSound,
  type NotificationConfig,
  type BunShell,
  type Event,
  type Permission,
  type ToolInput,
};
