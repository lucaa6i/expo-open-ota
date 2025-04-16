import { ExpoConfig } from '@expo/config';
import { Env } from '@expo/eas-build-job';
export declare enum RequestedPlatform {
    Android = "android",
    Ios = "ios",
    All = "all"
}
export type PublicExpoConfig = Omit<ExpoConfig, '_internal' | 'hooks' | 'ios' | 'android' | 'updates'> & {
    ios?: Omit<ExpoConfig['ios'], 'config'>;
    android?: Omit<ExpoConfig['android'], 'config'>;
    updates?: Omit<ExpoConfig['updates'], 'codeSigningCertificate' | 'codeSigningMetadata'>;
};
export interface ExpoConfigOptions {
    env?: Env;
    skipSDKVersionRequirement?: boolean;
    skipPlugins?: boolean;
}
export declare function getPrivateExpoConfigAsync(projectDir: string, opts?: ExpoConfigOptions): Promise<ExpoConfig>;
export declare function ensureExpoConfigExists(projectDir: string): void;
export declare function isUsingStaticExpoConfig(projectDir: string): boolean;
export declare function getPublicExpoConfigAsync(projectDir: string, opts?: ExpoConfigOptions): Promise<PublicExpoConfig>;
export declare function getExpoConfigUpdateUrl(config: ExpoConfig): string | undefined;
export declare function createOrModifyExpoConfigAsync(projectDir: string, exp: Partial<ExpoConfig>): Promise<void>;
export declare function resolveServerUrl(config: ExpoConfig): Promise<string>;
