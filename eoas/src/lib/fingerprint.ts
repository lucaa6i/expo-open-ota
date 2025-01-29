import { Env, Workflow } from '@expo/eas-build-job';
import { silent as silentResolveFrom } from 'resolve-from';

import Log from './log';
import { ora } from './ora';

export interface HashSourceFile {
  type: 'file';
  filePath: string;
  reasons: string[];
}

export interface HashSourceDir {
  type: 'dir';
  filePath: string;
  reasons: string[];
}

export interface HashSourceContents {
  type: 'contents';
  id: string;
  contents: string | Buffer;
  reasons: string[];
}

export type HashSource = HashSourceFile | HashSourceDir | HashSourceContents;

export type FingerprintSource = HashSource & {
  hash: string | null;
  debugInfo?: DebugInfo;
};

export interface Fingerprint {
  sources: FingerprintSource[];
  hash: string;
}

export interface DebugInfoFile {
  path: string;
  hash: string;
}

export interface DebugInfoDir {
  path: string;
  hash: string;
  children: (DebugInfoFile | DebugInfoDir | undefined)[];
}

export interface DebugInfoContents {
  hash: string;
}

export type DebugInfo = DebugInfoFile | DebugInfoDir | DebugInfoContents;

export type FingerprintOptions = {
  workflow: Workflow;
  platforms: string[];
  debug?: boolean;
  env: Env | undefined;
  cwd?: string;
};

async function createFingerprintWithoutLoggingAsync(
  projectDir: string,
  fingerprintPath: string,
  options: FingerprintOptions
): Promise<
  Fingerprint & {
    isDebugSource: boolean;
  }
> {
  const Fingerprint = require(fingerprintPath);
  const fingerprintOptions: Record<string, any> = {};
  if (options.platforms) {
    fingerprintOptions.platforms = [...options.platforms];
  }
  if (options.workflow === Workflow.MANAGED) {
    fingerprintOptions.ignorePaths = ['android/**/*', 'ios/**/*'];
  }
  if (options.debug) {
    fingerprintOptions.debug = true;
  }
  console.log('fingerprintPath', fingerprintPath, Fingerprint);
  // eslint-disable-next-line @typescript-eslint/return-await
  return await Fingerprint.createFingerprintAsync(projectDir, fingerprintOptions);
}

export async function createFingerprintAsync(
  projectDir: string,
  options: FingerprintOptions
): Promise<
  | (Fingerprint & {
      isDebugSource: boolean;
    })
  | null
> {
  // @expo/fingerprint is exported in the expo package for SDK 52+
  const fingerprintPath = silentResolveFrom(projectDir, '@expo/fingerprint');
  console.log('fingerprintPath', fingerprintPath);
  if (!fingerprintPath) {
    return null;
  }

  if (process.env.EAS_SKIP_AUTO_FINGERPRINT) {
    Log.log('Skipping project fingerprint');
    return null;
  }

  const timeoutId = setTimeout(() => {
    Log.log('⌛️ Computing the project fingerprint is taking longer than expected...');
    Log.log('⏩ To skip this step, set the environment variable: EAS_SKIP_AUTO_FINGERPRINT=1');
  }, 5000);

  const spinner = ora(`Computing project fingerprint`).start();
  try {
    const fingerprint = await createFingerprintWithoutLoggingAsync(
      projectDir,
      fingerprintPath,
      options
    );
    spinner.succeed(`Computed project fingerprint`);
    return fingerprint;
  } catch (e) {
    spinner.fail(`Failed to compute project fingerprint`);
    Log.log('⏩ To skip this step, set the environment variable: EAS_SKIP_AUTO_FINGERPRINT=1');
    throw e;
  } finally {
    // Clear the timeout if the operation finishes before the time limit
    clearTimeout(timeoutId);
    spinner.stop();
  }
}
