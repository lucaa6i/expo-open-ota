import { ExpoConfig } from '@expo/config-types'
import { ConfigContext } from '@expo/config'

export default ({ config }: ConfigContext): ExpoConfig => {
  return {
    ...(config as ExpoConfig),
    runtimeVersion: '1.0.0',
    updates: {
      url: 'http://localhost:3000/manifest',
      codeSigningMetadata: {
        keyid: 'main',
        alg: 'rsa-v1_5-sha256',
      },
      codeSigningCertificate: './certs/certificate-dev.pem',
      enabled: true,
      disableAntiBrickingMeasures: true,
      requestHeaders: {
        'expo-channel-name': process.env.RELEASE_CHANNEL,
      },
    },
  }
}
