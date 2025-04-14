"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getAuthExpoHeaders = exports.retrieveExpoCredentials = void 0;
const tslib_1 = require("tslib");
const os_1 = require("os");
const path_1 = tslib_1.__importDefault(require("path"));
function dotExpoHomeDirectory() {
    const home = (0, os_1.homedir)();
    if (!home) {
        throw new Error("Can't determine your home directory; make sure your $HOME environment variable is set.");
    }
    let dirPath;
    if (process.env.EXPO_STAGING) {
        dirPath = path_1.default.join(home, '.expo-staging');
    }
    else if (process.env.EXPO_LOCAL) {
        dirPath = path_1.default.join(home, '.expo-local');
    }
    else {
        dirPath = path_1.default.join(home, '.expo');
    }
    return dirPath;
}
function getStateJsonPath() {
    return path_1.default.join(dotExpoHomeDirectory(), 'state.json');
}
function getExpoSessionData() {
    try {
        const stateJsonPath = getStateJsonPath();
        const stateJson = require(stateJsonPath);
        return stateJson['auth'] || null;
    }
    catch {
        return null;
    }
}
function retrieveExpoCredentials() {
    const token = process.env.EXPO_TOKEN;
    const sessionData = getExpoSessionData();
    const sessionSecret = sessionData?.sessionSecret;
    return { token, sessionSecret };
}
exports.retrieveExpoCredentials = retrieveExpoCredentials;
function getAuthExpoHeaders(credentials) {
    if (credentials.token) {
        return {
            Authorization: `Bearer ${credentials.token}`,
        };
    }
    if (credentials.sessionSecret) {
        return {
            'expo-session': credentials.sessionSecret,
        };
    }
    return {};
}
exports.getAuthExpoHeaders = getAuthExpoHeaders;
