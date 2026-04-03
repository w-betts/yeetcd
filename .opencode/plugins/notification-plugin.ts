/**
 * Sound Notification Plugin for OpenCode
 *
 * Plays sound notifications when specific events occur:
 * - session goes idle (session.idle event)
 * - user is asked for permission (permission.ask hook)
 * - question tool is invoked (tool.execute.before hook)
 *
 * Platform support:
 * - macOS: Uses afplay for sound playback
 * - Windows: Uses PowerShell toast notifications
 * - Linux: Uses notify-send for desktop notifications
 */

import type { Plugin } from '@opencode-ai/plugin';
import { $ } from 'bun';

// Platform detection
const platform = process.platform;
const isMac = platform === 'darwin';
const isWindows = platform === 'win32';
const isLinux = platform === 'linux';

// Default sounds per platform
const DEFAULT_SOUND_MAC = '/System/Library/Sounds/Ping.aiff';
const NOTIFICATION_TITLE = 'OpenCode';
const NOTIFICATION_MESSAGE = 'Agent requires your attention';

/**
 * Detect the current platform
 */
function getPlatform(): 'macos' | 'windows' | 'linux' | 'unknown' {
  if (isMac) return 'macos';
  if (isWindows) return 'windows';
  if (isLinux) return 'linux';
  return 'unknown';
}

/**
 * Play notification on macOS using afplay
 */
async function notifyMacOS(soundFile?: string): Promise<void> {
  const sound = soundFile || DEFAULT_SOUND_MAC;
  try {
    await $`afplay ${sound}`;
  } catch (err) {
    console.error('[notification-plugin] Error playing sound on macOS:', err);
  }
}

/**
 * Show notification on Windows using PowerShell toast
 */
async function notifyWindows(title?: string, message?: string): Promise<void> {
  const notificationTitle = title || NOTIFICATION_TITLE;
  const notificationMessage = message || NOTIFICATION_MESSAGE;
  
  try {
    // Use PowerShell to show a toast notification
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
    
    await $`powershell -Command ${psScript}`;
  } catch (err) {
    // Fallback to simple system beep if toast fails
    console.warn('[notification-plugin] Toast notification failed, using system beep:', err);
    try {
      await $`powershell -Command (New-Object Media.SoundPlayer "C:\\Windows\\Media\\notify.wav").PlaySync()`;
    } catch (beepErr) {
      console.error('[notification-plugin] Error playing Windows notification:', beepErr);
    }
  }
}

/**
 * Show notification on Linux using notify-send
 */
async function notifyLinux(title?: string, message?: string): Promise<void> {
  const notificationTitle = title || NOTIFICATION_TITLE;
  const notificationMessage = message || NOTIFICATION_MESSAGE;
  
  try {
    await $`notify-send ${notificationTitle} ${notificationMessage}`;
  } catch (err) {
    console.error('[notification-plugin] Error showing Linux notification:', err);
    // Try fallback with sound
    try {
      await $`paplay /usr/share/sounds/freedesktop/stereo/message.oga`;
    } catch (soundErr) {
      console.error('[notification-plugin] Error playing Linux sound:', soundErr);
    }
  }
}

/**
 * Play notification based on current platform
 */
async function playNotification(options?: { 
  soundFile?: string; 
  title?: string; 
  message?: string;
  delay?: number;
}): Promise<void> {
  const { soundFile, title, message, delay } = options || {};
  
  // Apply delay if specified
  if (delay && delay > 0) {
    await new Promise(resolve => setTimeout(resolve, delay));
  }
  
  const currentPlatform = getPlatform();
  
  switch (currentPlatform) {
    case 'macos':
      await notifyMacOS(soundFile);
      break;
    case 'windows':
      await notifyWindows(title, message);
      break;
    case 'linux':
      await notifyLinux(title, message);
      break;
    default:
      console.warn(`[notification-plugin] Unsupported platform: ${platform}`);
  }
}

/**
 * Sound Notification Plugin
 * 
 * Plays a sound/notification when:
 * - Session goes idle (agent finished responding) - only for primary sessions
 * - Permission is requested
 * - Question tool is invoked
 */
export const NotificationPlugin: Plugin = async (input) => {
  return {
    // Play notification when session goes idle (only for primary sessions)
    event: async ({ event }) => {
      if (event.type === 'session.idle') {
        const sessionID = event.properties.sessionID;
        
        // Fetch session details to check if it's a subagent
        try {
          const sessionResult = await input.client.session.get({
            path: { id: sessionID },
          });
          
          if (sessionResult.data?.parentID) {
            // This is a subagent session, suppress notification
            return;
          }
        } catch (err) {
          console.warn('[notification-plugin] Could not fetch session details:', err);
          // Continue to play notification if we can't determine session type
        }
        
        await playNotification({
          message: 'Session is now idle',
        });
      }
    },

    // Play notification when permission is requested
    'permission.ask': async (input, output) => {
      await playNotification({
        message: 'Permission requested',
      });
    },

    // Play notification when question tool is invoked (non-blocking via setTimeout)
    'tool.execute.before': async (input, output) => {
      if (input.tool === 'question') {
        // Use setTimeout to not block the question from appearing
        setTimeout(() => {
          playNotification({
            message: 'Question tool invoked - awaiting your response',
          });
        }, 0);
      }
    },
  };
};
