/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Permission } from './model_Permission';
export type model_Role = {
    created_at?: string;
    description?: string;
    /**
     * Use string for ID
     */
    id?: string;
    name?: string;
    permissions?: Array<model_Permission>;
    updated_at?: string;
};
