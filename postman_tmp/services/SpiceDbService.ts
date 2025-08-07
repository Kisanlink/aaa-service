/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { helper_Response } from '../models/helper_Response';
import type { model_Role } from '../models/model_Role';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class SpiceDbService {
    /**
     * update spice db schema
     * update schema by Retrieves all roles
     * @returns any Roles retrieved successfully
     * @throws ApiError
     */
    public static getUpdateSchema(): CancelablePromise<(helper_Response & {
        data?: Array<model_Role>;
    })> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/update/schema',
            errors: {
                500: `Failed to retrieve roles`,
            },
        });
    }
}
