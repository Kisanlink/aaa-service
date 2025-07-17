/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_AddressRes } from './model_AddressRes';
import type { model_RoleDetail } from './model_RoleDetail';
export type model_UserResponse = {
    aadhaar_number?: string;
    address?: model_AddressRes;
    care_of?: string;
    country_code?: string;
    created_at?: string;
    date_of_birth?: string;
    email_hash?: string;
    id?: string;
    is_validated?: boolean;
    message?: string;
    mobile_number?: number;
    name?: string;
    photo?: string;
    roles?: Array<model_RoleDetail>;
    share_code?: string;
    status?: string;
    updated_at?: string;
    username?: string;
    year_of_birth?: string;
};
