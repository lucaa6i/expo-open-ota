export interface ExpoCredentials {
    token?: string;
    sessionSecret?: string;
}
export declare function retrieveExpoCredentials(): ExpoCredentials;
export declare function getAuthExpoHeaders(credentials: ExpoCredentials): Record<string, string>;
