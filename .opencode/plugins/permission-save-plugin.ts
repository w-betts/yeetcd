/**
 * Permission Save Plugin for OpenCode
 *
 * When a user grants a permission, this plugin stores it and asks the user
 * about saving it when the session goes idle.
 *
 * Flow:
 * 1. permission.asked event stores permission details
 * 2. permission.replied event stores granted permissions
 * 3. session.idle event triggers asking about saved permissions
 * 4. Agent uses question tool to ask about each permission
 * 5. User can customize the pattern
 * 6. Agent saves to opencode.json
 */

import type { Plugin } from '@opencode-ai/plugin';
import { appendFileSync, mkdirSync } from 'fs';
import { join } from 'path';

// Store permission details by ID for later use
interface PermissionInfo {
  id: string;
  sessionID: string;
  permission: string;
  patterns: string[];
  metadata: Record<string, unknown>;
}

// Store for permission.asked events (short-lived, just to get details)
const permissionAskStore = new Map<string, PermissionInfo>();

// Store for granted permissions waiting to be asked about (persists until idle)
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
 * Permission Save Plugin
 * 
 * Asks users if they want to permanently save granted permissions to opencode.json
 */
export const PermissionSavePlugin: Plugin = async (input) => {
  log('Plugin initialized');
  
  return {
    // Listen for all events
    event: async ({ event }) => {
      // Handle permission.asked event - store the permission details temporarily
      if (event.type === 'permission.asked') {
        const props = event.properties as any;
        log(`Permission asked event: id=${props.id} permission=${props.permission} patterns=${JSON.stringify(props.patterns)}`);
        
        permissionAskStore.set(props.id, {
          id: props.id,
          sessionID: props.sessionID,
          permission: props.permission,
          patterns: props.patterns || [],
          metadata: props.metadata || {},
        });
        log(`Stored permission in ask store. Ask store has ${permissionAskStore.size} entries`);
        return;
      }
      
      // Handle permission.replied event - store granted permissions for later
      if (event.type === 'permission.replied') {
        const props = event.properties as any;
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
        grantedPermissionsStore.set(requestID, permissionInfo);
        log(`Stored granted permission. Granted store has ${grantedPermissionsStore.size} entries`);
        return;
      }
      
      // Handle session.idle event - ask about saved permissions
      if (event.type === 'session.idle') {
        const sessionID = (event.properties as any).sessionID;
        log(`Session idle event: sessionID=${sessionID}`);
        
        // Check if we have any granted permissions to ask about
        if (grantedPermissionsStore.size === 0) {
          log('No granted permissions to ask about');
          return;
        }
        
        // Get all permissions for this session
        const permissionsToAsk = Array.from(grantedPermissionsStore.values())
          .filter(p => p.sessionID === sessionID);
        
        if (permissionsToAsk.length === 0) {
          log(`No granted permissions for session ${sessionID}`);
          return;
        }
        
        log(`Found ${permissionsToAsk.length} granted permissions to ask about`);
        
        // Clear the store for this session
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
