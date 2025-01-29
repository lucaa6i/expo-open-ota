export function isValidUpdateUrl(updateUrl: string): boolean {
  return updateUrl.match(/^https?:\/\/[^/]+$/) !== null;
}
