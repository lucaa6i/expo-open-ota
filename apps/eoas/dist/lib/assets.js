"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.requestUploadUrls = exports.computeFilesRequests = exports.MetadataJoi = void 0;
const tslib_1 = require("tslib");
const fs_extra_1 = tslib_1.__importDefault(require("fs-extra"));
const joi_1 = tslib_1.__importDefault(require("joi"));
const path_1 = tslib_1.__importDefault(require("path"));
const auth_1 = require("./auth");
const expoConfig_1 = require("./expoConfig");
const fetch_1 = require("./fetch");
const log_1 = tslib_1.__importDefault(require("./log"));
const fileMetadataJoi = joi_1.default.object({
    assets: joi_1.default.array()
        .required()
        .items(joi_1.default.object({ path: joi_1.default.string().required(), ext: joi_1.default.string().required() })),
    bundle: joi_1.default.string().required(),
}).optional();
exports.MetadataJoi = joi_1.default.object({
    version: joi_1.default.number().required(),
    bundler: joi_1.default.string().required(),
    fileMetadata: joi_1.default.object({
        android: fileMetadataJoi,
        ios: fileMetadataJoi,
        web: fileMetadataJoi,
    }).required(),
}).required();
function loadMetadata(distRoot) {
    // eslint-disable-next-line
    const fileContent = fs_extra_1.default.readFileSync(path_1.default.join(distRoot, 'metadata.json'), 'utf8');
    let metadata;
    try {
        metadata = JSON.parse(fileContent);
    }
    catch (e) {
        log_1.default.error(`Failed to read metadata.json: ${e.message}`);
        throw e;
    }
    const { error } = exports.MetadataJoi.validate(metadata);
    if (error) {
        throw error;
    }
    // Check version and bundler by hand (instead of with Joi) so
    // more informative error messages can be returned.
    if (metadata.version !== 0) {
        throw new Error('Only bundles with metadata version 0 are supported');
    }
    if (metadata.bundler !== 'metro') {
        throw new Error('Only bundles created with Metro are currently supported');
    }
    const platforms = Object.keys(metadata.fileMetadata);
    if (platforms.length === 0) {
        log_1.default.warn('No updates were exported for any platform');
    }
    log_1.default.debug(`Loaded ${platforms.length} platform(s): ${platforms.join(', ')}`);
    return metadata;
}
function computeFilesRequests(projectDir, outputDir, requestedPlatform) {
    const metadata = loadMetadata(path_1.default.join(projectDir, outputDir));
    const assets = [
        { path: 'metadata.json', name: 'metadata.json', ext: 'json' },
        { path: 'expoConfig.json', name: 'expoConfig.json', ext: 'json' },
    ];
    for (const platform of Object.keys(metadata.fileMetadata)) {
        if (requestedPlatform !== expoConfig_1.RequestedPlatform.All && requestedPlatform !== platform) {
            continue;
        }
        const bundle = metadata.fileMetadata[platform].bundle;
        assets.push({ path: bundle, name: path_1.default.basename(bundle), ext: 'hbc' });
        for (const asset of metadata.fileMetadata[platform].assets) {
            assets.push({ path: asset.path, name: path_1.default.basename(asset.path), ext: asset.ext });
        }
    }
    return assets;
}
exports.computeFilesRequests = computeFilesRequests;
async function requestUploadUrls({ body, requestUploadUrl, auth, runtimeVersion, platform, commitHash, }) {
    const response = await (0, fetch_1.fetchWithRetries)(`${requestUploadUrl}?runtimeVersion=${runtimeVersion}&platform=${platform}&commitHash=${commitHash || ''}`, {
        method: 'POST',
        headers: {
            ...(0, auth_1.getAuthExpoHeaders)(auth),
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
    });
    if (!response.ok) {
        const text = await response.text();
        throw new Error(`Failed to request upload URL: ${text}`);
    }
    return await response.json();
}
exports.requestUploadUrls = requestUploadUrls;
