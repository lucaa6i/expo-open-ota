"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const eas_build_job_1 = require("@expo/eas-build-job");
const spawn_async_1 = tslib_1.__importDefault(require("@expo/spawn-async"));
const core_1 = require("@oclif/core");
const form_data_1 = tslib_1.__importDefault(require("form-data"));
const fs_extra_1 = tslib_1.__importDefault(require("fs-extra"));
const mime_1 = tslib_1.__importDefault(require("mime"));
const path_1 = tslib_1.__importDefault(require("path"));
const assets_1 = require("../lib/assets");
const auth_1 = require("../lib/auth");
const expoConfig_1 = require("../lib/expoConfig");
const fetch_1 = require("../lib/fetch");
const log_1 = tslib_1.__importDefault(require("../lib/log"));
const ora_1 = require("../lib/ora");
const package_1 = require("../lib/package");
const prompts_1 = require("../lib/prompts");
const repo_1 = require("../lib/repo");
const runtimeVersion_1 = require("../lib/runtimeVersion");
const vcs_1 = require("../lib/vcs");
const workflow_1 = require("../lib/workflow");
class Publish extends core_1.Command {
    static args = {};
    static description = 'Publish a new update to the self-hosted update server';
    static examples = ['<%= config.bin %> <%= command.id %>'];
    static flags = {
        platform: core_1.Flags.string({
            type: 'option',
            options: Object.values(expoConfig_1.RequestedPlatform),
            default: expoConfig_1.RequestedPlatform.All,
            required: false,
        }),
        channel: core_1.Flags.string({
            description: 'Name of the channel to publish the update to',
            required: true,
        }),
        branch: core_1.Flags.string({
            description: 'Name of the branch to point to',
            required: true,
        }),
        nonInteractive: core_1.Flags.boolean({
            description: 'Run command in non-interactive mode',
            default: false,
        }),
        outputDir: core_1.Flags.string({
            description: "Where to write build output. You can override the default dist output directory if it's being used by something else",
            default: 'dist',
        }),
    };
    sanitizeFlags(flags) {
        return {
            platform: flags.platform,
            branch: flags.branch,
            nonInteractive: flags.nonInteractive,
            channel: flags.channel,
            outputDir: flags.outputDir,
        };
    }
    async run() {
        const credentials = (0, auth_1.retrieveExpoCredentials)();
        if (!credentials.token && !credentials.sessionSecret) {
            log_1.default.error('You are not logged to eas, please run `eas login`');
            process.exit(1);
        }
        const { flags } = await this.parse(Publish);
        const { platform, nonInteractive, branch, channel, outputDir } = this.sanitizeFlags(flags);
        if (!branch) {
            log_1.default.error('Branch name is required');
            process.exit(1);
        }
        if (!channel) {
            log_1.default.error('Channel name is required');
            process.exit(1);
        }
        const vcsClient = (0, vcs_1.resolveVcsClient)(true);
        await vcsClient.ensureRepoExistsAsync();
        const commitHash = await vcsClient.getCommitHashAsync();
        await (0, repo_1.ensureRepoIsCleanAsync)(vcsClient, nonInteractive);
        const projectDir = process.cwd();
        const hasExpo = (0, package_1.isExpoInstalled)(projectDir);
        if (!hasExpo) {
            log_1.default.error('Expo is not installed in this project. Please install Expo first.');
            process.exit(1);
        }
        const privateConfig = await (0, expoConfig_1.getPrivateExpoConfigAsync)(projectDir, {
            env: {
                RELEASE_CHANNEL: channel,
            },
        });
        const updateUrl = (0, expoConfig_1.getExpoConfigUpdateUrl)(privateConfig);
        if (!updateUrl) {
            log_1.default.error("Update url is not setup in your config. Please run 'eoas init' to setup the update url");
            process.exit(1);
        }
        let baseUrl;
        try {
            const parsedUrl = new URL(updateUrl);
            baseUrl = parsedUrl.origin;
        }
        catch (e) {
            log_1.default.error('Invalid URL', e);
            process.exit(1);
        }
        if (!nonInteractive) {
            const confirmed = await (0, prompts_1.confirmAsync)({
                message: `Is this the correct URL of your self-hosted update server? ${baseUrl}`,
                name: 'export',
                type: 'confirm',
            });
            if (!confirmed) {
                log_1.default.error('Please run `eoas init` to setup the correct update url');
                process.exit(1);
            }
        }
        const runtimeSpinner = (0, ora_1.ora)('🔄 Resolving runtime version...').start();
        const runtimeVersions = [
            ...(!platform || platform === expoConfig_1.RequestedPlatform.All || platform === expoConfig_1.RequestedPlatform.Ios
                ? [
                    {
                        runtimeVersion: (await (0, runtimeVersion_1.resolveRuntimeVersionAsync)({
                            exp: privateConfig,
                            platform: 'ios',
                            workflow: await (0, workflow_1.resolveWorkflowAsync)(projectDir, eas_build_job_1.Platform.IOS, vcsClient),
                            projectDir,
                            env: {
                                RELEASE_CHANNEL: channel,
                            },
                        }))?.runtimeVersion,
                        platform: 'ios',
                    },
                ]
                : []),
            ...(!platform || platform === expoConfig_1.RequestedPlatform.All || platform === expoConfig_1.RequestedPlatform.Android
                ? [
                    {
                        runtimeVersion: (await (0, runtimeVersion_1.resolveRuntimeVersionAsync)({
                            exp: privateConfig,
                            platform: 'android',
                            workflow: await (0, workflow_1.resolveWorkflowAsync)(projectDir, eas_build_job_1.Platform.ANDROID, vcsClient),
                            projectDir,
                            env: {
                                RELEASE_CHANNEL: channel,
                            },
                        }))?.runtimeVersion,
                        platform: 'android',
                    },
                ]
                : []),
        ].filter(({ runtimeVersion }) => !!runtimeVersion);
        if (!runtimeVersions.length) {
            runtimeSpinner.fail('Could not resolve runtime versions for the requested platforms');
            log_1.default.error('Could not resolve runtime versions for the requested platforms');
            process.exit(1);
        }
        runtimeSpinner.succeed('✅ Runtime versions resolved');
        const exportSpinner = (0, ora_1.ora)('📦 Exporting project files...').start();
        try {
            await (0, spawn_async_1.default)('rm', ['-rf', outputDir], { cwd: projectDir });
            const { stdout } = await (0, spawn_async_1.default)('npx', ['expo', 'export', '--output-dir', outputDir], {
                cwd: projectDir,
                env: {
                    ...process.env,
                    EXPO_NO_DOTENV: '1',
                },
            });
            exportSpinner.succeed('🚀 Project exported successfully');
            log_1.default.withInfo(stdout);
        }
        catch (e) {
            exportSpinner.fail(`❌ Failed to export the project, ${e}`);
            process.exit(1);
        }
        const publicConfig = await (0, expoConfig_1.getPublicExpoConfigAsync)(projectDir, {
            skipSDKVersionRequirement: true,
        });
        if (!publicConfig) {
            log_1.default.error('Could not find Expo config in this project. Please make sure you have an Expo config.');
            process.exit(1);
        }
        // eslint-disable-next-line
        fs_extra_1.default.writeJsonSync(path_1.default.join(projectDir, outputDir, 'expoConfig.json'), publicConfig, {
            spaces: 2,
        });
        log_1.default.withInfo(`expoConfig.json file created in ${outputDir} directory`);
        const uploadFilesSpinner = (0, ora_1.ora)('📤 Uploading files...').start();
        const files = (0, assets_1.computeFilesRequests)(projectDir, outputDir, platform || expoConfig_1.RequestedPlatform.All);
        if (!files.length) {
            uploadFilesSpinner.fail('No files to upload');
            process.exit(1);
        }
        let uploadUrls = [];
        try {
            uploadUrls = await Promise.all(runtimeVersions.map(async ({ runtimeVersion, platform }) => {
                if (!runtimeVersion) {
                    throw new Error('Runtime version is not resolved');
                }
                return {
                    ...(await (0, assets_1.requestUploadUrls)({
                        body: {
                            fileNames: files.map(file => file.path),
                        },
                        requestUploadUrl: `${baseUrl}/requestUploadUrl/${branch}`,
                        auth: credentials,
                        runtimeVersion,
                        platform,
                        commitHash,
                    })),
                    runtimeVersion,
                    platform,
                };
            }));
            const allItems = uploadUrls.flatMap(({ uploadRequests }) => uploadRequests);
            await Promise.all(allItems.map(async (itm) => {
                const isLocalBucketFileUpload = itm.requestUploadUrl.startsWith(`${baseUrl}/uploadLocalFile`);
                const formData = new form_data_1.default();
                let file;
                try {
                    file = fs_extra_1.default.createReadStream(path_1.default.join(projectDir, outputDir, itm.filePath));
                }
                catch {
                    throw new Error(`Failed to read file ${itm.filePath}`);
                }
                formData.append(itm.fileName, file);
                if (isLocalBucketFileUpload) {
                    const response = await (0, fetch_1.fetchWithRetries)(itm.requestUploadUrl, {
                        method: 'PUT',
                        headers: {
                            ...formData.getHeaders(),
                            ...(0, auth_1.getAuthExpoHeaders)(credentials),
                        },
                        body: formData,
                    });
                    if (!response.ok) {
                        log_1.default.error('Failed to upload file', await response.text());
                        throw new Error('Failed to upload file');
                    }
                    file.close();
                    return;
                }
                const findFile = files.find(f => f.path === itm.filePath || f.name === itm.fileName);
                if (!findFile) {
                    log_1.default.error(`File ${itm.filePath} not found`);
                    throw new Error(`File ${itm.filePath} not found`);
                }
                let contentType = mime_1.default.getType(findFile.ext);
                if (!contentType) {
                    contentType = 'application/octet-stream';
                }
                const buffer = await fs_extra_1.default.readFile(path_1.default.join(projectDir, outputDir, itm.filePath));
                const response = await (0, fetch_1.fetchWithRetries)(itm.requestUploadUrl, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': contentType,
                        'Cache-Control': 'max-age=31556926',
                    },
                    body: buffer,
                });
                if (!response.ok) {
                    log_1.default.error('❌ File upload failed', await response.text());
                    process.exit(1);
                }
                file.close();
            }));
            uploadFilesSpinner.succeed('✅ Files uploaded successfully');
        }
        catch (e) {
            uploadFilesSpinner.fail('❌ Failed to upload static files');
            log_1.default.error(e);
            process.exit(1);
        }
        const markAsFinishedSpinner = (0, ora_1.ora)('🔗 Marking the updates as finished...').start();
        const results = await Promise.all(uploadUrls.map(async ({ updateId, platform, runtimeVersion }) => {
            const response = await (0, fetch_1.fetchWithRetries)(`${baseUrl}/markUpdateAsUploaded/${branch}?platform=${platform}&updateId=${updateId}&runtimeVersion=${runtimeVersion}`, {
                method: 'POST',
                headers: {
                    ...(0, auth_1.getAuthExpoHeaders)(credentials),
                    'Content-Type': 'application/json',
                },
            });
            // If success and status code = 200
            if (response.ok) {
                log_1.default.withInfo(`✅ Update ready for ${platform}`);
                return 'deployed';
            }
            // If response.status === 406 duplicate update
            if (response.status === 406) {
                log_1.default.withInfo(`⚠️ There is no change in the update for ${platform}, ignored...`);
                return 'identical';
            }
            log_1.default.error('❌ Failed to mark the update as finished for platform', platform);
            log_1.default.newLine();
            log_1.default.error(await response.text());
            return 'error';
        }));
        const erroredUpdates = results.filter(result => result === 'error');
        const hasSuccess = results.some(result => result === 'deployed');
        const allIdentical = results.every(result => result === 'identical');
        if (allIdentical) {
            markAsFinishedSpinner.warn('⚠️ No changes found in the update, nothing to deploy');
            return;
        }
        if (erroredUpdates.length) {
            markAsFinishedSpinner.fail('❌ Some errors occurred while marking updates as finished');
            throw new Error();
        }
        else {
            markAsFinishedSpinner.succeed(`\n✅ Your update has been successfully pushed to ${updateUrl}`);
        }
        if (hasSuccess) {
            log_1.default.withInfo(`🔗 Channel: \`${channel}\``);
            log_1.default.withInfo(`🌿 Branch: \`${branch}\``);
            log_1.default.withInfo(`⏳ Deployed at: \`${new Date().toUTCString()}\`\n`);
            log_1.default.withInfo('🔥 Your users will receive the latest update automatically!');
        }
    }
}
exports.default = Publish;
