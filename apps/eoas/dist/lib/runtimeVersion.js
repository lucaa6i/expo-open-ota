"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.resolveRuntimeVersionAsync = exports.resolveRuntimeVersionUsingCLIAsync = exports.isModernExpoUpdatesCLIWithRuntimeVersionCommandSupportedAsync = exports.expoUpdatesCommandAsync = exports.ExpoUpdatesCLICommandFailedError = exports.ExpoUpdatesCLIInvalidCommandError = exports.ExpoUpdatesCLIModuleNotFoundError = void 0;
const tslib_1 = require("tslib");
const config_plugins_1 = require("@expo/config-plugins");
const spawn_async_1 = tslib_1.__importDefault(require("@expo/spawn-async"));
const fs_extra_1 = tslib_1.__importDefault(require("fs-extra"));
const resolve_from_1 = tslib_1.__importStar(require("resolve-from"));
const semver_1 = tslib_1.__importDefault(require("semver"));
const log_1 = tslib_1.__importStar(require("./log"));
class ExpoUpdatesCLIModuleNotFoundError extends Error {
}
exports.ExpoUpdatesCLIModuleNotFoundError = ExpoUpdatesCLIModuleNotFoundError;
class ExpoUpdatesCLIInvalidCommandError extends Error {
}
exports.ExpoUpdatesCLIInvalidCommandError = ExpoUpdatesCLIInvalidCommandError;
class ExpoUpdatesCLICommandFailedError extends Error {
}
exports.ExpoUpdatesCLICommandFailedError = ExpoUpdatesCLICommandFailedError;
async function expoUpdatesCommandAsync(projectDir, args, options) {
    let expoUpdatesCli;
    try {
        expoUpdatesCli =
            (0, resolve_from_1.silent)(projectDir, 'expo-updates/bin/cli') ??
                (0, resolve_from_1.default)(projectDir, 'expo-updates/bin/cli.js');
    }
    catch (e) {
        if (e.code === 'MODULE_NOT_FOUND') {
            throw new ExpoUpdatesCLIModuleNotFoundError(`The \`expo-updates\` package was not found. Follow the installation directions at ${(0, log_1.link)('https://docs.expo.dev/bare/installing-expo-modules/')}`);
        }
        throw e;
    }
    try {
        return (await (0, spawn_async_1.default)(expoUpdatesCli, args, {
            stdio: 'pipe',
            env: { ...process.env, ...options.env },
            cwd: options.cwd,
        })).stdout;
    }
    catch (e) {
        if (e.stderr && typeof e.stderr === 'string') {
            if (e.stderr.includes('Invalid command')) {
                throw new ExpoUpdatesCLIInvalidCommandError(`The command specified by ${args} was not valid in the \`expo-updates\` CLI.`);
            }
            else {
                throw new ExpoUpdatesCLICommandFailedError(e.stderr);
            }
        }
        throw e;
    }
}
exports.expoUpdatesCommandAsync = expoUpdatesCommandAsync;
async function getExpoUpdatesPackageVersionIfInstalledAsync(projectDir) {
    const maybePackageJson = resolve_from_1.default.silent(projectDir, 'expo-updates/package.json');
    if (!maybePackageJson) {
        return null;
    }
    const { version } = await fs_extra_1.default.readJson(maybePackageJson);
    return version ?? null;
}
async function isModernExpoUpdatesCLIWithRuntimeVersionCommandSupportedAsync(projectDir) {
    const expoUpdatesPackageVersion = await getExpoUpdatesPackageVersionIfInstalledAsync(projectDir);
    if (expoUpdatesPackageVersion === null) {
        return false;
    }
    if (expoUpdatesPackageVersion.includes('canary')) {
        return true;
    }
    // Anything SDK 51 or greater uses the expo-updates CLI
    return semver_1.default.gte(expoUpdatesPackageVersion, '0.25.4');
}
exports.isModernExpoUpdatesCLIWithRuntimeVersionCommandSupportedAsync = isModernExpoUpdatesCLIWithRuntimeVersionCommandSupportedAsync;
async function resolveRuntimeVersionUsingCLIAsync({ platform, workflow, projectDir, env, cwd, }) {
    log_1.default.debug('Using expo-updates runtimeversion:resolve CLI for runtime version resolution');
    const useDebugFingerprintSource = log_1.default.isDebug;
    const extraArgs = useDebugFingerprintSource ? ['--debug'] : [];
    const resolvedRuntimeVersionJSONResult = await expoUpdatesCommandAsync(projectDir, ['runtimeversion:resolve', '--platform', platform, '--workflow', workflow, ...extraArgs], { env, cwd });
    const runtimeVersionResult = JSON.parse(resolvedRuntimeVersionJSONResult);
    log_1.default.debug('runtimeversion:resolve output:');
    log_1.default.debug(resolvedRuntimeVersionJSONResult);
    return {
        runtimeVersion: runtimeVersionResult.runtimeVersion ?? null,
        expoUpdatesRuntimeFingerprint: runtimeVersionResult.fingerprintSources
            ? {
                fingerprintSources: runtimeVersionResult.fingerprintSources,
                isDebugFingerprintSource: useDebugFingerprintSource,
            }
            : null,
        expoUpdatesRuntimeFingerprintHash: runtimeVersionResult.fingerprintSources
            ? runtimeVersionResult.runtimeVersion
            : null,
    };
}
exports.resolveRuntimeVersionUsingCLIAsync = resolveRuntimeVersionUsingCLIAsync;
async function resolveRuntimeVersionAsync({ exp, platform, workflow, projectDir, env, cwd, }) {
    if (!(await isModernExpoUpdatesCLIWithRuntimeVersionCommandSupportedAsync(projectDir))) {
        // fall back to the previous behavior (using the @expo/config-plugins eas-cli dependency rather
        // than the versioned @expo/config-plugins dependency in the project)
        return {
            runtimeVersion: await config_plugins_1.Updates.getRuntimeVersionNullableAsync(projectDir, exp, platform),
            expoUpdatesRuntimeFingerprint: null,
            expoUpdatesRuntimeFingerprintHash: null,
        };
    }
    try {
        return await resolveRuntimeVersionUsingCLIAsync({ platform, workflow, projectDir, env, cwd });
    }
    catch (e) {
        // if expo-updates is not installed, there's no need for a runtime version in the build
        if (e instanceof ExpoUpdatesCLIModuleNotFoundError) {
            return null;
        }
        throw e;
    }
}
exports.resolveRuntimeVersionAsync = resolveRuntimeVersionAsync;
