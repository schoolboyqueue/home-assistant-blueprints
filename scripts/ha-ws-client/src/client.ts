/**
 * WebSocket client utilities for communicating with the Home Assistant API.
 * @module client
 */

import type WebSocket from 'ws';
import { HAClientError } from './errors.js';
import type { PendingRequest } from './types.js';

/** Map of pending requests by message ID. */
export const pendingRequests = new Map<number, PendingRequest>();

/** Current message ID counter. */
let messageId = 0;

/**
 * Generate the next unique message ID for WebSocket requests.
 * Message IDs are used to correlate requests with responses.
 * @returns The next available message ID
 */
export function nextId(): number {
  return ++messageId;
}

/**
 * Send a message to Home Assistant and wait for the response.
 * Handles the request/response correlation using message IDs.
 *
 * @typeParam T - The expected response type
 * @param ws - Active WebSocket connection
 * @param type - Message type (e.g., 'get_states', 'call_service')
 * @param data - Additional message data
 * @returns Promise resolving to the result from Home Assistant
 * @throws {HAClientError} If Home Assistant returns an error response
 *
 * @example
 * ```typescript
 * // Get all entity states
 * const states = await sendMessage<HAState[]>(ws, 'get_states');
 *
 * // Call a service
 * await sendMessage(ws, 'call_service', {
 *   domain: 'light',
 *   service: 'turn_on',
 *   service_data: { entity_id: 'light.kitchen' }
 * });
 * ```
 */
export function sendMessage<T = unknown>(
  ws: WebSocket,
  type: string,
  data: Record<string, unknown> = {}
): Promise<T> {
  return new Promise((resolve, reject) => {
    const id = nextId();
    const message = { id, type, ...data };
    pendingRequests.set(id, {
      resolve: resolve as (value: unknown) => void,
      reject,
    });
    ws.send(JSON.stringify(message));
  });
}

/**
 * Re-export HAClientError for convenience in handlers.
 */
export { HAClientError };
