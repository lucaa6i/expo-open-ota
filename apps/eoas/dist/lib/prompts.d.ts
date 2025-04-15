import prompts, { Answers, Choice, Options } from 'prompts';
export interface ExpoChoice<T> extends Choice {
    value: T;
}
export declare function promptAsync<T extends string = string>(questions: prompts.PromptObject<T> | prompts.PromptObject<T>[], options?: Options): Promise<Answers<T>>;
export declare function confirmAsync(question: prompts.PromptObject<any>, options?: Options): Promise<boolean>;
export declare function selectAsync<T>(message: string, choices: ExpoChoice<T>[], config?: {
    options?: Options;
    initial?: T;
    warningMessageForDisabledEntries?: string;
}): Promise<T>;
export declare function toggleConfirmAsync(questions: prompts.PromptObject<any>, options?: Options): Promise<boolean>;
export declare function pressAnyKeyToContinueAsync(): Promise<void>;
