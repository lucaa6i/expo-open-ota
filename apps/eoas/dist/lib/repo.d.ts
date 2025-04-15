import { Client } from './vcs/vcs';
export declare function commitPromptAsync(vcsClient: Client, { initialCommitMessage, commitAllFiles, }?: {
    initialCommitMessage?: string;
    commitAllFiles?: boolean;
}): Promise<void>;
export declare function ensureRepoIsCleanAsync(vcsClient: Client, nonInteractive?: boolean): Promise<void>;
