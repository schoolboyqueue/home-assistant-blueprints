/**
 * Custom error types for the Home Assistant WebSocket API client.
 * @module errors
 */

/**
 * Base error class for Home Assistant client errors.
 * All custom errors in this module extend this class.
 */
export class HAClientError extends Error {
  /**
   * Create a new HAClientError.
   * @param message - Human-readable error message
   * @param code - Optional error code for programmatic handling
   */
  constructor(
    message: string,
    public readonly code?: string
  ) {
    super(message);
    this.name = 'HAClientError';
  }
}

/**
 * Error thrown when a requested entity is not found.
 * Typically occurs when querying for a non-existent entity_id.
 */
export class EntityNotFoundError extends HAClientError {
  /**
   * Create a new EntityNotFoundError.
   * @param entityId - The entity_id that was not found
   */
  constructor(entityId: string) {
    super(`Entity not found: ${entityId}`, 'ENTITY_NOT_FOUND');
    this.name = 'EntityNotFoundError';
  }
}

/**
 * Error thrown when authentication with Home Assistant fails.
 * Typically occurs when the SUPERVISOR_TOKEN is invalid or expired.
 */
export class AuthenticationError extends HAClientError {
  /**
   * Create a new AuthenticationError.
   * @param message - Description of why authentication failed
   */
  constructor(message: string) {
    super(message, 'AUTH_FAILED');
    this.name = 'AuthenticationError';
  }
}

/**
 * Error thrown when a date string cannot be parsed.
 * Supports ISO format, "YYYY-MM-DD HH:MM", and "MM/DD HH:MM" formats.
 */
export class DateParseError extends HAClientError {
  /**
   * Create a new DateParseError.
   * @param dateStr - The date string that could not be parsed
   */
  constructor(dateStr: string) {
    super(`Invalid date format: ${dateStr}. Use "YYYY-MM-DD HH:MM" or ISO format.`, 'INVALID_DATE');
    this.name = 'DateParseError';
  }
}

/**
 * Error thrown when JSON parsing fails for a user-provided argument.
 * Provides structured error messages with field context.
 */
export class JsonParseError extends HAClientError {
  /**
   * Create a new JsonParseError.
   * @param fieldName - The name of the field/argument that contained invalid JSON
   * @param originalError - The original SyntaxError from JSON.parse
   * @param jsonString - The JSON string that failed to parse
   */
  constructor(fieldName: string, originalError: Error, jsonString: string) {
    // Extract position info from the original error message if available
    const positionMatch = originalError.message.match(/position\s+(\d+)/i);
    const position = positionMatch?.[1] ? parseInt(positionMatch[1], 10) : null;

    // Build a helpful error message
    let message = `Invalid JSON in ${fieldName}: ${originalError.message}`;

    // Add context showing where the error occurred
    if (position !== null && jsonString.length > 0) {
      const start = Math.max(0, position - 20);
      const end = Math.min(jsonString.length, position + 20);
      const snippet = jsonString.slice(start, end);
      const pointer = `${' '.repeat(Math.min(position, 20))}^`;
      message += `\n  Near: ...${snippet}...\n        ${pointer}`;
    }

    message += `\n  Tip: Ensure your JSON is properly quoted. Example: '{"key": "value"}'`;

    super(message, 'INVALID_JSON');
    this.name = 'JsonParseError';
  }
}
