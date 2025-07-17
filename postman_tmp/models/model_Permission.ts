/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type model_Permission = {
    /**
     * Actions  pq.StringArray `json:"actions" gorm:"type:text[]"`
     */
    actions?: Array<string>;
    created_at?: string;
    effect?: string;
    /**
     * Use string for ID
     */
    id?: string;
    resource?: string;
    updated_at?: string;
};
