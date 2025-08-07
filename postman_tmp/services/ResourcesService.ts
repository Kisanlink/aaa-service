/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_CreateResourceRequest } from '../models/model_CreateResourceRequest';
import type { model_Resource } from '../models/model_Resource';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ResourcesService {
    /**
     * Get resources
     * Get resources with optional filtering by ID or name and pagination
     * @param id Filter by resource ID
     * @param name Filter by resource name
     * @param page Page number (starts from 1)
     * @param limit Number of items per page
     * @returns any OK
     * @throws ApiError
     */
    public static getResources(
        id?: string,
        name?: string,
        page?: number,
        limit?: number,
    ): CancelablePromise<(helper_Response & {
        data?: Array<model_Resource>;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/resources',
            query: {
                'id': id,
                'name': name,
                'page': page,
                'limit': limit,
            },
        });
    }
    /**
     * Create a new resource
     * Creates a new resource with the provided details
     * @param request Resource creation data
     * @returns any Resource created successfully
     * @throws ApiError
     */
    public static postResources(
        request: model_CreateResourceRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Resource;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/resources',
            body: request,
            errors: {
                400: `Invalid request body`,
                409: `Resource already exists`,
                500: `Failed to create resource`,
            },
        });
    }
    /**
     * Delete a resource
     * Deletes an existing resource by ID
     * @param id Resource ID
     * @returns helper_Response Resource deleted successfully
     * @throws ApiError
     */
    public static deleteResources(
        id: string,
    ): CancelablePromise<helper_Response> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/resources/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Invalid resource ID`,
                500: `Failed to delete resource`,
            },
        });
    }
    /**
     * Update a resource
     * Updates an existing resource with the provided details
     * @param id Resource ID
     * @param request Resource update data
     * @returns any Resource updated successfully
     * @throws ApiError
     */
    public static putResources(
        id: string,
        request: model_CreateResourceRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Resource;
    })> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/resources/{id}',
            path: {
                'id': id,
            },
            body: request,
            errors: {
                400: `Invalid request body`,
                404: `Resource not found`,
                500: `Failed to update resource`,
            },
        });
    }
}
