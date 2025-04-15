"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.isValidUpdateUrl = void 0;
function isValidUpdateUrl(updateUrl) {
    return updateUrl.match(/^https?:\/\/[^/]+$/) !== null;
}
exports.isValidUpdateUrl = isValidUpdateUrl;
