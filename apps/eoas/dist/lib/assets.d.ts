import Joi from 'joi';
import { ExpoCredentials } from './auth';
import { RequestedPlatform } from './expoConfig';
export declare const MetadataJoi: Joi.ObjectSchema<any>;
interface AssetToUpload {
    path: string;
    name: string;
    ext: string;
}
export declare function computeFilesRequests(projectDir: string, outputDir: string, requestedPlatform: RequestedPlatform): AssetToUpload[];
export interface RequestUploadUrlItem {
    requestUploadUrl: string;
    fileName: string;
    filePath: string;
}
export declare function requestUploadUrls({ body, requestUploadUrl, auth, runtimeVersion, platform, commitHash, }: {
    body: {
        fileNames: string[];
    };
    requestUploadUrl: string;
    auth: ExpoCredentials;
    runtimeVersion: string;
    platform: string;
    commitHash?: string;
}): Promise<{
    uploadRequests: RequestUploadUrlItem[];
    updateId: string;
}>;
export {};
