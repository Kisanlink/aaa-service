/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_LoginRequest } from '../models/model_LoginRequest';
import type { model_PasswordResetFlowRequest } from '../models/model_PasswordResetFlowRequest';
import type { model_UserResponse } from '../models/model_UserResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AuthenticationService {
    /**
     * Password reset flow
     * Handles the complete password reset flow in three steps: 1) Request OTP, 2) Verify OTP, 3) Reset password. Each step requires different request parameters.
     * @param request Password reset request
     * @returns any Success responses vary by step: 1) 'OTP sent successfully', 2) 'OTP verified. Proceed to reset password.', 3) 'Password reset successfully'
     * @throws ApiError
     */
    public static postForgotPassword(
        request: model_PasswordResetFlowRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/forgot-password',
            body: request,
            errors: {
                400: `Invalid request body or parameters`,
                401: `Invalid or expired OTP`,
                404: `User not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * User login
     * Authenticates a user with username and password, returns JWT tokens in response headers and optional user details in body. If 'source: admin/panel' header is provided, only users with ADMIN, SUPER_ADMIN, or CUSTOMER_SUPPORT roles can login.
     * @param request Login credentials
     * @param userDetails Include full user details
     * @param source Access source (e.g., 'admin/panel' for admin panel access)
     * @returns any Login successful
     * @throws ApiError
     */
    public static postLogin(
        request: model_LoginRequest,
        userDetails: boolean = false,
        source?: string,
    ): CancelablePromise<(helper_Response & {
        data?: model_UserResponse;
    })> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/login',
            headers: {
                'source': source,
            },
            query: {
                'user_details': userDetails,
            },
            body: request,
            errors: {
                400: `Invalid request body or missing credentials`,
                401: `Invalid credentials`,
                403: `Access Denied: Insufficient permission (when source=admin/panel but user lacks required role)`,
                500: `Internal server error`,
            },
        });
    }
}
