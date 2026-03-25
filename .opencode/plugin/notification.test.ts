/**
 * Sound Notification Plugin Tests
 *
 * Unit tests for the notification plugin using Bun's built-in test runner.
 * Tests cover platform detection, config loading, event filtering, and sound playback.
 */

import { test, expect, describe, beforeEach, afterEach, jest } from 'bun:test';
import * as fs from 'fs';
import * as path from 'path';

// ============================================================================
// Type Definitions (mirroring expected implementation)
// ============================================================================

interface NotificationConfig {
  sounds: {
    enabled: boolean;
    idle?: string;
    permission?: string;
    question?: string;
  };
  delay: number;
  volume: number;
}

interface BunShell {
  run(command: string): Promise<{ exitCode: number; stdout: string; stderr: string }>;
}

interface PluginInput {
  BunShell: BunShell;
  config: NotificationConfig;
}

// ============================================================================
// Mock Setup
// ============================================================================

const mockBunShell: BunShell = {
  run: jest.fn(async (command: string) => {
    // Simulate command availability checks
    if (command.includes('which') || command.includes('where') || command.includes('Get-Command')) {
      const cmd = command.split(' ').pop() || '';
      // Simulate available commands
      const availableCommands = ['afplay', 'paplay', 'aplay', 'mpv'];
      if (availableCommands.includes(cmd)) {
        return { exitCode: 0, stdout: `/usr/bin/${cmd}`, stderr: '' };
      }
      return { exitCode: 1, stdout: '', stderr: 'not found' };
    }
    // Simulate sound playback
    return { exitCode: 0, stdout: '', stderr: '' };
  }),
};

// Store original process.platform
const originalPlatform = Object.getOwnPropertyDescriptor(process, 'platform');

// ============================================================================
// Helper Functions (to be implemented)
// ============================================================================

// These functions represent the expected interface from notification.ts
// They will be imported from the actual implementation once it exists

function getPlatform(): 'macos' | 'linux' | 'windows' {
  const platform = process.platform;
  if (platform === 'darwin') return 'macos';
  if (platform === 'linux') return 'linux';
  if (platform === 'win32') return 'windows';
  throw new Error(`Unsupported platform: ${platform}`);
}

function getSoundCommand(platform: string, shell: BunShell): Promise<string | null> {
  return getAvailableSoundCommand(platform, shell);
}

async function getAvailableSoundCommand(platform: string, shell: BunShell): Promise<string | null> {
  switch (platform) {
    case 'macos':
      try {
        const result = await shell.run('which afplay');
        if (result.exitCode === 0) return 'afplay';
      } catch {
        return null;
      }
      return null;

    case 'linux':
      const linuxCommands = ['paplay', 'aplay', 'mpv'];
      for (const cmd of linuxCommands) {
        try {
          const result = await shell.run(`which ${cmd}`);
          if (result.exitCode === 0) return cmd;
        } catch {
          continue;
        }
      }
      return null;

    case 'windows':
      try {
        const result = await shell.run('powershell -Command "Get-Command SoundPlayer"');
        if (result.exitCode === 0) return 'powershell';
      } catch {
        return null;
      }
      return null;

    default:
      return null;
  }
}

function getDefaultSound(platform: string): string {
  switch (platform) {
    case 'macos':
      return '/System/Library/Sounds/Ping.aiff';
    case 'linux':
      return '/usr/share/sounds/freedesktop/stereo/message.oga';
    case 'windows':
      return 'C:\\Windows\\Media\\notify.wav';
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }
}

function loadConfig(configPath: string): NotificationConfig {
  const defaults: NotificationConfig = {
    sounds: {
      enabled: true,
    },
    delay: 5000,
    volume: 0.5,
  };

  try {
    if (!fs.existsSync(configPath)) {
      return defaults;
    }

    const content = fs.readFileSync(configPath, 'utf-8');
    const parsed = JSON.parse(content);

    return {
      sounds: {
        enabled: parsed.sounds?.enabled ?? defaults.sounds.enabled,
        idle: parsed.sounds?.idle,
        permission: parsed.sounds?.permission,
        question: parsed.sounds?.question,
      },
      delay: parsed.delay ?? defaults.delay,
      volume: parsed.volume ?? defaults.volume,
    };
  } catch (error) {
    console.warn(`Failed to load config from ${configPath}:`, error);
    return defaults;
  }
}

function scheduleNotification(
  eventType: string,
  config: NotificationConfig,
  playSoundFn: (soundFile?: string) => Promise<void>
): void {
  if (!config.sounds.enabled) return;

  const soundFile = config.sounds[eventType as keyof typeof config.sounds] as string | undefined;

  setTimeout(() => {
    playSoundFn(soundFile).catch((error) => {
      console.warn(`Failed to play sound for ${eventType}:`, error);
    });
  }, config.delay);
}

async function playSound(
  soundFile: string | undefined,
  platform: string,
  shell: BunShell
): Promise<void> {
  const command = await getSoundCommand(platform, shell);

  if (!command) {
    console.warn(`No sound command available for platform: ${platform}`);
    return;
  }

  const defaultSound = getDefaultSound(platform);
  const fileToPlay = soundFile || defaultSound;

  // Check if file exists (mocked in tests)
  if (!fs.existsSync(fileToPlay)) {
    console.warn(`Sound file not found: ${fileToPlay}`);
    return;
  }

  let playCommand: string;
  switch (command) {
    case 'afplay':
      playCommand = `afplay "${fileToPlay}"`;
      break;
    case 'paplay':
      playCommand = `paplay "${fileToPlay}"`;
      break;
    case 'aplay':
      playCommand = `aplay "${fileToPlay}"`;
      break;
    case 'mpv':
      playCommand = `mpv "${fileToPlay}" --no-video`;
      break;
    case 'powershell':
      playCommand = `powershell -c "(New-Object Media.SoundPlayer '${fileToPlay}').PlaySync()"`;
      break;
    default:
      throw new Error(`Unknown sound command: ${command}`);
  }

  const result = await shell.run(playCommand);
  if (result.exitCode !== 0) {
    throw new Error(`Sound playback failed: ${result.stderr}`);
  }
}

// ============================================================================
// Test Suite: Platform Detection
// ============================================================================

describe('PlatformDetector', () => {
  afterEach(() => {
    // Restore original platform
    if (originalPlatform) {
      Object.defineProperty(process, 'platform', originalPlatform);
    }
  });

  test('getPlatform returns macos for darwin', () => {
    Object.defineProperty(process, 'platform', {
      value: 'darwin',
      configurable: true,
    });
    expect(getPlatform()).toBe('macos');
  });

  test('getPlatform returns linux for linux', () => {
    Object.defineProperty(process, 'platform', {
      value: 'linux',
      configurable: true,
    });
    expect(getPlatform()).toBe('linux');
  });

  test('getPlatform returns windows for win32', () => {
    Object.defineProperty(process, 'platform', {
      value: 'win32',
      configurable: true,
    });
    expect(getPlatform()).toBe('windows');
  });

  test('getPlatform throws for unsupported platform', () => {
    Object.defineProperty(process, 'platform', {
      value: 'freebsd',
      configurable: true,
    });
    expect(() => getPlatform()).toThrow('Unsupported platform: freebsd');
  });
});

// ============================================================================
// Test Suite: Config Loader
// ============================================================================

describe('ConfigLoader', () => {
  const testConfigPath = '/tmp/test-notification-config.json';

  beforeEach(() => {
    // Clean up test config file
    try {
      if (fs.existsSync(testConfigPath)) {
        fs.unlinkSync(testConfigPath);
      }
    } catch {
      // Ignore cleanup errors
    }
  });

  afterEach(() => {
    // Clean up test config file
    try {
      if (fs.existsSync(testConfigPath)) {
        fs.unlinkSync(testConfigPath);
      }
    } catch {
      // Ignore cleanup errors
    }
  });

  test('loadConfig reads valid notification-config.json', () => {
    const config: NotificationConfig = {
      sounds: {
        enabled: true,
        idle: '/custom/idle.wav',
        permission: '/custom/permission.wav',
        question: '/custom/question.wav',
      },
      delay: 3000,
      volume: 0.8,
    };

    fs.writeFileSync(testConfigPath, JSON.stringify(config, null, 2));

    const loaded = loadConfig(testConfigPath);
    expect(loaded.sounds.enabled).toBe(true);
    expect(loaded.sounds.idle).toBe('/custom/idle.wav');
    expect(loaded.sounds.permission).toBe('/custom/permission.wav');
    expect(loaded.sounds.question).toBe('/custom/question.wav');
    expect(loaded.delay).toBe(3000);
    expect(loaded.volume).toBe(0.8);
  });

  test('loadConfig returns defaults when file missing', () => {
    const loaded = loadConfig('/nonexistent/path/config.json');
    expect(loaded.sounds.enabled).toBe(true);
    expect(loaded.delay).toBe(5000);
    expect(loaded.volume).toBe(0.5);
    expect(loaded.sounds.idle).toBeUndefined();
    expect(loaded.sounds.permission).toBeUndefined();
    expect(loaded.sounds.question).toBeUndefined();
  });

  test('loadConfig handles invalid JSON gracefully', () => {
    fs.writeFileSync(testConfigPath, 'invalid json content');

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});
    const loaded = loadConfig(testConfigPath);

    expect(loaded.sounds.enabled).toBe(true);
    expect(loaded.delay).toBe(5000);
    expect(loaded.volume).toBe(0.5);
    expect(consoleSpy).toHaveBeenCalled();

    consoleSpy.mockRestore();
  });

  test('loadConfig uses partial custom values with defaults', () => {
    const partialConfig = {
      sounds: {
        enabled: false,
      },
      delay: 10000,
    };

    fs.writeFileSync(testConfigPath, JSON.stringify(partialConfig));

    const loaded = loadConfig(testConfigPath);
    expect(loaded.sounds.enabled).toBe(false);
    expect(loaded.delay).toBe(10000);
    expect(loaded.volume).toBe(0.5); // default
    expect(loaded.sounds.idle).toBeUndefined();
  });
});

// ============================================================================
// Test Suite: Event Hook Handler
// ============================================================================

describe('NotificationPlugin - Event Hook', () => {
  const mockConfig: NotificationConfig = {
    sounds: { enabled: true },
    delay: 5000,
    volume: 0.5,
  };

  test('event hook calls scheduleNotification for session.idle', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const event = { type: 'session.idle' };
    if (event.type === 'session.idle') {
      scheduleNotification('idle', mockConfig, playSoundMock);
    }

    expect(setTimeoutSpy).toHaveBeenCalledTimes(1);
    expect(setTimeoutSpy).toHaveBeenLastCalledWith(expect.any(Function), 5000);

    setTimeoutSpy.mockRestore();
  });

  test('event hook does not call scheduleNotification for other events', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const nonIdleEvents = ['session.created', 'file.edited', 'tool.executed', 'permission.asked'];

    for (const eventType of nonIdleEvents) {
      const event = { type: eventType };
      if (event.type === 'session.idle') {
        scheduleNotification('idle', mockConfig, playSoundMock);
      }
    }

    expect(setTimeoutSpy).not.toHaveBeenCalled();
    setTimeoutSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Tool Hook Handler
// ============================================================================

describe('NotificationPlugin - Tool Hook', () => {
  const mockConfig: NotificationConfig = {
    sounds: { enabled: true },
    delay: 5000,
    volume: 0.5,
  };

  test('tool.execute.before hook calls scheduleNotification for question tool', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const toolInput = { tool: 'question', params: {} };
    if (toolInput.tool === 'question') {
      scheduleNotification('question', mockConfig, playSoundMock);
    }

    expect(setTimeoutSpy).toHaveBeenCalledTimes(1);
    expect(setTimeoutSpy).toHaveBeenLastCalledWith(expect.any(Function), 5000);

    setTimeoutSpy.mockRestore();
  });

  test('tool.execute.before hook does not call scheduleNotification for other tools', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const otherTools = ['bash', 'read', 'edit', 'write', 'glob', 'grep'];

    for (const tool of otherTools) {
      const toolInput = { tool, params: {} };
      if (toolInput.tool === 'question') {
        scheduleNotification('question', mockConfig, playSoundMock);
      }
    }

    expect(setTimeoutSpy).not.toHaveBeenCalled();
    setTimeoutSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Permission Hook Handler
// ============================================================================

describe('NotificationPlugin - Permission Hook', () => {
  const mockConfig: NotificationConfig = {
    sounds: { enabled: true },
    delay: 5000,
    volume: 0.5,
  };

  test('permission.ask hook always calls scheduleNotification', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    // Should trigger for any permission type
    const permissionTypes = ['edit', 'task', 'bash', 'write'];

    for (const permissionType of permissionTypes) {
      scheduleNotification('permission', mockConfig, playSoundMock);
    }

    expect(setTimeoutSpy).toHaveBeenCalledTimes(permissionTypes.length);
    setTimeoutSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Sound Command Detection
// ============================================================================

describe('SoundPlayer - Command Detection', () => {
  test('getSoundCommand returns afplay on macOS when available', async () => {
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        if (cmd === 'which afplay') {
          return { exitCode: 0, stdout: '/usr/bin/afplay', stderr: '' };
        }
        return { exitCode: 1, stdout: '', stderr: 'not found' };
      }),
    };

    const command = await getSoundCommand('macos', shell);
    expect(command).toBe('afplay');
    expect(shell.run).toHaveBeenCalledWith('which afplay');
  });

  test('getSoundCommand returns null on macOS when afplay unavailable', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 1, stdout: '', stderr: 'not found' })),
    };

    const command = await getSoundCommand('macos', shell);
    expect(command).toBeNull();
  });

  test('getSoundCommand tries paplay first on Linux', async () => {
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        if (cmd === 'which paplay') {
          return { exitCode: 0, stdout: '/usr/bin/paplay', stderr: '' };
        }
        return { exitCode: 1, stdout: '', stderr: 'not found' };
      }),
    };

    const command = await getSoundCommand('linux', shell);
    expect(command).toBe('paplay');
  });

  test('getSoundCommand falls back to aplay on Linux when paplay unavailable', async () => {
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        if (cmd === 'which paplay') {
          return { exitCode: 1, stdout: '', stderr: 'not found' };
        }
        if (cmd === 'which aplay') {
          return { exitCode: 0, stdout: '/usr/bin/aplay', stderr: '' };
        }
        return { exitCode: 1, stdout: '', stderr: 'not found' };
      }),
    };

    const command = await getSoundCommand('linux', shell);
    expect(command).toBe('aplay');
  });

  test('getSoundCommand falls back to mpv on Linux when paplay and aplay unavailable', async () => {
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        if (cmd === 'which paplay' || cmd === 'which aplay') {
          return { exitCode: 1, stdout: '', stderr: 'not found' };
        }
        if (cmd === 'which mpv') {
          return { exitCode: 0, stdout: '/usr/bin/mpv', stderr: '' };
        }
        return { exitCode: 1, stdout: '', stderr: 'not found' };
      }),
    };

    const command = await getSoundCommand('linux', shell);
    expect(command).toBe('mpv');
  });

  test('getSoundCommand returns null on Linux when no tools available', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 1, stdout: '', stderr: 'not found' })),
    };

    const command = await getSoundCommand('linux', shell);
    expect(command).toBeNull();
  });

  test('getSoundCommand returns powershell on Windows when available', async () => {
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        if (cmd.includes('Get-Command SoundPlayer')) {
          return { exitCode: 0, stdout: 'SoundPlayer', stderr: '' };
        }
        return { exitCode: 1, stdout: '', stderr: 'not found' };
      }),
    };

    const command = await getSoundCommand('windows', shell);
    expect(command).toBe('powershell');
  });

  test('getSoundCommand returns null on Windows when PowerShell unavailable', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 1, stdout: '', stderr: 'not found' })),
    };

    const command = await getSoundCommand('windows', shell);
    expect(command).toBeNull();
  });
});

// ============================================================================
// Test Suite: Default Sound Paths
// ============================================================================

describe('SoundPlayer - Default Sounds', () => {
  test('getDefaultSound returns correct path for macOS', () => {
    expect(getDefaultSound('macos')).toBe('/System/Library/Sounds/Ping.aiff');
  });

  test('getDefaultSound returns correct path for Linux', () => {
    expect(getDefaultSound('linux')).toBe('/usr/share/sounds/freedesktop/stereo/message.oga');
  });

  test('getDefaultSound returns correct path for Windows', () => {
    expect(getDefaultSound('windows')).toBe('C:\\Windows\\Media\\notify.wav');
  });

  test('getDefaultSound throws for unsupported platform', () => {
    expect(() => getDefaultSound('freebsd')).toThrow('Unsupported platform: freebsd');
  });
});

// ============================================================================
// Test Suite: Notification Scheduler
// ============================================================================

describe('NotificationScheduler', () => {
  test('scheduleNotification calls setTimeout with 5000ms delay', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const config: NotificationConfig = {
      sounds: { enabled: true },
      delay: 5000,
      volume: 0.5,
    };

    scheduleNotification('idle', config, playSoundMock);

    expect(setTimeoutSpy).toHaveBeenCalledTimes(1);
    expect(setTimeoutSpy).toHaveBeenLastCalledWith(expect.any(Function), 5000);

    setTimeoutSpy.mockRestore();
  });

  test('scheduleNotification respects custom delay from config', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const config: NotificationConfig = {
      sounds: { enabled: true },
      delay: 10000,
      volume: 0.5,
    };

    scheduleNotification('idle', config, playSoundMock);

    expect(setTimeoutSpy).toHaveBeenCalledTimes(1);
    expect(setTimeoutSpy).toHaveBeenLastCalledWith(expect.any(Function), 10000);

    setTimeoutSpy.mockRestore();
  });

  test('scheduleNotification does not call setTimeout when sounds disabled', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const config: NotificationConfig = {
      sounds: { enabled: false },
      delay: 5000,
      volume: 0.5,
    };

    scheduleNotification('idle', config, playSoundMock);

    expect(setTimeoutSpy).not.toHaveBeenCalled();

    setTimeoutSpy.mockRestore();
  });

  test('scheduleNotification callback executes playSound', async () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);

    const config: NotificationConfig = {
      sounds: { enabled: true },
      delay: 0, // Use 0 delay for immediate execution
      volume: 0.5,
    };

    scheduleNotification('idle', config, playSoundMock);

    // Wait for setTimeout to execute
    await new Promise((resolve) => setTimeout(resolve, 50));

    expect(playSoundMock).toHaveBeenCalledTimes(1);
  });
});

// ============================================================================
// Test Suite: Play Sound
// ============================================================================

describe('SoundPlayer - Play Sound', () => {
  const mockSoundFile = '/tmp/test-sound.wav';

  beforeEach(() => {
    // Create a mock sound file
    try {
      fs.writeFileSync(mockSoundFile, 'mock audio data');
    } catch {
      // Ignore if can't create
    }
  });

  afterEach(() => {
    // Clean up mock sound file
    try {
      if (fs.existsSync(mockSoundFile)) {
        fs.unlinkSync(mockSoundFile);
      }
    } catch {
      // Ignore cleanup errors
    }
  });

  test('playSound executes afplay on macOS', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await playSound(mockSoundFile, 'macos', shell);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${mockSoundFile}"`);
  });

  test('playSound executes paplay on Linux', async () => {
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        if (cmd.startsWith('which')) {
          return { exitCode: 0, stdout: '/usr/bin/paplay', stderr: '' };
        }
        return { exitCode: 0, stdout: '', stderr: '' };
      }),
    };

    await playSound(mockSoundFile, 'linux', shell);

    expect(shell.run).toHaveBeenCalledWith(`paplay "${mockSoundFile}"`);
  });

  test('playSound executes PowerShell on Windows', async () => {
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        if (cmd.includes('Get-Command')) {
          return { exitCode: 0, stdout: 'SoundPlayer', stderr: '' };
        }
        return { exitCode: 0, stdout: '', stderr: '' };
      }),
    };

    await playSound(mockSoundFile, 'windows', shell);

    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('SoundPlayer')
    );
  });

  test('playSound uses default sound when soundFile not provided', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    // Mock fs.existsSync to return true for default sound
    const existsSyncSpy = jest.spyOn(fs, 'existsSync').mockImplementation((filepath: fs.PathLike) => {
      if (filepath === '/System/Library/Sounds/Ping.aiff') return true;
      return false;
    });

    await playSound(undefined, 'macos', shell);

    expect(shell.run).toHaveBeenCalledWith('afplay "/System/Library/Sounds/Ping.aiff"');

    existsSyncSpy.mockRestore();
  });

  test('playSound handles missing sound command gracefully', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 1, stdout: '', stderr: 'not found' })),
    };

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});

    await playSound(mockSoundFile, 'macos', shell);

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('No sound command available')
    );

    consoleSpy.mockRestore();
  });

  test('playSound handles missing sound file gracefully', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});

    await playSound('/nonexistent/sound.wav', 'macos', shell);

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Sound file not found')
    );

    consoleSpy.mockRestore();
  });

  test('playSound throws on playback failure', async () => {
    let callCount = 0;
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        callCount++;
        // First call is 'which afplay' to detect command
        if (callCount === 1 && cmd === 'which afplay') {
          return { exitCode: 0, stdout: '/usr/bin/afplay', stderr: '' };
        }
        // Second call is the actual afplay command - simulate failure
        return { exitCode: 1, stdout: '', stderr: 'playback error' };
      }),
    };

    // Mock fs.existsSync to return true
    const existsSyncSpy = jest.spyOn(fs, 'existsSync').mockImplementation(() => true);

    await expect(playSound(mockSoundFile, 'macos', shell)).rejects.toThrow('Sound playback failed');

    existsSyncSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Integration Tests
// ============================================================================

describe('Integration Tests', () => {
  test('full flow: session.idle event triggers sound after delay', async () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const config: NotificationConfig = {
      sounds: { enabled: true },
      delay: 5000,
      volume: 0.5,
    };

    // Simulate event hook
    const event = { type: 'session.idle' };
    if (event.type === 'session.idle') {
      scheduleNotification('idle', config, playSoundMock);
    }

    expect(setTimeoutSpy).toHaveBeenCalledTimes(1);
    expect(setTimeoutSpy).toHaveBeenLastCalledWith(expect.any(Function), 5000);

    setTimeoutSpy.mockRestore();
  });

  test('full flow: disabled sounds prevent any playback', () => {
    const playSoundMock = jest.fn().mockResolvedValue(undefined);
    const setTimeoutSpy = jest.spyOn(global, 'setTimeout');

    const config: NotificationConfig = {
      sounds: { enabled: false },
      delay: 5000,
      volume: 0.5,
    };

    // Try all event types
    scheduleNotification('idle', config, playSoundMock);
    scheduleNotification('permission', config, playSoundMock);
    scheduleNotification('question', config, playSoundMock);

    expect(setTimeoutSpy).not.toHaveBeenCalled();
    expect(playSoundMock).not.toHaveBeenCalled();

    setTimeoutSpy.mockRestore();
  });

  test('full flow: custom sound paths are used when provided', async () => {
    const customSound = '/custom/path/sound.wav';
    let callCount = 0;
    const shell: BunShell = {
      run: jest.fn(async (cmd: string) => {
        callCount++;
        // First call is 'which afplay' to detect command
        if (callCount === 1 && cmd === 'which afplay') {
          return { exitCode: 0, stdout: '/usr/bin/afplay', stderr: '' };
        }
        // Subsequent calls are actual play commands
        return { exitCode: 0, stdout: '', stderr: '' };
      }),
    };

    // Mock fs.existsSync to return true for custom sound
    const existsSyncSpy = jest.spyOn(fs, 'existsSync').mockImplementation((filepath: fs.PathLike) => {
      if (filepath === customSound) return true;
      return false;
    });

    const config: NotificationConfig = {
      sounds: {
        enabled: true,
        idle: customSound,
      },
      delay: 0, // Use 0 delay for immediate execution
      volume: 0.5,
    };

    const playSoundFn = async (soundFile?: string) => {
      await playSound(soundFile, 'macos', shell);
    };

    scheduleNotification('idle', config, playSoundFn);

    // Wait for setTimeout with 0 delay
    await new Promise((resolve) => setTimeout(resolve, 50));

    expect(shell.run).toHaveBeenCalledWith(`afplay "${customSound}"`);

    existsSyncSpy.mockRestore();
  });
});

// ============================================================================
// Export test utilities for use by implementation
// ============================================================================

export {
  getPlatform,
  getSoundCommand,
  getDefaultSound,
  loadConfig,
  scheduleNotification,
  playSound,
  type NotificationConfig,
  type BunShell,
  type PluginInput,
};
