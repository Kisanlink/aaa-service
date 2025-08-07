/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_AssignRolePermission } from '../models/model_AssignRolePermission';
import type { model_AssignRoleRequest } from '../models/model_AssignRoleRequest';
import type { model_CreateUserRequest } from '../models/model_CreateUserRequest';
import type { model_CreditUsageRequest } from '../models/model_CreditUsageRequest';
import type { model_MinimalUser } from '../models/model_MinimalUser';
import type { model_UpdateUserRequest } from '../models/model_UpdateUserRequest';
import type { model_UserRes } from '../models/model_UserRes';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class UsersService {
    /**
     * delete a role to a user
     * API to delete a specific role assigned to a user
     * @param request Assign Role Request
     * @returns any delete assigned Role successfully
     * @throws ApiError
     */
    public static deleteAssignRole(
        request: model_AssignRoleRequest,
    ): CancelablePromise<(helper_Response & {
        data?: any;
    })> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/assign-role',
            body: request,
            errors: {
                400: `Invalid request body`,
                404: `User or Role not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Assign a role to a user
     * Assigns a specified role to a user and returns the updated user details with roles and permissions
     * @param request Assign Role Request
     * @returns any Role assigned successfully
     * @throws ApiError
     */
    public static postAssignRole(
        request: model_AssignRoleRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_AssignRolePermission;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/assign-role',
            body: request,
            errors: {
                400: `Invalid request body`,
                404: `User or Role not found`,
                409: `Role already assigned to user`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Create a new user
     * Creates a new user account with the provided details. Optionally sends OTP for Aadhaar verification if Aadhaar number is provided.
     * @param request User creation request
     * @returns any User created successfully
     * @throws ApiError
     */
    public static postRegister(
        request: model_CreateUserRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_MinimalUser;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/register',
            body: request,
            errors: {
                400: `Invalid request body or validation failed`,
                409: `Username, mobile number or Aadhaar already exists`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Manage user tokens
     * Handles token transactions (credit/debit) or fetches token balance when no transaction type is specified
     * @param request Token transaction request
     * @returns any Returns remaining tokens in all cases" example({"remaining_tokens": 100})
     * @throws ApiError
     */
    public static postTokenTransaction(
        request: model_CreditUsageRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/token-transaction',
            body: request,
            errors: {
                400: `Invalid request, insufficient tokens, or invalid transaction type`,
                404: `User not found`,
            },
        });
    }
    /**
     * Get users with pagination
     * Retrieves a list of users including their roles, permissions, and address information with optional pagination
     * @param page Page number (starts from 1)
     * @param limit Number of items per page
     * @returns any Users fetched successfully
     * @throws ApiError
     */
    public static getUsers(
        page?: number,
        limit?: number,
    ): CancelablePromise<(helper_Response & {
        data?: Array<model_UserRes>;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users',
            query: {
                'page': page,
                'limit': limit,
            },
            errors: {
                500: `Internal server error when fetching users or their details`,
            },
        });
    }
    /**
     * Get user by ID
     * Retrieves a single user's details including roles, permissions, and address information by their unique ID
     * @param id User ID
     * @returns any User fetched successfully
     * @throws ApiError
     */
    public static getUsers1(
        id: string,
    ): CancelablePromise<(helper_Response & {
        data?: model_UserRes;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `ID is required`,
                404: `User not found`,
                500: `Internal server error when fetching user or related data`,
            },
        });
    }
    /**
     * Update user
     * Updates user information by ID. Only provided fields will be updated (partial update supported).
     * @param id User ID
     * @param request User update data (partial updates allowed)
     * @returns any User updated successfully
     * @throws ApiError
     */
    public static putUsers(
        id: string,
        request: model_UpdateUserRequest,
    ): CancelablePromise<(helper_Response & {
        data?: model_UserRes;
    })> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/users/{id}',
            path: {
                'id': id,
            },
            body: request,
            errors: {
                400: `Invalid ID or request body`,
                404: `User not found`,
                500: `Failed to update user or fetch related data`,
            },
        });
    }
}
