"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.pressAnyKeyToContinueAsync = exports.toggleConfirmAsync = exports.selectAsync = exports.confirmAsync = exports.promptAsync = void 0;
const tslib_1 = require("tslib");
// This file is copied from eas-cli[https://github.com/expo/eas-cli] to ensure consistent user experience across the CLI.
const os_1 = require("os");
const prompts_1 = tslib_1.__importDefault(require("prompts"));
async function promptAsync(questions, options = {}) {
    if (!process.stdin.isTTY) {
        const message = Array.isArray(questions) ? questions[0]?.message : questions.message;
        throw new Error(`Input is required, but stdin is not readable. Failed to display prompt: ${message}`);
    }
    return await (0, prompts_1.default)(questions, {
        onCancel() {
            process.exit(os_1.constants.signals.SIGINT + 128); // Exit code 130 used when process is interrupted with ctrl+c.
        },
        ...options,
    });
}
exports.promptAsync = promptAsync;
async function confirmAsync(question, options) {
    const { value } = await promptAsync({
        initial: true,
        ...question,
        name: 'value',
        type: 'confirm',
    }, options);
    return value;
}
exports.confirmAsync = confirmAsync;
async function selectAsync(message, choices, config) {
    const initial = config?.initial ? choices.findIndex(({ value }) => value === config.initial) : 0;
    const { value } = await promptAsync({
        message,
        choices,
        initial,
        name: 'value',
        type: 'select',
        warn: config?.warningMessageForDisabledEntries,
    }, config?.options ?? {});
    return value ?? null;
}
exports.selectAsync = selectAsync;
async function toggleConfirmAsync(questions, options) {
    const { value } = await promptAsync({
        active: 'yes',
        inactive: 'no',
        ...questions,
        name: 'value',
        type: 'toggle',
    }, options);
    return value ?? null;
}
exports.toggleConfirmAsync = toggleConfirmAsync;
async function pressAnyKeyToContinueAsync() {
    process.stdin.setRawMode(true);
    process.stdin.resume();
    process.stdin.setEncoding('utf8');
    await new Promise(res => {
        process.stdin.on('data', key => {
            if (String(key) === '\u0003') {
                process.exit(os_1.constants.signals.SIGINT + 128); // ctrl-c
            }
            res();
        });
    });
}
exports.pressAnyKeyToContinueAsync = pressAnyKeyToContinueAsync;
