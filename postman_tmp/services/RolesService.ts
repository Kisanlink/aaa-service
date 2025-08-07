/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_CreateRoleRequest } from '../models/model_CreateRoleRequest';
import type { model_Role } from '../models/model_Role';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RolesService {
    /**
     * Get roles with pagination
     * Retrieves roles with optional filtering by ID or name and pagination support
     * @param id Filter by role ID
     * @param name Filter by role name
     * @param page Page number (starts from 1)
     * @param limit Number of items per page
     * @returns any Roles retrieved successfully
     * @throws ApiError
     */
    public static getRoles(
        id?: string,
        name?: string,
        page?: number,
        limit?: number,
    ): CancelablePromise<(helper_Response & {
        data?: Array<model_Role>;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/roles',
            query: {
                'id': id,
                'name': name,
                'page': page,
                'limit': limit,
            },
            errors: {
                500: `Failed to retrieve roles`,
            },
        });
    }
    /**
     * Create a new role with permissions
     * Creates a new role with associated permissions
     * @param request Role and permissions data
     * @returns any Role created successfully
     * @throws ApiError
     */
    public static postRoles(
        request: model_CreateRoleRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Role;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/roles',
            body: request,
            errors: {
                400: `Invalid request`,
                409: `Role already exists`,
                500: `Failed to create role`,
            },
        });
    }
    /**
     * Delete a role
     * Deletes a role and all its associated permissions
     * @param id Role ID
     * @returns helper_Response Role deleted successfully
     * @throws ApiError
     */
    public static deleteRoles(
        id: string,
    ): CancelablePromise<helper_Response> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/roles/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Invalid role ID`,
                500: `Failed to delete role`,
            },
        });
    }
    /**
     * Update a role with permissions
     * Updates an existing role and its permissions
     * @param id Role ID
     * @param request Role and permissions data
     * @returns any Role updated successfully
     * @throws ApiError
     */
    public static putRoles(
        id: string,
        request: model_CreateRoleRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_Role;
    })> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/roles/{id}',
            path: {
                'id': id,
            },
            body: request,
            errors: {
                400: `Invalid request`,
                404: `Role not found`,
                500: `Failed to update role`,
            },
        });
    }
}
