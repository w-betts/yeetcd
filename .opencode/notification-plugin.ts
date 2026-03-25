/**
 * Sound Notification Plugin for OpenCode
 *
 * Plays sound notifications when specific events occur:
 * - session goes idle (session.idle event)
 * - user is asked for permission (permission.ask hook)
 * - question tool is invoked (tool.execute.before hook)
 *
 * Uses afplay (macOS) for sound playback.
 */

import type { Plugin, PluginInput } from '@opencode-ai/plugin';

// ============================================================================
// Type Definitions
// ============================================================================

export interface NotificationConfig {
  enabled: boolean;
  idle?: string;
  permission?: string;
  question?: string;
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
// Config
// ============================================================================

const DEFAULT_CONFIG: NotificationConfig = {
  enabled: true,
};

const DEFAULT_SOUND = '/System/Library/Sounds/Ping.aiff';

// ============================================================================
// Sound Playback
// ============================================================================

export async function playSound(
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
// Plugin Implementation
// ============================================================================

let cachedConfig: NotificationConfig | null = null;

function getConfig(): NotificationConfig {
  if (!cachedConfig) {
    cachedConfig = DEFAULT_CONFIG;
  }
  return cachedConfig;
}

/**
 * Main plugin function that registers hooks with OpenCode
 */
const notificationPlugin: Plugin = async (input: PluginInput) => {
  const config = getConfig();
  const shell = input.$;

  return {
    // Hook: event - triggers on session.idle
    event: async (input: { event: Event }) => {
      if (input.event.type === 'session.idle' && config.enabled) {
        await playSound(config.idle, shell);
      }
    },

    // Hook: permission.ask - triggers on any permission request
    'permission.ask': async (input: Permission, output: { status: "ask" | "deny" | "allow" }) => {
      if (config.enabled) {
        await playSound(config.permission, shell);
      }
    },

    // Hook: tool.execute.before - triggers on question tool
    'tool.execute.before': async (input: ToolInput, output: { args: any }) => {
      if (input.tool === 'question' && config.enabled) {
        await playSound(config.question, shell);
      }
    },
  };
}

export default notificationPlugin;
