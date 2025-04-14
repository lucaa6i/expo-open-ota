import { Command } from '@oclif/core';
export default class Publish extends Command {
    static args: {};
    static description: string;
    static examples: string[];
    static flags: {
        platform: import("@oclif/core/lib/interfaces").OptionFlag<string, import("@oclif/core/lib/interfaces").CustomOptions>;
        channel: import("@oclif/core/lib/interfaces").OptionFlag<string, import("@oclif/core/lib/interfaces").CustomOptions>;
        branch: import("@oclif/core/lib/interfaces").OptionFlag<string, import("@oclif/core/lib/interfaces").CustomOptions>;
        nonInteractive: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
        outputDir: import("@oclif/core/lib/interfaces").OptionFlag<string, import("@oclif/core/lib/interfaces").CustomOptions>;
    };
    private sanitizeFlags;
    run(): Promise<void>;
}
