import { Platform } from '@expo/eas-build-job';
import { Command, Flags } from '@oclif/core';

import { getAuthExpoHeaders, retrieveExpoCredentials } from '../lib/auth';
import {
  RequestedPlatform,
  getExpoConfigUpdateUrl,
  getPrivateExpoConfigAsync,
} from '../lib/expoConfig';
import { fetchWithRetries } from '../lib/fetch';
import Log from '../lib/log';
import { ora } from '../lib/ora';
import { isExpoInstalled } from '../lib/package';
import { confirmAsync } from '../lib/prompts';
import { resolveRuntimeVersionAsync } from '../lib/runtimeVersion';
import { resolveVcsClient } from '../lib/vcs';
import { resolveWorkflowAsync } from '../lib/workflow';

export default class Publish extends Command {
  static override args = {};
  static override description = 'Publish a new rollback to the self-hosted update server';
  static override examples = ['<%= config.bin %> <%= command.id %>'];
  static override flags = {
    platform: Flags.string({
      type: 'option',
      options: Object.values(RequestedPlatform),
      default: RequestedPlatform.All,
      required: false,
    }),
    channel: Flags.string({
      description: 'Name of the channel to publish the rollback to',
      required: true,
    }),
    branch: Flags.string({
      description: 'Name of the branch to point to',
      required: true,
    }),
  };
  private sanitizeFlags(flags: any): {
    platform: RequestedPlatform;
    branch: string;
    channel: string;
  } {
    return {
      platform: flags.platform,
      branch: flags.branch,
      channel: flags.channel,
    };
  }
  public async run(): Promise<void> {
    const credentials = retrieveExpoCredentials();
    if (!credentials.token && !credentials.sessionSecret) {
      Log.error('You are not logged to eas, please run `eas login`');
      process.exit(1);
    }
    const { flags } = await this.parse(Publish);
    const { platform, branch, channel } = this.sanitizeFlags(flags);
    if (!branch) {
      Log.error('Branch name is required');
      process.exit(1);
    }
    if (!channel) {
      Log.error('Channel name is required');
      process.exit(1);
    }
    const vcsClient = resolveVcsClient(true);
    await vcsClient.ensureRepoExistsAsync();
    const commitHash = await vcsClient.getCommitHashAsync();
    const projectDir = process.cwd();
    const hasExpo = isExpoInstalled(projectDir);
    if (!hasExpo) {
      Log.error('Expo is not installed in this project. Please install Expo first.');
      process.exit(1);
    }
    const confirmed = await confirmAsync({
      message: `Are you sure you want to publish a rollback to the branch ${branch} ?`,
      name: 'export',
      type: 'confirm',
    });
    if (!confirmed) {
      Log.error('Operation cancelled');
      process.exit(1);
    }

    const privateConfig = await getPrivateExpoConfigAsync(projectDir, {
      env: {
        RELEASE_CHANNEL: channel,
      },
    });
    if (privateConfig?.updates?.disableAntiBrickingMeasures) {
      Log.error(
        'When using disableAntiBrickingMeasures, expo-updates is ignoring the embeded update of the app, please use republish command instead'
      );
      process.exit(1);
    }
    const updateUrl = getExpoConfigUpdateUrl(privateConfig);
    if (!updateUrl) {
      Log.error(
        "Update url is not setup in your config. Please run 'eoas init' to setup the update url"
      );
      process.exit(1);
    }
    let baseUrl: string;
    try {
      const parsedUrl = new URL(updateUrl);
      baseUrl = parsedUrl.origin;
    } catch (e) {
      Log.error('Invalid URL', e);
      process.exit(1);
    }
    const runtimeSpinner = ora('🔄 Resolving runtime version...').start();
    const runtimeVersions = [
      ...(!platform || platform === RequestedPlatform.All || platform === RequestedPlatform.Ios
        ? [
            {
              runtimeVersion: (
                await resolveRuntimeVersionAsync({
                  exp: privateConfig,
                  platform: 'ios',
                  workflow: await resolveWorkflowAsync(projectDir, Platform.IOS, vcsClient),
                  projectDir,
                  env: {
                    RELEASE_CHANNEL: channel,
                  },
                })
              )?.runtimeVersion,
              platform: 'ios',
            },
          ]
        : []),
      ...(!platform || platform === RequestedPlatform.All || platform === RequestedPlatform.Android
        ? [
            {
              runtimeVersion: (
                await resolveRuntimeVersionAsync({
                  exp: privateConfig,
                  platform: 'android',
                  workflow: await resolveWorkflowAsync(projectDir, Platform.ANDROID, vcsClient),
                  projectDir,
                  env: {
                    RELEASE_CHANNEL: channel,
                  },
                })
              )?.runtimeVersion,
              platform: 'android',
            },
          ]
        : []),
    ].filter(({ runtimeVersion }) => !!runtimeVersion);
    if (!runtimeVersions.length) {
      runtimeSpinner.fail('Could not resolve runtime versions for the requested platforms');
      Log.error('Could not resolve runtime versions for the requested platforms');
      process.exit(1);
    }
    runtimeSpinner.succeed('✅ Runtime versions resolved');
    const rollbackSpinner = ora('📦 Uploading rollback...').start();
    const erroredPlatforms: { platform: string; reason: string }[] = [];
    await Promise.all(
      runtimeVersions.map(async ({ runtimeVersion, platform }) => {
        const endpoint = `${baseUrl}/rollback/${branch}?commitHash=${commitHash}&channel=${channel}&platform=${platform}&runtimeVersion=${runtimeVersion}`;
        const response = await fetchWithRetries(endpoint, {
          method: 'POST',
          headers: {
            ...getAuthExpoHeaders(credentials),
          },
        });
        if (!response.ok) {
          erroredPlatforms.push({
            platform,
            reason: await response.text(),
          });
        }
      })
    );
    if (erroredPlatforms.length) {
      rollbackSpinner.fail('❌ Rollback failed');
      erroredPlatforms.forEach(({ platform, reason }) => {
        Log.error(`Failed to publish rollback for ${platform}: ${reason}`);
      });
      process.exit(1);
    } else {
      rollbackSpinner.succeed('✅ Rollback published successfully');
    }
  }
}
