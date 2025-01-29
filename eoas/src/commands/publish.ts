import { Platform } from '@expo/eas-build-job';
import spawnAsync from '@expo/spawn-async';
import { Command, Flags } from '@oclif/core';
import { Config } from '@oclif/core/lib/config';
import FormData from 'form-data';
import fs from 'fs-extra';
import mime from 'mime';
import fetch from 'node-fetch';
import path from 'path';

import { computeFilesRequests, requestUploadUrls } from '../lib/assets';
import { getAuthExpoHeaders, retrieveExpoCredentials } from '../lib/auth';
import {
  RequestedPlatform,
  getExpoConfigUpdateUrl,
  getPrivateExpoConfigAsync,
  getPublicExpoConfigAsync,
} from '../lib/expoConfig';
import Log from '../lib/log';
import { ora } from '../lib/ora';
import { isExpoInstalled } from '../lib/package';
import { confirmAsync } from '../lib/prompts';
import { ensureRepoIsCleanAsync } from '../lib/repo';
import { resolveRuntimeVersionAsync } from '../lib/runtimeVersion';
import { resolveVcsClient } from '../lib/vcs';
import { Client } from '../lib/vcs/vcs';
import { resolveWorkflowAsync } from '../lib/workflow';

export default class Publish extends Command {
  vcsClient: Client;
  constructor(argv: string[], config: Config) {
    super(argv, config);
    this.vcsClient = resolveVcsClient(false);
  }
  static override args = {};
  static override description = 'Publish a new update to the self-hosted update server';
  static override examples = ['<%= config.bin %> <%= command.id %>'];
  static override flags = {
    platform: Flags.string({
      type: 'option',
      options: Object.values(RequestedPlatform),
      default: RequestedPlatform.All,
      required: false,
    }),
    branch: Flags.string({
      description: 'Name of the branch to point to',
      required: true,
    }),
    nonInteractive: Flags.boolean({
      description: 'Run command in non-interactive mode',
      default: false,
    }),
  };
  private sanitizeFlags(flags: any): {
    platform: RequestedPlatform;
    branch: string;
    nonInteractive: boolean;
  } {
    return {
      platform: flags.platform,
      branch: flags.branch,
      nonInteractive: flags.nonInteractive,
    };
  }
  public async run(): Promise<void> {
    const credentials = retrieveExpoCredentials();

    if (!credentials.token && !credentials.sessionSecret) {
      Log.error('You are not logged to eas, please run `eas login`');
      return;
    }
    const { flags } = await this.parse(Publish);
    const { platform, nonInteractive, branch } = this.sanitizeFlags(flags);
    if (!branch) {
      Log.error('Branch name is required');
      return;
    }
    await this.vcsClient.ensureRepoExistsAsync();
    await ensureRepoIsCleanAsync(this.vcsClient, nonInteractive);
    const projectDir = process.cwd();
    const hasExpo = isExpoInstalled(projectDir);
    if (!hasExpo) {
      Log.error('Expo is not installed in this project. Please install Expo first.');
      return;
    }

    const privateConfig = await getPrivateExpoConfigAsync(projectDir);
    const updateUrl = getExpoConfigUpdateUrl(privateConfig);
    if (!updateUrl) {
      Log.error(
        "Update url is not setup in your config. Please run 'eoas init' to setup the update url"
      );
      return;
    }
    let baseUrl: string;
    try {
      const parsedUrl = new URL(updateUrl);
      baseUrl = parsedUrl.origin;
    } catch (e) {
      Log.error('Invalid URL', e);
      return;
    }
    if (!nonInteractive) {
      const confirmed = await confirmAsync({
        message: `Is this the correct URL of your self-hosted update server? ${baseUrl}`,
        name: 'export',
        type: 'confirm',
      });
      if (!confirmed) {
        Log.error('Please run `eoas init` to setup the correct update url');
      }
    }
    const runtimeSpinner = ora('Resolving runtime version').start();
    const runtimeVersions = [
      ...(!platform || platform === RequestedPlatform.All || platform === RequestedPlatform.Ios
        ? [
            (
              await resolveRuntimeVersionAsync({
                exp: privateConfig,
                platform: 'ios',
                workflow: await resolveWorkflowAsync(projectDir, Platform.IOS, this.vcsClient),
                projectDir,
                env: undefined,
              })
            )?.runtimeVersion,
          ]
        : []),
      ...(!platform || platform === RequestedPlatform.All || platform === RequestedPlatform.Android
        ? [
            (
              await resolveRuntimeVersionAsync({
                exp: privateConfig,
                platform: 'android',
                workflow: await resolveWorkflowAsync(projectDir, Platform.ANDROID, this.vcsClient),
                projectDir,
                env: undefined,
              })
            )?.runtimeVersion,
          ]
        : []),
    ].filter(Boolean);
    if (!runtimeVersions.length) {
      runtimeSpinner.fail('Could not resolve runtime versions for the requested platforms');
      Log.error('Could not resolve runtime versions for the requested platforms');
      return;
    }
    runtimeSpinner.succeed('Runtime versions resolved');

    const exportSpinner = ora("Exporting project's static files").start();
    try {
      await spawnAsync('rm', ['-rf', 'dist'], { cwd: projectDir });
      const { stdout } = await spawnAsync('npx', ['expo', 'export', '--output-dir', 'dist'], {
        cwd: projectDir,
        env: {
          ...process.env,
          EXPO_NO_DOTENV: '1',
        },
      });
      exportSpinner.succeed('Project exported successfully');
      Log.withInfo(stdout);
    } catch (err: any) {
      exportSpinner.fail('Failed to export the project');
    }
    const publicConfig = await getPublicExpoConfigAsync(projectDir, {
      skipSDKVersionRequirement: true,
    });
    if (!publicConfig) {
      Log.error(
        'Could not find Expo config in this project. Please make sure you have an Expo config.'
      );
      return;
    }
    // eslint-disable-next-line
    fs.writeJsonSync(path.join(projectDir, 'dist', 'expoConfig.json'), publicConfig, {
      spaces: 2,
    });
    Log.withInfo('expoConfig.json file created in dist directory');
    const uploadFilesSpinner = ora('Uploading files to the server').start();
    const files = computeFilesRequests(projectDir, platform || RequestedPlatform.All);
    if (!files.length) {
      uploadFilesSpinner.fail('No files to upload');
    }
    try {
      const uploadUrls = await Promise.all(
        runtimeVersions.map(runtimeVersion => {
          if (!runtimeVersion) {
            throw new Error('Runtime version is not resolved');
          }
          return requestUploadUrls(
            {
              fileNames: files.map(file => file.path),
            },
            `${baseUrl}/requestUploadUrl/${branch}`,
            credentials,
            runtimeVersion
          );
        })
      );
      const allItems = uploadUrls.flat();
      await Promise.all(
        allItems.map(async itm => {
          const isLocalBucketFileUpload = itm.requestUploadUrl.startsWith(
            `${baseUrl}/uploadLocalFile`
          );
          const formData = new FormData();
          let file: fs.ReadStream;
          try {
            file = fs.createReadStream(path.join(projectDir, 'dist', itm.filePath));
          } catch {
            throw new Error(`Failed to read file ${itm.filePath}`);
          }
          formData.append(itm.fileName, file);
          if (isLocalBucketFileUpload) {
            const response = await fetch(itm.requestUploadUrl, {
              method: 'PUT',
              headers: {
                ...formData.getHeaders(),
                ...getAuthExpoHeaders(credentials),
              },
              body: formData,
            });
            if (!response.ok) {
              Log.error('Failed to upload file', await response.text());
              throw new Error('Failed to upload file');
            }
            file.close();
            return;
          }
          const findFile = files.find(f => f.path === itm.filePath || f.name === itm.fileName);
          if (!findFile) {
            Log.error(`File ${itm.filePath} not found`);
            throw new Error(`File ${itm.filePath} not found`);
          }
          let contentType = mime.getType(findFile.ext);
          if (!contentType) {
            contentType = 'application/octet-stream';
          }
          const buffer = await fs.readFile(path.join(projectDir, 'dist', itm.filePath));
          const response = await fetch(itm.requestUploadUrl, {
            method: 'PUT',
            headers: {
              'Content-Type': contentType,
              'Cache-Control': 'max-age=31556926',
            },
            body: buffer,
          });
          if (!response.ok) {
            Log.error('Failed to upload file', await response.text());
            throw new Error('Failed to upload file');
          }
          file.close();
        })
      );
      uploadFilesSpinner.succeed('Files uploaded successfully');
    } catch {
      uploadFilesSpinner.fail('Failed to upload static files');
    }
  }
}
