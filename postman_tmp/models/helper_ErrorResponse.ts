/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Standard API error response structure with multiple error messages
 */
export type helper_ErrorResponse = {
    /**
     * The actual data payload (nil for error responses)
     * @example null
     */
    data?: any;
    /**
     * List of error messages describing what went wrong
     * @example ["Invalid email format", "Password must be at least 8 characters"]
     */
    errors?: Array<string>;
    /**
     * Human-readable summary message about the error
     * @example "Validation failed"
     */
    message?: string;
    /**
     * HTTP status code
     * @example 400
     */
    status_code?: number;
    /**
     * Indicates if the request was successfully processed (always false for error responses)
     * @example false
     */
    success?: boolean;
    /**
     * Timestamp of when the response was generated
     * @example "2023-05-15T10:00:00Z"
     */
    timestamp?: string;
};
