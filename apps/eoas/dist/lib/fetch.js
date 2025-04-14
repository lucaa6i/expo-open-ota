"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.fetchWithRetries = void 0;
const tslib_1 = require("tslib");
const fetch_retry_1 = tslib_1.__importDefault(require("fetch-retry"));
const node_fetch_1 = tslib_1.__importDefault(require("node-fetch"));
const log_1 = tslib_1.__importDefault(require("./log"));
const fetch = (0, fetch_retry_1.default)(node_fetch_1.default);
async function fetchWithRetries(url, options) {
    return await fetch(url, {
        ...options,
        retryDelay(attempt) {
            return Math.pow(2, attempt) * 500;
        },
        retryOn: (attempt, error) => {
            if (attempt > 3) {
                return false;
            }
            if (error) {
                log_1.default.warn(`Retry ${attempt} after network error:`, error.message);
                return true;
            }
            return false;
        },
    });
}
exports.fetchWithRetries = fetchWithRetries;
