"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const core_1 = require("@oclif/core");
const fs_extra_1 = tslib_1.__importDefault(require("fs-extra"));
const path_1 = tslib_1.__importDefault(require("path"));
const expoConfig_1 = require("../lib/expoConfig");
const log_1 = tslib_1.__importDefault(require("../lib/log"));
const ora_1 = require("../lib/ora");
const package_1 = require("../lib/package");
const prompts_1 = require("../lib/prompts");
const utils_1 = require("../lib/utils");
class Init extends core_1.Command {
    static args = {};
    static description = 'Configure your existing expo project with Expo Open OTA';
    static examples = ['<%= config.bin %> <%= command.id %>'];
    static flags = {};
    async run() {
        const projectDir = process.cwd();
        const hasExpo = (0, package_1.isExpoInstalled)(projectDir);
        if (!hasExpo) {
            log_1.default.error('Expo is not installed in this project. Please install Expo first.');
            return;
        }
        const config = await (0, expoConfig_1.getPrivateExpoConfigAsync)(projectDir);
        if (!config) {
            log_1.default.error('Could not find Expo config in this project. Please make sure you have an Expo config.');
            return;
        }
        const { updateUrl: promptedUrl } = await (0, prompts_1.promptAsync)({
            message: 'Enter the URL of your update server (ex: https://customota.com)',
            name: 'updateUrl',
            type: 'text',
            initial: (0, expoConfig_1.getExpoConfigUpdateUrl)(config),
            validate: v => {
                return !!v && (0, utils_1.isValidUpdateUrl)(v);
            },
        });
        let manifestEndpoint = `${promptedUrl}/manifest`;
        const updateUrl = (0, expoConfig_1.getExpoConfigUpdateUrl)(config);
        if (updateUrl && !updateUrl.includes('expo.dev')) {
            const confirmed = await (0, prompts_1.confirmAsync)({
                message: `Expo config already has an update URL set to ${updateUrl}. Do you want to replace it?`,
                name: 'replace',
                type: 'confirm',
            });
            if (!confirmed) {
                manifestEndpoint = updateUrl;
            }
        }
        const confirmed = await (0, prompts_1.confirmAsync)({
            message: 'Do you have already generated your certificates for code signing?',
            name: 'certificates',
            type: 'confirm',
        });
        if (!confirmed) {
            log_1.default.fail('You need to generate your certificates first by using npx eoas generate-certs');
            return;
        }
        const { codeSigningCertificatePath } = await (0, prompts_1.promptAsync)({
            message: 'Enter the path to your code signing certificate (ex: ./certs/certificate.pem)',
            name: 'codeSigningCertificatePath',
            type: 'text',
            initial: './certs/certificate.pem',
            validate: v => {
                try {
                    const fullPath = path_1.default.resolve(projectDir, v);
                    // eslint-disable-next-line
                    const fileExists = fs_extra_1.default.existsSync(fullPath);
                    if (!fileExists) {
                        log_1.default.newLine();
                        log_1.default.error('File does not exist');
                        return false;
                    }
                    // eslint-disable-next-line
                    const key = fs_extra_1.default.readFileSync(fullPath, 'utf8');
                    if (!key) {
                        log_1.default.error('Empty key');
                        return false;
                    }
                    return true;
                }
                catch {
                    return false;
                }
            },
        });
        const newUpdateConfig = {
            url: manifestEndpoint,
            codeSigningMetadata: {
                keyid: 'main',
                alg: 'rsa-v1_5-sha256',
            },
            codeSigningCertificate: codeSigningCertificatePath,
            enabled: true,
            requestHeaders: {
                'expo-channel-name': 'process.env.RELEASE_CHANNEL',
            },
        };
        const updateConfigSpinner = (0, ora_1.ora)('Updating Expo config').start();
        try {
            await (0, expoConfig_1.createOrModifyExpoConfigAsync)(projectDir, {
                updates: newUpdateConfig,
            });
            updateConfigSpinner.succeed('Expo config successfully updated do not forget to format the file with prettier or eslint');
        }
        catch (e) {
            updateConfigSpinner.fail('Failed to update Expo config');
            log_1.default.error(e);
        }
    }
}
exports.default = Init;
