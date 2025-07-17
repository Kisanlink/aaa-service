/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_CreatePermissionRequest } from '../models/model_CreatePermissionRequest';
import type { model_Permission } from '../models/model_Permission';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class PermissionsService {
    /**
     * Get all permissions
     * Retrieves all permissions with optional filtering
     * @param roleId Filter by role ID
     * @param resource Filter by resource
     * @param action Filter by action
     * @param page Page number
     * @param limit Items per page
     * @returns any Permissions retrieved successfully
     * @throws ApiError
     */
    public static getPermissions(
        roleId?: string,
        resource?: string,
        action?: string,
        page?: number,
        limit?: number,
    ): CancelablePromise<(helper_Response & {
        data?: Array<model_Permission>;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/permissions',
            query: {
                'roleId': roleId,
                'resource': resource,
                'action': action,
                'page': page,
                'limit': limit,
            },
            errors: {
                400: `Invalid filter parameters`,
                500: `Failed to retrieve permissions`,
            },
        });
    }
    /**
     * Create a new permission
     * Creates a new permission with the provided details
     * @param request Permission creation data
     * @returns any Permission created successfully
     * @throws ApiError
     */
    public static postPermissions(
        request: model_CreatePermissionRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Permission;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/permissions',
            body: request,
            errors: {
                400: `Invalid request or missing required fields`,
                409: `Permission already exists for this role+resource`,
                500: `Failed to create permission`,
            },
        });
    }
    /**
     * Delete a permission
     * Deletes an existing permission by its ID
     * @param id Permission ID
     * @returns helper_Response Permission deleted successfully
     * @throws ApiError
     */
    public static deletePermissions(
        id: string,
    ): CancelablePromise<helper_Response> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/permissions/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Invalid permission ID`,
                404: `Permission not found`,
                500: `Failed to delete permission`,
            },
        });
    }
    /**
     * Get permission by ID
     * Retrieves a permission by its ID
     * @param id Permission ID
     * @returns any Permission retrieved successfully
     * @throws ApiError
     */
    public static getPermissions1(
        id: string,
    ): CancelablePromise<(helper_Response & {
        data?: model_Permission;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/permissions/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Invalid permission ID`,
                404: `Permission not found`,
                500: `Failed to retrieve permission`,
            },
        });
    }
    /**
     * Update a permission
     * Updates an existing permission with the provided details
     * @param id Permission ID
     * @param request Permission update data
     * @returns any Permission updated successfully
     * @throws ApiError
     */
    public static putPermissions(
        id: string,
        request: model_CreatePermissionRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Permission;
    })> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/permissions/{id}',
            path: {
                'id': id,
            },
            body: request,
            errors: {
                400: `Invalid request or missing required fields`,
                404: `Permission not found`,
                409: `Permission already exists for this role+resource`,
                500: `Failed to update permission`,
            },
        });
    }
}
