/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_Action } from '../models/model_Action';
import type { model_CreateActionRequest } from '../models/model_CreateActionRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ActionsService {
    /**
     * Get actions with pagination
     * Retrieves actions with optional filtering by ID or name and pagination support
     * @param id Filter by action ID
     * @param name Filter by action name
     * @param page Page number (starts from 1)
     * @param limit Number of items per page
     * @returns any List of actions retrieved successfully
     * @throws ApiError
     */
    public static getActions(
        id?: string,
        name?: string,
        page?: number,
        limit?: number,
    ): CancelablePromise<(helper_Response & {
        data?: Array<model_Action>;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/actions',
            query: {
                'id': id,
                'name': name,
                'page': page,
                'limit': limit,
            },
            errors: {
                500: `Failed to retrieve actions`,
            },
        });
    }
    /**
     * Create a new action
     * Creates a new action with the provided details
     * @param request Action creation data
     * @returns any Action created successfully
     * @throws ApiError
     */
    public static postActions(
        request: model_CreateActionRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Action;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/actions',
            body: request,
            errors: {
                400: `Invalid request or missing required fields`,
                409: `Action already exists`,
                500: `Failed to create action`,
            },
        });
    }
    /**
     * Delete an action
     * Deletes an existing action by ID
     * @param id Action ID
     * @returns helper_Response Action deleted successfully
     * @throws ApiError
     */
    public static deleteActions(
        id: string,
    ): CancelablePromise<helper_Response> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/actions/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Invalid action ID`,
                404: `Action not found`,
                500: `Failed to delete action`,
            },
        });
    }
    /**
     * Update an action
     * Updates an existing action with the provided details
     * @param id Action ID
     * @param request Action update data
     * @returns any Action updated successfully
     * @throws ApiError
     */
    public static putActions(
        id: string,
        request: model_CreateActionRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Action;
    })> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/actions/{id}',
            path: {
                'id': id,
            },
            body: request,
            errors: {
                400: `Invalid request body`,
                404: `Action not found`,
                500: `Failed to update action`,
            },
        });
    }
}
