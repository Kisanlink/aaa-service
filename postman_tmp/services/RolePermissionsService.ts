/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_GetRolePermissionResponse } from '../models/model_GetRolePermissionResponse';
import type { model_RolePermission } from '../models/model_RolePermission';
import type { model_RolePermissionRequest } from '../models/model_RolePermissionRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RolePermissionsService {
    /**
     * Get roles with permissions
     * Retrieves roles with their associated permissions, with optional filtering and pagination
     * @param roleId Filter by role ID
     * @param roleName Filter by role name (case-insensitive partial match)
     * @param permissionId Filter by permission ID
     * @param page Page number (starts from 1)
     * @param limit Number of items per page
     * @returns any List of roles with permissions retrieved successfully
     * @throws ApiError
     */
    public static getRolePermissions(
        roleId?: string,
        roleName?: string,
        permissionId?: string,
        page?: number,
        limit?: number,
    ): CancelablePromise<(helper_Response & {
        data?: Array<model_GetRolePermissionResponse>;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/role-permissions',
            query: {
                'role_id': roleId,
                'role_name': roleName,
                'permission_id': permissionId,
                'page': page,
                'limit': limit,
            },
            errors: {
                400: `Invalid request parameters`,
                500: `Failed to retrieve roles with permissions`,
            },
        });
    }
    /**
     * Assign permission to role
     * Creates an association between a role and permission
     * @param request Assignment data
     * @returns any Permission assigned successfully
     * @throws ApiError
     */
    public static postRolePermissions(
        request: model_RolePermissionRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_RolePermission;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/role-permissions',
            body: request,
            errors: {
                400: `Invalid request or missing required fields`,
                404: `Role or permission not found`,
                409: `Association already exists`,
                500: `Failed to assign permission`,
            },
        });
    }
    /**
     * Delete role-permission
     * Deletes a role-permission relationship by its ID
     * @param id Role-Permission ID
     * @returns helper_Response Role-permission deleted successfully
     * @throws ApiError
     */
    public static deleteRolePermissions(
        id: string,
    ): CancelablePromise<helper_Response> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/role-permissions/{id}',
            path: {
                'id': id,
            },
            errors: {
                404: `Role-permission not found`,
                500: `Failed to delete role-permission`,
            },
        });
    }
    /**
     * Get role-permission by ID
     * Retrieves a single role-permission relationship by its ID
     * @param id Role-Permission ID
     * @returns any Role-permission retrieved successfully
     * @throws ApiError
     */
    public static getRolePermissions1(
        id: string,
    ): CancelablePromise<(helper_Response & {
        data?: model_GetRolePermissionResponse;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/role-permissions/{id}',
            path: {
                'id': id,
            },
            errors: {
                404: `Role-permission not found`,
                500: `Failed to retrieve role-permission`,
            },
        });
    }
}
