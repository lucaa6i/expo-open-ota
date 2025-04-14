"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.isExpoInstalled = void 0;
const config_1 = require("@expo/config");
function isExpoInstalled(projectDir) {
    const packageJson = (0, config_1.getPackageJson)(projectDir);
    return !!(packageJson.dependencies && 'expo' in packageJson.dependencies);
}
exports.isExpoInstalled = isExpoInstalled;
