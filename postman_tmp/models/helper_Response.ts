/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Standard API response structure
 */
export type helper_Response = {
    /**
     * The actual data payload (can be any type)
     * @example {"id": 1, "name": "John Doe"}
     */
    data?: any;
    /**
     * List of error messages (if any)
     * @example null
     */
    error?: Array<string>;
    /**
     * Human-readable message about the response
     * @example "Request processed successfully"
     */
    message?: string;
    /**
     * HTTP status code
     * @example 200
     */
    status_code?: number;
    /**
     * Indicates if the request was successfully processed
     * @example true
     */
    success?: boolean;
    /**
     * Timestamp of when the response was generated
     * @example "2023-05-15T10:00:00Z"
     */
    timestamp?: string;
};
