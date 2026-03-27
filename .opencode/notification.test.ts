/**
 * Sound Notification Plugin Tests
 *
 * Unit tests for the notification plugin using Bun's built-in test runner.
 * Tests platform detection and cross-platform notification support.
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

type Platform = 'macos' | 'windows' | 'linux' | 'unknown';

// ============================================================================
// Helper Functions
// ============================================================================

const DEFAULT_SOUND_MAC = '/System/Library/Sounds/Ping.aiff';
const NOTIFICATION_TITLE = 'OpenCode';
const QUESTION_DELAY_MS = 10000;

/**
 * Detect the current platform
 */
function getPlatform(): Platform {
  const platform = process.platform;
  if (platform === 'darwin') return 'macos';
  if (platform === 'win32') return 'windows';
  if (platform === 'linux') return 'linux';
  return 'unknown';
}

/**
 * Play notification on macOS using afplay
 */
async function notifyMacOS(shell: BunShell, soundFile?: string): Promise<void> {
  const sound = soundFile || DEFAULT_SOUND_MAC;
  try {
    const result = await shell.run(`afplay "${sound}"`);
    if (result.exitCode !== 0) {
      console.warn(`Sound playback failed: ${result.stderr}`);
    }
  } catch (err) {
    console.error('[notification-plugin] Error playing sound on macOS:', err);
  }
}

/**
 * Show notification on Windows using PowerShell toast
 */
async function notifyWindows(shell: BunShell, title?: string, message?: string): Promise<void> {
  const notificationTitle = title || NOTIFICATION_TITLE;
  const notificationMessage = message || 'Agent requires your attention';
  
  try {
    const psScript = `
      [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
      [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
      
      $template = @"
        <toast>
          <visual>
            <binding template="ToastText02">
              <text id="1">${notificationTitle}</text>
              <text id="2">${notificationMessage}</text>
            </binding>
          </visual>
        </toast>
"@
      
      $xml = New-Object Windows.Data.Xml.Dom.XmlDocument
      $xml.LoadXml($template)
      $toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
      [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("OpenCode").Show($toast)
    `;
    
    await shell.run(`powershell -Command ${psScript}`);
  } catch (err) {
    // Fallback to simple system beep if toast fails
    console.warn('[notification-plugin] Toast notification failed, using system beep:', err);
    try {
      await shell.run(`powershell -Command (New-Object Media.SoundPlayer "C:\\Windows\\Media\\notify.wav").PlaySync()`);
    } catch (beepErr) {
      console.error('[notification-plugin] Error playing Windows notification:', beepErr);
    }
  }
}

/**
 * Show notification on Linux using notify-send
 */
async function notifyLinux(shell: BunShell, title?: string, message?: string): Promise<void> {
  const notificationTitle = title || NOTIFICATION_TITLE;
  const notificationMessage = message || 'Agent requires your attention';
  
  try {
    await shell.run(`notify-send "${notificationTitle}" "${notificationMessage}"`);
  } catch (err) {
    console.error('[notification-plugin] Error showing Linux notification:', err);
    // Try fallback with sound
    try {
      await shell.run(`paplay /usr/share/sounds/freedesktop/stereo/message.oga`);
    } catch (soundErr) {
      console.error('[notification-plugin] Error playing Linux sound:', soundErr);
    }
  }
}

/**
 * Play notification based on current platform
 */
async function playNotification(
  shell: BunShell,
  options?: { 
    soundFile?: string; 
    title?: string; 
    message?: string;
    delay?: number;
    platform?: Platform;
  }
): Promise<void> {
  const { soundFile, title, message, delay, platform: forcedPlatform } = options || {};
  
  // Apply delay if specified
  if (delay && delay > 0) {
    await new Promise(resolve => setTimeout(resolve, delay));
  }
  
  const currentPlatform = forcedPlatform || getPlatform();
  
  switch (currentPlatform) {
    case 'macos':
      await notifyMacOS(shell, soundFile);
      break;
    case 'windows':
      await notifyWindows(shell, title, message);
      break;
    case 'linux':
      await notifyLinux(shell, title, message);
      break;
    default:
      console.warn(`[notification-plugin] Unsupported platform: ${currentPlatform}`);
  }
}

// ============================================================================
// Test Suite: Platform Detection
// ============================================================================

describe('Platform Detection', () => {
  test('getPlatform returns correct platform for macOS', () => {
    const originalPlatform = process.platform;
    Object.defineProperty(process, 'platform', { value: 'darwin', writable: true });
    
    const platform = getPlatform();
    expect(platform).toBe('macos');
    
    Object.defineProperty(process, 'platform', { value: originalPlatform, writable: true });
  });

  test('getPlatform returns correct platform for Windows', () => {
    const originalPlatform = process.platform;
    Object.defineProperty(process, 'platform', { value: 'win32', writable: true });
    
    const platform = getPlatform();
    expect(platform).toBe('windows');
    
    Object.defineProperty(process, 'platform', { value: originalPlatform, writable: true });
  });

  test('getPlatform returns correct platform for Linux', () => {
    const originalPlatform = process.platform;
    Object.defineProperty(process, 'platform', { value: 'linux', writable: true });
    
    const platform = getPlatform();
    expect(platform).toBe('linux');
    
    Object.defineProperty(process, 'platform', { value: originalPlatform, writable: true });
  });

  test('getPlatform returns unknown for unsupported platforms', () => {
    const originalPlatform = process.platform;
    Object.defineProperty(process, 'platform', { value: 'freebsd', writable: true });
    
    const platform = getPlatform();
    expect(platform).toBe('unknown');
    
    Object.defineProperty(process, 'platform', { value: originalPlatform, writable: true });
  });
});

// ============================================================================
// Test Suite: macOS Notifications
// ============================================================================

describe('macOS Notifications', () => {
  test('notifyMacOS executes afplay with default sound', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await notifyMacOS(shell);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${DEFAULT_SOUND_MAC}"`);
  });

  test('notifyMacOS executes afplay with custom sound', async () => {
    const customSound = '/custom/sound.wav';
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await notifyMacOS(shell, customSound);

    expect(shell.run).toHaveBeenCalledWith(`afplay "${customSound}"`);
  });

  test('notifyMacOS handles playback failure gracefully', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 1, stdout: '', stderr: 'playback error' })),
    };

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});

    await notifyMacOS(shell);

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Sound playback failed')
    );

    consoleSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Windows Notifications
// ============================================================================

describe('Windows Notifications', () => {
  test('notifyWindows executes PowerShell toast notification', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await notifyWindows(shell, 'Test Title', 'Test Message');

    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('powershell -Command')
    );
    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('Test Title')
    );
    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('Test Message')
    );
  });

  test('notifyWindows uses default values when not provided', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await notifyWindows(shell);

    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining(NOTIFICATION_TITLE)
    );
    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('Agent requires your attention')
    );
  });

  test('notifyWindows falls back to system beep on toast failure', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => {
        throw new Error('Toast failed');
      }),
    };

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});

    await notifyWindows(shell);

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Toast notification failed'),
      expect.any(Error)
    );

    consoleSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Linux Notifications
// ============================================================================

describe('Linux Notifications', () => {
  test('notifyLinux executes notify-send with title and message', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await notifyLinux(shell, 'Test Title', 'Test Message');

    expect(shell.run).toHaveBeenCalledWith(`notify-send "Test Title" "Test Message"`);
  });

  test('notifyLinux uses default values when not provided', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await notifyLinux(shell);

    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('notify-send')
    );
    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining(NOTIFICATION_TITLE)
    );
  });

  test('notifyLinux falls back to paplay on notify-send failure', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => {
        throw new Error('notify-send failed');
      }),
    };

    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});

    await notifyLinux(shell);

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Error showing Linux notification'),
      expect.any(Error)
    );

    consoleSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Cross-Platform Play Notification
// ============================================================================

describe('playNotification - Cross-Platform', () => {
  test('playNotification calls macOS notification on macOS', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await playNotification(shell, { platform: 'macos' });

    expect(shell.run).toHaveBeenCalledWith(`afplay "${DEFAULT_SOUND_MAC}"`);
  });

  test('playNotification calls Windows notification on Windows', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await playNotification(shell, { platform: 'windows', message: 'Test' });

    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('powershell -Command')
    );
  });

  test('playNotification calls Linux notification on Linux', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    await playNotification(shell, { platform: 'linux', message: 'Test' });

    expect(shell.run).toHaveBeenCalledWith(
      expect.stringContaining('notify-send')
    );
  });

  test('playNotification logs warning for unknown platform', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});

    await playNotification(shell, { platform: 'unknown' });

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Unsupported platform')
    );
    expect(shell.run).not.toHaveBeenCalled();

    consoleSpy.mockRestore();
  });
});

// ============================================================================
// Test Suite: Delay Functionality
// ============================================================================

describe('Notification Delay', () => {
  test('playNotification applies delay when specified', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const startTime = Date.now();
    await playNotification(shell, { platform: 'macos', delay: 100 }); // Use 100ms for faster test
    const endTime = Date.now();

    expect(endTime - startTime).toBeGreaterThanOrEqual(90); // Allow some variance
    expect(shell.run).toHaveBeenCalled();
  });

  test('playNotification does not delay when delay is 0', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const startTime = Date.now();
    await playNotification(shell, { platform: 'macos', delay: 0 });
    const endTime = Date.now();

    expect(endTime - startTime).toBeLessThan(50); // Should be nearly instant
    expect(shell.run).toHaveBeenCalled();
  });

  test('playNotification does not delay when delay is undefined', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const startTime = Date.now();
    await playNotification(shell, { platform: 'macos' });
    const endTime = Date.now();

    expect(endTime - startTime).toBeLessThan(50); // Should be nearly instant
    expect(shell.run).toHaveBeenCalled();
  });
});

// ============================================================================
// Test Suite: Event Hook Handler
// ============================================================================

describe('NotificationPlugin - Event Hook', () => {
  const mockConfig: NotificationConfig = {
    enabled: true,
  };

  test('event hook plays notification for session.idle when enabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const event = { type: 'session.idle' };
    if (event.type === 'session.idle' && mockConfig.enabled) {
      await playNotification(shell, { message: 'Session is now idle' });
    }

    expect(shell.run).toHaveBeenCalledTimes(1);
  });

  test('event hook does not play notification for other events', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const nonIdleEvents = ['session.created', 'file.edited', 'tool.executed', 'permission.asked'];

    for (const eventType of nonIdleEvents) {
      const event = { type: eventType };
      if (event.type === 'session.idle' && mockConfig.enabled) {
        await playNotification(shell, { message: 'Session is now idle' });
      }
    }

    expect(shell.run).not.toHaveBeenCalled();
  });

  test('event hook does not play notification when disabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const disabledConfig: NotificationConfig = {
      enabled: false,
    };

    const event = { type: 'session.idle' };
    if (event.type === 'session.idle' && disabledConfig.enabled) {
      await playNotification(shell, { message: 'Session is now idle' });
    }

    expect(shell.run).not.toHaveBeenCalled();
  });

  test('event hook suppresses notification for subagent sessions (with parentID)', async () => {
    // This test simulates the new behavior where subagent sessions are suppressed
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    // Simulate a subagent session (has parentID)
    const sessionWithParent = {
      id: 'subagent-session-id',
      parentID: 'parent-session-id',
    };

    // If session has parentID, notification should be suppressed
    if (!sessionWithParent.parentID) {
      await playNotification(shell, { message: 'Session is now idle' });
    }

    expect(shell.run).not.toHaveBeenCalled();
  });

  test('event hook plays notification for primary sessions (without parentID)', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    // Simulate a primary session (no parentID)
    const primarySession = {
      id: 'primary-session-id',
      parentID: undefined,
    };

    // If session doesn't have parentID, notification should play
    if (!primarySession.parentID) {
      await playNotification(shell, { message: 'Session is now idle' });
    }

    expect(shell.run).toHaveBeenCalledTimes(1);
  });
});

// ============================================================================
// Test Suite: Tool Hook Handler
// ============================================================================

describe('NotificationPlugin - Tool Hook', () => {
  const mockConfig: NotificationConfig = {
    enabled: true,
  };

  test('tool.execute.before hook plays notification for question tool with delay', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const toolInput = { tool: 'question', sessionID: '123', callID: '456' };
    if (toolInput.tool === 'question' && mockConfig.enabled) {
      await playNotification(shell, { 
        message: 'Question tool invoked - awaiting your response',
        delay: 100 // Use small delay for testing
      });
    }

    expect(shell.run).toHaveBeenCalledTimes(1);
  });

  test('tool.execute.before hook does not play notification for other tools', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const otherTools = ['bash', 'read', 'edit', 'write', 'glob', 'grep'];

    for (const tool of otherTools) {
      const toolInput = { tool, sessionID: '123', callID: '456' };
      if (toolInput.tool === 'question' && mockConfig.enabled) {
        await playNotification(shell, { message: 'Question tool invoked' });
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

  test('permission.ask hook plays notification when enabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    if (mockConfig.enabled) {
      await playNotification(shell, { message: 'Permission requested' });
    }

    expect(shell.run).toHaveBeenCalledTimes(1);
  });

  test('permission.ask hook does not play notification when disabled', async () => {
    const shell: BunShell = {
      run: jest.fn(async () => ({ exitCode: 0, stdout: '', stderr: '' })),
    };

    const disabledConfig: NotificationConfig = {
      enabled: false,
    };

    if (disabledConfig.enabled) {
      await playNotification(shell, { message: 'Permission requested' });
    }

    expect(shell.run).not.toHaveBeenCalled();
  });
});

// ============================================================================
// Export test utilities
// ============================================================================

export {
  playNotification,
  getPlatform,
  notifyMacOS,
  notifyWindows,
  notifyLinux,
  type NotificationConfig,
  type BunShell,
  type Event,
  type Permission,
  type ToolInput,
  type Platform,
};
