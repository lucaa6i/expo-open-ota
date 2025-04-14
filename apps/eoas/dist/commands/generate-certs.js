"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const code_signing_certificates_1 = require("@expo/code-signing-certificates");
const core_1 = require("@oclif/core");
const fs_extra_1 = require("fs-extra");
const path_1 = tslib_1.__importDefault(require("path"));
const log_1 = tslib_1.__importDefault(require("../lib/log"));
const prompts_1 = require("../lib/prompts");
class GenerateCerts extends core_1.Command {
    static args = {};
    static description = 'Generate private & public certificates for code signing';
    static examples = ['<%= config.bin %> <%= command.id %>'];
    static flags = {};
    async run() {
        const { certificateOutputDir } = await (0, prompts_1.promptAsync)({
            message: 'In which directory would you like to store your code signing certificate (used by your expo app)?',
            name: 'certificateOutputDir',
            type: 'text',
            initial: './certs',
            validate: v => {
                try {
                    // eslint-disable-next-line
                    (0, fs_extra_1.ensureDirSync)(path_1.default.join(process.cwd(), v));
                    return true;
                }
                catch {
                    return false;
                }
            },
        });
        const { keyOutputDir } = await (0, prompts_1.promptAsync)({
            message: 'In which directory would you like to store your key pair (used by your OTA Server) ?. ⚠️ Those certss are sensitive and should be kept private.',
            name: 'keyOutputDir',
            type: 'text',
            initial: './certs',
            validate: v => {
                try {
                    // eslint-disable-next-line
                    (0, fs_extra_1.ensureDirSync)(path_1.default.join(process.cwd(), v));
                    return true;
                }
                catch {
                    return false;
                }
            },
        });
        const { certificateCommonName } = await (0, prompts_1.promptAsync)({
            message: 'Please enter your Organization name',
            name: 'certificateCommonName',
            type: 'text',
            initial: 'Your Organization Name',
            validate: v => {
                return !!v;
            },
        });
        const { certificateValidityDurationYears } = await (0, prompts_1.promptAsync)({
            message: 'How many years should the certificate be valid for?',
            name: 'certificateValidityDurationYears',
            type: 'number',
            initial: 10,
            validate: v => {
                return v > 0 && Number.isInteger(v);
            },
        });
        const validityDurationYears = Math.floor(Number(certificateValidityDurationYears));
        const certificateOutput = path_1.default.resolve(process.cwd(), certificateOutputDir);
        const keyOutput = path_1.default.resolve(process.cwd(), keyOutputDir);
        const validityNotBefore = new Date();
        const validityNotAfter = new Date();
        validityNotAfter.setFullYear(validityNotAfter.getFullYear() + validityDurationYears);
        const keyPair = (0, code_signing_certificates_1.generateKeyPair)();
        const certificate = (0, code_signing_certificates_1.generateSelfSignedCodeSigningCertificate)({
            keyPair,
            validityNotBefore,
            validityNotAfter,
            commonName: certificateCommonName,
        });
        const keyPairPEM = (0, code_signing_certificates_1.convertKeyPairToPEM)(keyPair);
        const certificatePEM = (0, code_signing_certificates_1.convertCertificateToCertificatePEM)(certificate);
        await Promise.all([
            (0, fs_extra_1.writeFile)(path_1.default.join(keyOutput, 'public-key.pem'), keyPairPEM.publicKeyPEM),
            (0, fs_extra_1.writeFile)(path_1.default.join(keyOutput, 'private-key.pem'), keyPairPEM.privateKeyPEM),
            (0, fs_extra_1.writeFile)(path_1.default.join(certificateOutput, 'certificate.pem'), certificatePEM),
        ]);
        log_1.default.succeed(`Generated public and private keys output in ${keyOutputDir}. Please follow the documentation to securely store them and do not commit them to your repository.`);
        log_1.default.succeed(`Generated code signing certificate output in ${certificateOutputDir}.`);
    }
}
exports.default = GenerateCerts;
