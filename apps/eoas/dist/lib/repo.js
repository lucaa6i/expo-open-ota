"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ensureRepoIsCleanAsync = exports.commitPromptAsync = void 0;
const tslib_1 = require("tslib");
// This file is copied from eas-cli[https://github.com/expo/eas-cli] to ensure consistent user experience across the CLI.
const chalk_1 = tslib_1.__importDefault(require("chalk"));
const log_1 = tslib_1.__importDefault(require("./log"));
const prompts_1 = require("./prompts");
async function commitPromptAsync(vcsClient, { initialCommitMessage, commitAllFiles, } = {}) {
    const { message } = await (0, prompts_1.promptAsync)({
        type: 'text',
        name: 'message',
        message: 'Commit message:',
        initial: initialCommitMessage,
        validate: (input) => input !== '',
    });
    await vcsClient.commitAsync({
        commitAllFiles,
        commitMessage: message,
        nonInteractive: false,
    });
}
exports.commitPromptAsync = commitPromptAsync;
async function ensureRepoIsCleanAsync(vcsClient, nonInteractive = false) {
    if (!(await vcsClient.isCommitRequiredAsync())) {
        return;
    }
    log_1.default.addNewLineIfNone();
    log_1.default.warn(`${chalk_1.default.bold('Warning!')} Your repository working tree is dirty.`);
    log_1.default.log(`This operation needs to be run on a clean working tree. ${chalk_1.default.bold('Commit all your changes before proceeding')}.`);
    if (nonInteractive) {
        log_1.default.log('The following files need to be committed:');
        await vcsClient.showChangedFilesAsync();
        throw new Error('Commit all changes. Aborting...');
    }
    const answer = await (0, prompts_1.confirmAsync)({
        message: `Commit changes to git?`,
        type: 'confirm',
        name: 'confirm git commit',
    });
    if (answer) {
        await commitPromptAsync(vcsClient, { commitAllFiles: true });
    }
    else {
        throw new Error('Commit all changes. Aborting...');
    }
}
exports.ensureRepoIsCleanAsync = ensureRepoIsCleanAsync;
