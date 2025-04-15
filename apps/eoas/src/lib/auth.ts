import { homedir } from 'os';
import path from 'path';

export interface ExpoCredentials {
  token?: string;
  sessionSecret?: string;
}
type SessionData = {
  sessionSecret: string;
  userId: string;
  username: string;
  currentConnection: 'Username-Password-Authentication' | 'Browser-Flow-Authentication';
};

function dotExpoHomeDirectory(): string {
  const home = homedir();
  if (!home) {
    throw new Error(
      "Can't determine your home directory; make sure your $HOME environment variable is set."
    );
  }

  let dirPath;
  if (process.env.EXPO_STAGING) {
    dirPath = path.join(home, '.expo-staging');
  } else if (process.env.EXPO_LOCAL) {
    dirPath = path.join(home, '.expo-local');
  } else {
    dirPath = path.join(home, '.expo');
  }
  return dirPath;
}

function getStateJsonPath(): string {
  return path.join(dotExpoHomeDirectory(), 'state.json');
}

function getExpoSessionData(): SessionData | null {
  try {
    const stateJsonPath = getStateJsonPath();
    const stateJson = require(stateJsonPath);
    return stateJson['auth'] || null;
  } catch {
    return null;
  }
}

export function retrieveExpoCredentials(): ExpoCredentials {
  const token = process.env.EXPO_TOKEN;
  const sessionData = getExpoSessionData();
  const sessionSecret = sessionData?.sessionSecret;
  return { token, sessionSecret };
}

export function getAuthExpoHeaders(credentials: ExpoCredentials): Record<string, string> {
  if (credentials.token) {
    return {
      Authorization: `Bearer ${credentials.token}`,
    };
  }
  if (credentials.sessionSecret) {
    return {
      'expo-session': credentials.sessionSecret,
    };
  }
  return {};
}
