/**
 * Permission Save Plugin for OpenCode
 *
 * When a user grants a permission, this plugin stores it and asks the user
 * about saving it when the PRIMARY session goes idle.
 *
 * Key insight: Plugin state (module-level variables) is shared across all
 * plugin invocations, including subagents. This allows us to track permissions
 * from subagents and only prompt the user when the primary session is idle.
 *
 * Flow:
 * 1. permission.asked event stores permission details
 * 2. permission.replied event stores granted permissions (with root session ID)
 * 3. session.idle event triggers asking about saved permissions ONLY for primary sessions
 * 4. Agent uses question tool to ask about each permission
 * 5. User can customize the pattern
 * 6. Agent saves to opencode.json
 */

import type { Plugin, PluginInput } from '@opencode-ai/plugin';
import { appendFileSync, mkdirSync } from 'fs';
import { join } from 'path';

// Type definitions for events that may not be in the SDK types yet
// The server uses v2 SDK types which have different property names
interface PermissionAskedEvent {
  type: 'permission.asked';
  properties: {
    id: string;
    sessionID: string;
    permission: string;
    patterns: string[];
    metadata: Record<string, unknown>;
  };
}

interface PermissionRepliedEvent {
  type: 'permission.replied';
  properties: {
    sessionID: string;
    requestID: string;
    reply: 'once' | 'always' | 'reject';
  };
}

interface SessionIdleEvent {
  type: 'session.idle';
  properties: {
    sessionID: string;
  };
}

// Type guard to check if event is a permission.asked event
function isPermissionAskedEvent(event: unknown): event is PermissionAskedEvent {
  return typeof event === 'object' && event !== null && 
    'type' in event && (event as { type: string }).type === 'permission.asked';
}

// Type guard to check if event is a permission.replied event
function isPermissionRepliedEvent(event: unknown): event is PermissionRepliedEvent {
  return typeof event === 'object' && event !== null && 
    'type' in event && (event as { type: string }).type === 'permission.replied';
}

// Type guard to check if event is a session.idle event
function isSessionIdleEvent(event: unknown): event is SessionIdleEvent {
  return typeof event === 'object' && event !== null && 
    'type' in event && (event as { type: string }).type === 'session.idle';
}

// Store permission details by ID for later use
interface PermissionInfo {
  id: string;
  sessionID: string;        // Original session that asked for permission
  rootSessionID: string;     // Root (primary) session ID
  permission: string;
  patterns: string[];
  metadata: Record<string, unknown>;
}

// Store for permission.asked events (short-lived, just to get details)
const permissionAskStore = new Map<string, PermissionInfo>();

// Store for granted permissions waiting to be asked about (persists until primary session idle)
// Keyed by permission ID, value includes root session ID for filtering
const grantedPermissionsStore = new Map<string, PermissionInfo>();

// Log file path
const LOG_FILE = join(process.env.HOME || '/tmp', '.opencode', 'permission-save-plugin.log');

// Ensure log directory exists
try {
  mkdirSync(join(process.env.HOME || '/tmp', '.opencode'), { recursive: true });
} catch {}

// Logging function that writes to file
function log(message: string): void {
  const timestamp = new Date().toISOString();
  const logLine = `[${timestamp}] ${message}\n`;
  try {
    appendFileSync(LOG_FILE, logLine);
  } catch (err) {
    console.error('[permission-save-plugin] Failed to write log:', err);
  }
}

log('=== Plugin module loaded ===');

/**
 * Find the root (primary) session ID by traversing up the parent chain.
 * If the session has no parent, it is the root.
 */
async function findRootSessionID(
  client: PluginInput['client'],
  sessionID: string
): Promise<string> {
  try {
    const result = await client.session.get({ path: { id: sessionID } });
    if (result.data?.parentID) {
      // This is a subagent, recurse to find the root
      return findRootSessionID(client, result.data.parentID);
    }
    // This is a primary session (no parent)
    return sessionID;
  } catch (err) {
    log(`Error finding root session for ${sessionID}: ${err}`);
    // Fallback to the original session ID
    return sessionID;
  }
}

/**
 * Permission Save Plugin
 * 
 * Asks users if they want to permanently save granted permissions to opencode.json
 * Only prompts when the PRIMARY session goes idle, not subagent sessions.
 */
export const PermissionSavePlugin: Plugin = async (input) => {
  log('Plugin initialized');
  
  return {
    // Listen for all events
    event: async ({ event }) => {
      // Cast to unknown to work around SDK type limitations
      // The server uses v2 SDK types which include permission.asked/replied events
      const unknownEvent = event as unknown;
      
      // Handle permission.asked event - store the permission details temporarily
      if (isPermissionAskedEvent(unknownEvent)) {
        const props = unknownEvent.properties;
        log(`Permission asked event: id=${props.id} permission=${props.permission} patterns=${JSON.stringify(props.patterns)}`);
        
        // Find the root session ID
        const rootSessionID = await findRootSessionID(input.client, props.sessionID);
        log(`Root session ID for ${props.sessionID}: ${rootSessionID}`);
        
        permissionAskStore.set(props.id, {
          id: props.id,
          sessionID: props.sessionID,
          rootSessionID: rootSessionID,
          permission: props.permission,
          patterns: props.patterns || [],
          metadata: props.metadata || {},
        });
        log(`Stored permission in ask store. Ask store has ${permissionAskStore.size} entries`);
        return;
      }
      
      // Handle permission.replied event - store granted permissions for later
      if (isPermissionRepliedEvent(unknownEvent)) {
        const props = unknownEvent.properties;
        const requestID = props.requestID;
        const reply = props.reply;
        
        log(`Permission replied event: requestID=${requestID} reply=${reply}`);
        
        // Only proceed if user granted permission (reply is "once" or "always")
        if (reply !== 'once' && reply !== 'always') {
          log(`Permission not granted, reply was: ${reply}`);
          // Clean up from ask store
          permissionAskStore.delete(requestID);
          return;
        }

        // Get the stored permission details from ask store
        const permissionInfo = permissionAskStore.get(requestID);
        if (!permissionInfo) {
          log(`Permission details not found for ID: ${requestID}`);
          log(`Available IDs in ask store: ${Array.from(permissionAskStore.keys()).join(', ')}`);
          return;
        }

        // Clean up from ask store
        permissionAskStore.delete(requestID);
        
        // Store in granted permissions store for later asking
        // The rootSessionID is already set from the permission.asked handler
        grantedPermissionsStore.set(requestID, permissionInfo);
        log(`Stored granted permission for root session ${permissionInfo.rootSessionID}. Granted store has ${grantedPermissionsStore.size} entries`);
        return;
      }
      
      // Handle session.idle event - ask about saved permissions ONLY for primary sessions
      if (isSessionIdleEvent(unknownEvent)) {
        const sessionID = unknownEvent.properties.sessionID;
        log(`Session idle event: sessionID=${sessionID}`);
        
        // Check if this is a primary session (no parentID)
        // Subagent sessions should NOT trigger the permission save prompt
        try {
          const sessionResult = await input.client.session.get({
            path: { id: sessionID },
          });
          
          if (sessionResult.data?.parentID) {
            log(`Session ${sessionID} is a subagent (parent: ${sessionResult.data.parentID}), skipping permission save prompt`);
            return;
          }
          log(`Session ${sessionID} is a primary session, proceeding with permission save prompt`);
        } catch (err) {
          log(`Could not fetch session details for ${sessionID}: ${err}`);
          // Continue anyway - better to prompt than to miss permissions
        }
        
        // Check if we have any granted permissions to ask about
        if (grantedPermissionsStore.size === 0) {
          log('No granted permissions to ask about');
          return;
        }
        
        // Get all permissions for this root session (includes permissions from subagents)
        const permissionsToAsk = Array.from(grantedPermissionsStore.values())
          .filter(p => p.rootSessionID === sessionID);
        
        if (permissionsToAsk.length === 0) {
          log(`No granted permissions for root session ${sessionID}`);
          return;
        }
        
        log(`Found ${permissionsToAsk.length} granted permissions to ask about for root session ${sessionID}`);
        
        // Clear the store for this root session
        for (const perm of permissionsToAsk) {
          grantedPermissionsStore.delete(perm.id);
        }
        
        // Build the message
        if (permissionsToAsk.length === 1) {
          // Single permission - simple question
          const perm = permissionsToAsk[0];
          const originalPattern = perm.patterns.length > 0 ? perm.patterns.join(', ') : perm.permission;
          // Default pattern has wildcard appended for bash commands
          const defaultPattern = perm.permission === 'bash' ? `${originalPattern} *` : originalPattern;
          
          const message = `SYSTEM: The user granted a ${perm.permission} permission with pattern "${originalPattern}" during this session.

Use the question tool to ask if they want to save this permission permanently. Ask a SINGLE question with these options:
1. "Save with pattern: ${defaultPattern}" - Add to opencode.json (recommended)
2. "Save with custom pattern" - Specify a different pattern
3. "Don't save" - Skip

If saving, use the config tool to update opencode.json:
- For bash: permission.bash["${defaultPattern}"] = "allow"
- For edit: permission.edit["${defaultPattern}"] = "allow"
- For other types: permission.${perm.permission} = "allow"`;
          
          try {
            log(`Sending single permission message to session: ${sessionID}`);
            await input.client.session.prompt({
              path: { id: sessionID },
              body: {
                parts: [{
                  type: 'text',
                  text: message,
                }],
              },
            });
            log('Message sent successfully');
          } catch (err) {
            log(`Error sending message: ${err}`);
          }
        } else {
          // Multiple permissions - ask about each
          const permList = permissionsToAsk.map((perm, i) => {
            const originalPattern = perm.patterns.length > 0 ? perm.patterns.join(', ') : perm.permission;
            const defaultPattern = perm.permission === 'bash' ? `${originalPattern} *` : originalPattern;
            return `${i + 1}. ${perm.permission}: "${defaultPattern}"`;
          }).join('\n');
          
          const message = `SYSTEM: The user granted ${permissionsToAsk.length} permissions during this session:

${permList}

Use the question tool to ask which permissions they want to save permanently. Ask a SINGLE question allowing multiple selections, with an option to customize patterns.

For each permission they want to save, use the config tool to update opencode.json:
- For bash: permission.bash["pattern *"] = "allow" (append wildcard to the pattern)
- For edit: permission.edit["pattern"] = "allow"
- For other types: permission.type = "allow"`;
          
          try {
            log(`Sending multiple permissions message to session: ${sessionID}`);
            await input.client.session.prompt({
              path: { id: sessionID },
              body: {
                parts: [{
                  type: 'text',
                  text: message,
                }],
              },
            });
            log('Message sent successfully');
          } catch (err) {
            log(`Error sending message: ${err}`);
          }
        }
      }
    },
  };
};
