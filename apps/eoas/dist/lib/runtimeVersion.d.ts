import { ExpoConfig } from '@expo/config';
import { Env, Workflow } from '@expo/eas-build-job';
export declare class ExpoUpdatesCLIModuleNotFoundError extends Error {
}
export declare class ExpoUpdatesCLIInvalidCommandError extends Error {
}
export declare class ExpoUpdatesCLICommandFailedError extends Error {
}
export declare function expoUpdatesCommandAsync(projectDir: string, args: string[], options: {
    env: Env | undefined;
    cwd?: string;
}): Promise<string>;
export declare function isModernExpoUpdatesCLIWithRuntimeVersionCommandSupportedAsync(projectDir: string): Promise<boolean>;
export declare function resolveRuntimeVersionUsingCLIAsync({ platform, workflow, projectDir, env, cwd, }: {
    platform: 'ios' | 'android';
    workflow: Workflow;
    projectDir: string;
    env: Env | undefined;
    cwd?: string;
}): Promise<{
    runtimeVersion: string | null;
    expoUpdatesRuntimeFingerprint: {
        fingerprintSources: object[];
        isDebugFingerprintSource: boolean;
    } | null;
    expoUpdatesRuntimeFingerprintHash: string | null;
}>;
export declare function resolveRuntimeVersionAsync({ exp, platform, workflow, projectDir, env, cwd, }: {
    exp: ExpoConfig;
    platform: 'ios' | 'android';
    workflow: Workflow;
    projectDir: string;
    env: Env | undefined;
    cwd?: string;
}): Promise<{
    runtimeVersion: string | null;
    expoUpdatesRuntimeFingerprint: {
        fingerprintSources: object[];
        isDebugFingerprintSource: boolean;
    } | null;
    expoUpdatesRuntimeFingerprintHash: string | null;
} | null>;
