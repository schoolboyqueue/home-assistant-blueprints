/**
 * WebSocket client utilities for communicating with the Home Assistant API.
 * @module client
 */

import type WebSocket from 'ws';
import { HAClientError } from './errors.js';
import type { HAMessage, PendingRequest, TriggerSubscription } from './types.js';

// Re-export TriggerSubscription for convenience
export type { TriggerSubscription };

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
 * Subscribe to a Home Assistant trigger and receive events via callback.
 * Handles subscription ID generation, event listener attachment, promise wrapping, and cleanup.
 *
 * @param ws - Active WebSocket connection
 * @param trigger - Trigger configuration (e.g., { platform: 'state', entity_id: 'light.kitchen' })
 * @param callback - Function called when trigger events are received. Receives the event variables.
 * @param timeout - Optional timeout in milliseconds. If provided, automatically cleans up after timeout.
 * @returns Promise resolving to TriggerSubscription with subscriptionId and cleanup function
 * @throws {Error} If subscription fails
 *
 * @example
 * ```typescript
 * // Subscribe to state changes for an entity
 * const { subscriptionId, cleanup } = await subscribeToTrigger(
 *   ws,
 *   { platform: 'state', entity_id: 'light.kitchen' },
 *   (variables) => {
 *     const trigger = variables.trigger as { to_state?: { state: string } };
 *     console.log('New state:', trigger.to_state?.state);
 *   }
 * );
 *
 * // Later, when done listening:
 * cleanup();
 *
 * // Or with automatic timeout cleanup:
 * const { cleanup } = await subscribeToTrigger(
 *   ws,
 *   { platform: 'state', entity_id: 'sensor.temperature' },
 *   (variables) => console.log('Temperature changed:', variables),
 *   30000 // Auto-cleanup after 30 seconds
 * );
 * ```
 */
export async function subscribeToTrigger(
  ws: WebSocket,
  trigger: Record<string, unknown>,
  callback: (variables: Record<string, unknown>) => void,
  timeout?: number
): Promise<TriggerSubscription> {
  // Generate unique subscription ID
  const subId = nextId();

  // Create promise for subscription confirmation
  const subPromise = new Promise<void>((resolve, reject) => {
    pendingRequests.set(subId, { resolve: resolve as (value: unknown) => void, reject });
  });

  // Send subscription request
  ws.send(
    JSON.stringify({
      id: subId,
      type: 'subscribe_trigger',
      trigger,
    })
  );

  // Wait for subscription confirmation
  await subPromise;

  // Create event handler that filters by subscription ID
  const eventHandler = (data: WebSocket.Data): void => {
    let msg: HAMessage;
    try {
      msg = JSON.parse(data.toString()) as HAMessage;
    } catch {
      return;
    }

    // Only process events for this subscription
    if (msg.type === 'event' && msg.id === subId && msg.event?.variables) {
      callback(msg.event.variables);
    }
  };

  // Attach event listener
  ws.on('message', eventHandler);

  // Create cleanup function
  const cleanup = (): void => {
    ws.removeListener('message', eventHandler);
  };

  // If timeout is specified, set up automatic cleanup
  if (timeout !== undefined && timeout > 0) {
    setTimeout(cleanup, timeout);
  }

  return {
    subscriptionId: subId,
    cleanup,
  };
}

/**
 * Re-export HAClientError for convenience in handlers.
 */
export { HAClientError };
