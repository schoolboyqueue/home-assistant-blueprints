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
