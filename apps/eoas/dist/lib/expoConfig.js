"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createOrModifyExpoConfigAsync = exports.getExpoConfigUpdateUrl = exports.getPublicExpoConfigAsync = exports.isUsingStaticExpoConfig = exports.ensureExpoConfigExists = exports.getPrivateExpoConfigAsync = exports.RequestedPlatform = void 0;
const tslib_1 = require("tslib");
// This file is copied from eas-cli[https://github.com/expo/eas-cli] to ensure consistent user experience across the CLI.
const config_1 = require("@expo/config");
const spawn_async_1 = tslib_1.__importDefault(require("@expo/spawn-async"));
const fs_extra_1 = tslib_1.__importDefault(require("fs-extra"));
const joi_1 = tslib_1.__importDefault(require("joi"));
const jscodeshift_1 = tslib_1.__importDefault(require("jscodeshift"));
const path_1 = tslib_1.__importDefault(require("path"));
const log_1 = tslib_1.__importDefault(require("./log"));
const package_1 = require("./package");
var RequestedPlatform;
(function (RequestedPlatform) {
    RequestedPlatform["Android"] = "android";
    RequestedPlatform["Ios"] = "ios";
    RequestedPlatform["All"] = "all";
})(RequestedPlatform || (exports.RequestedPlatform = RequestedPlatform = {}));
let wasExpoConfigWarnPrinted = false;
async function getExpoConfigInternalAsync(projectDir, opts = {}) {
    const originalProcessEnv = process.env;
    try {
        process.env = {
            ...process.env,
            ...opts.env,
        };
        let exp;
        if ((0, package_1.isExpoInstalled)(projectDir)) {
            try {
                const { stdout } = await (0, spawn_async_1.default)('npx', ['expo', 'config', '--json', ...(opts.isPublicConfig ? ['--type', 'public'] : [])], {
                    cwd: projectDir,
                    env: {
                        ...process.env,
                        ...opts.env,
                        EXPO_NO_DOTENV: '1',
                    },
                });
                exp = JSON.parse(stdout);
            }
            catch (err) {
                if (!wasExpoConfigWarnPrinted) {
                    log_1.default.warn(`Failed to read the app config from the project using "npx expo config" command: ${err.message}.`);
                    log_1.default.warn('Falling back to the version of "@expo/config" shipped with the EAS CLI.');
                    wasExpoConfigWarnPrinted = true;
                }
                exp = (0, config_1.getConfig)(projectDir, {
                    skipSDKVersionRequirement: true,
                    ...(opts.isPublicConfig ? { isPublicConfig: true } : {}),
                    ...(opts.skipPlugins ? { skipPlugins: true } : {}),
                }).exp;
            }
        }
        else {
            exp = (0, config_1.getConfig)(projectDir, {
                skipSDKVersionRequirement: true,
                ...(opts.isPublicConfig ? { isPublicConfig: true } : {}),
                ...(opts.skipPlugins ? { skipPlugins: true } : {}),
            }).exp;
        }
        const { error } = MinimalAppConfigSchema.validate(exp, {
            allowUnknown: true,
            abortEarly: true,
        });
        if (error) {
            throw new Error(`Invalid app config.\n${error.message}`);
        }
        return exp;
    }
    finally {
        process.env = originalProcessEnv;
    }
}
const MinimalAppConfigSchema = joi_1.default.object({
    slug: joi_1.default.string().required(),
    name: joi_1.default.string().required(),
    version: joi_1.default.string(),
    android: joi_1.default.object({
        versionCode: joi_1.default.number().integer(),
    }),
    ios: joi_1.default.object({
        buildNumber: joi_1.default.string(),
    }),
});
async function getPrivateExpoConfigAsync(projectDir, opts = {}) {
    ensureExpoConfigExists(projectDir);
    return await getExpoConfigInternalAsync(projectDir, { ...opts, isPublicConfig: false });
}
exports.getPrivateExpoConfigAsync = getPrivateExpoConfigAsync;
function ensureExpoConfigExists(projectDir) {
    const paths = (0, config_1.getConfigFilePaths)(projectDir);
    if (!paths?.staticConfigPath && !paths?.dynamicConfigPath) {
        // eslint-disable-next-line node/no-sync
        fs_extra_1.default.writeFileSync(path_1.default.join(projectDir, 'app.json'), JSON.stringify({ expo: {} }, null, 2));
    }
}
exports.ensureExpoConfigExists = ensureExpoConfigExists;
function isUsingStaticExpoConfig(projectDir) {
    const paths = (0, config_1.getConfigFilePaths)(projectDir);
    return !!(paths.staticConfigPath?.endsWith('app.json') && !paths.dynamicConfigPath);
}
exports.isUsingStaticExpoConfig = isUsingStaticExpoConfig;
async function getPublicExpoConfigAsync(projectDir, opts = {}) {
    ensureExpoConfigExists(projectDir);
    return await getExpoConfigInternalAsync(projectDir, { ...opts, isPublicConfig: true });
}
exports.getPublicExpoConfigAsync = getPublicExpoConfigAsync;
function getExpoConfigUpdateUrl(config) {
    return config.updates?.url;
}
exports.getExpoConfigUpdateUrl = getExpoConfigUpdateUrl;
async function createOrModifyExpoConfigAsync(projectDir, exp) {
    try {
        ensureExpoConfigExists(projectDir);
        const configPathJS = path_1.default.join(projectDir, 'app.config.js');
        const configPathTS = path_1.default.join(projectDir, 'app.config.ts');
        // eslint-disable-next-line node/no-sync
        const hasJsConfig = fs_extra_1.default.existsSync(configPathJS);
        if (isUsingStaticExpoConfig(projectDir)) {
            log_1.default.withInfo('You are using a static app config. We will create a dynamic config file for you.');
            const newConfigContent = `export default ({ config }) => ({
                                ...config,
                                ...${stringifyWithEnv(exp)}
                              });`;
            // eslint-disable-next-line node/no-sync
            fs_extra_1.default.writeFileSync(configPathJS, newConfigContent);
        }
        else if (hasJsConfig) {
            // eslint-disable-next-line node/no-sync
            const existingCode = fs_extra_1.default.readFileSync(configPathJS, 'utf8');
            const j = jscodeshift_1.default;
            const ast = j(existingCode);
            ast.find(j.ArrowFunctionExpression).forEach(path => {
                if (path.value.body &&
                    j.BlockStatement.check(path.value.body) &&
                    path.value.body.body.length > 0) {
                    const returnStatement = path.value.body.body.find(node => j.ReturnStatement.check(node));
                    if (returnStatement &&
                        j.ReturnStatement.check(returnStatement) &&
                        returnStatement.argument) {
                        const configObject = returnStatement.argument;
                        if (j.ObjectExpression.check(configObject)) {
                            updateObjectExpression(j, configObject, exp);
                        }
                    }
                }
            });
            const updatedCode = ast.toSource({
                quote: 'auto',
                trailingComma: true,
                reuseWhitespace: true,
            });
            // eslint-disable-next-line node/no-sync
            fs_extra_1.default.writeFileSync(configPathJS, updatedCode);
        }
        else if (configPathTS) {
            log_1.default.warn('TypeScript support is not yet implemented.');
            throw new Error('TypeScript support is not yet implemented.');
        }
    }
    catch (e) {
        log_1.default.withInfo('An error occurred while updating the Expo config. Please update it manually.');
        log_1.default.newLine();
        log_1.default.warn('Please modify your app.config.ts file manually by adding the following code:');
        log_1.default.newLine();
        log_1.default.withInfo(`${stringifyWithEnv(exp)}`);
        log_1.default.newLine();
        throw e;
    }
}
exports.createOrModifyExpoConfigAsync = createOrModifyExpoConfigAsync;
function updateObjectExpression(j, configObject, updates) {
    Object.entries(updates).forEach(([key, value]) => {
        const existingProperty = configObject.properties.find(prop => {
            return (prop.type === 'Property' &&
                ((prop.key.type === 'Identifier' && prop.key.name === key) ||
                    (prop.key.type === 'StringLiteral' && prop.key.value === key)));
        });
        if (existingProperty) {
            configObject.properties = configObject.properties.filter(prop => prop !== existingProperty);
        }
        const newProperty = j.objectProperty(j.identifier(key), createValueNode(j, value));
        configObject.properties.push(newProperty);
    });
}
function createValueNode(j, value) {
    if (typeof value === 'string' && value.startsWith('process.env.')) {
        return j.memberExpression(j.memberExpression(j.identifier('process'), j.identifier('env')), j.identifier(value.split('.')[2]));
    }
    if (typeof value === 'object' && value !== null) {
        return j.objectExpression(Object.entries(value).map(([key, val]) => j.objectProperty(j.stringLiteral(key), createValueNode(j, val)) // Force stringLiteral pour garder les guillemets
        ));
    }
    return j.literal(value);
}
function stringifyWithEnv(obj) {
    return JSON.stringify(obj, null, 2).replace(/"process\.env\.(\w+)"/g, 'process.env.$1');
}
