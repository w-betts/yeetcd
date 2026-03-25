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

import type { Plugin } from '@opencode-ai/plugin';
import { $ } from 'bun';

const DEFAULT_SOUND = '/System/Library/Sounds/Ping.aiff';

/**
 * Sound Notification Plugin
 * 
 * Plays a sound when:
 * - Session goes idle (agent finished responding)
 * - Permission is requested
 * - Question tool is invoked
 */
export const NotificationPlugin: Plugin = async () => {
  return {
    // Play sound when session goes idle
    event: async ({ event }) => {
      if (event.type === 'session.idle') {
        try {
          await $`afplay ${DEFAULT_SOUND}`;
        } catch (err) {
          console.error('[notification-plugin] Error playing idle sound:', err);
        }
      }
    },

    // Play sound when permission is requested
    'permission.ask': async (input, output) => {
      try {
        await $`afplay ${DEFAULT_SOUND}`;
      } catch (err) {
        console.error('[notification-plugin] Error playing permission sound:', err);
      }
    },

    // Play sound when question tool is invoked
    'tool.execute.before': async (input, output) => {
      if (input.tool === 'question') {
        try {
          await $`afplay ${DEFAULT_SOUND}`;
        } catch (err) {
          console.error('[notification-plugin] Error playing question sound:', err);
        }
      }
    },
  };
};
