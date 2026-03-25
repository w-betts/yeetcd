/**
 * Sound Notification Plugin for OpenCode
 *
 * Plays sound notifications when specific events occur:
 * - session goes idle (session.idle event)
 * - user is asked for permission (permission.ask hook)
 * - question tool is invoked (tool.execute.before hook)
 *
 * Features:
 * - Cross-platform support (macOS, Linux, Windows)
 * - Non-blocking 5-second delay before sound plays
 * - Configurable via notification-config.json
 * - Graceful degradation if sound tools unavailable
 */

import * as fs from 'fs';
import * as path from 'path';
import type { Plugin, PluginInput } from '@opencode-ai/plugin';

// ============================================================================
// Type Definitions
// ============================================================================

export interface NotificationConfig {
  sounds: {
    enabled: boolean;
    idle?: string;
    permission?: string;
    question?: string;
  };
  delay: number;
  volume: number;
}

export interface BunShell {
  run(command: string): Promise<{ exitCode: number; stdout: string; stderr: string }>;
}

export interface Event {
  type: string;
}

export interface Permission {
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

export interface ToolInput {
  tool: string;
  sessionID: string;
  callID: string;
}

// ============================================================================
// Platform Detection
// ============================================================================

export function getPlatform(): 'macos' | 'linux' | 'windows' {
  try {
    const platform = process.platform;
    if (typeof platform !== 'string') {
      // Fallback for environments where process.platform is not a string
      console.warn('process.platform is not a string, defaulting to macOS');
      return 'macos'; // Default to macOS
    }
    if (platform === 'darwin') return 'macos';
    if (platform === 'linux') return 'linux';
    if (platform === 'win32') return 'windows';
    // Default to macOS for unknown platforms
    console.warn(`Unknown platform: ${platform}, defaulting to macOS`);
    return 'macos';
  } catch (error) {
    console.error('Error in getPlatform():', error);
    return 'macos'; // Default to macOS on error
  }
}

// ============================================================================
// Sound Command Detection
// ============================================================================

export async function getAvailableSoundCommand(
  platform: string,
  shell: BunShell
): Promise<string | null> {
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

export async function getSoundCommand(
  platform: string,
  shell: BunShell
): Promise<string | null> {
  return getAvailableSoundCommand(platform, shell);
}

// ============================================================================
// Default Sound Paths
// ============================================================================

export function getDefaultSound(platform: string): string {
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

// ============================================================================
// Config Loader
// ============================================================================

export function loadConfig(configPath: string): NotificationConfig {
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

// ============================================================================
// Sound Playback
// ============================================================================

export async function playSound(
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

  // Check if file exists
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
// Notification Scheduler
// ============================================================================

export function scheduleNotification(
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

// ============================================================================
// Plugin Implementation
// ============================================================================

let cachedConfig: NotificationConfig | null = null;
let cachedPlatform: string | null = null;

function getConfig(): NotificationConfig {
  if (!cachedConfig) {
    const configPath = path.join(__dirname, 'notification-config.json');
    cachedConfig = loadConfig(configPath);
  }
  return cachedConfig;
}

function getCachedPlatform(): string {
  if (!cachedPlatform) {
    cachedPlatform = getPlatform();
  }
  return cachedPlatform;
}

/**
 * Main plugin function that registers hooks with OpenCode
 */
const notificationPlugin: Plugin = async (input: PluginInput) => {
  const config = getConfig();
  const platform = getCachedPlatform();
  const shell = input.$;

  // Create a bound playSound function for this platform and shell
  const playSoundBound = (soundFile?: string) => playSound(soundFile, platform, shell);

  return {
    // Hook: event - triggers on session.idle
    event: async (input: { event: Event }) => {
      if (input.event.type === 'session.idle') {
        scheduleNotification('idle', config, playSoundBound);
      }
    },

    // Hook: permission.ask - triggers on any permission request
    'permission.ask': async (input: Permission, output: { status: "ask" | "deny" | "allow" }) => {
      scheduleNotification('permission', config, playSoundBound);
    },

    // Hook: tool.execute.before - triggers on question tool
    'tool.execute.before': async (input: ToolInput, output: { args: any }) => {
      if (input.tool === 'question') {
        scheduleNotification('question', config, playSoundBound);
      }
    },
  };
}

export default notificationPlugin;
