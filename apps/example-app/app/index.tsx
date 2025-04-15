import {
  Platform,
  SafeAreaView,
  ScrollView,
  Alert,
  StyleSheet,
  Button,
} from 'react-native'
import * as Updates from 'expo-updates'
import { Picker } from '@react-native-picker/picker'
import { ThemedText } from '@/components/ThemedText'
import { ThemedView } from '@/components/ThemedView'
import Constants from 'expo-constants/src/Constants'
import { useState } from 'react'

const RELEASE_CHANNELS = ['production', 'staging']

export default function HomeScreen() {
  const [loading, load] = useState<boolean>(false)

  const onSelectReleaseChannel = async (channel: string) => {
    if (channel === Updates.channel) return
    if (__DEV__ || loading || Platform.OS === 'web') {
      return
    }
    Updates.setUpdateURLAndRequestHeadersOverride({
      updateUrl: Constants.expoConfig?.updates?.url as string,
      requestHeaders: {
        'expo-channel-name': channel,
      },
    })
    await checkUpdates()
  }

  const checkUpdates = async () => {
    if (__DEV__ || loading || Platform.OS === 'web') {
      return
    }
    try {
      const update = await Updates.checkForUpdateAsync()
      if (update.isAvailable) {
        load(true)
        await Updates.fetchUpdateAsync()
        return Updates.reloadAsync()
      } else {
        load(false)
        Alert.alert(
          'Update not available',
          'There is no new update available.',
          [
            {
              text: 'OK',
              style: 'cancel',
            },
          ],
          { cancelable: false },
        )
      }
    } catch (e) {
      load(false)
    }
  }

  return (
    <SafeAreaView style={styles.safeAreaView}>
      <ScrollView contentContainerStyle={styles.scrollView}>
        <ThemedView style={styles.titleContainer}>
          <ThemedText type="title">Current update</ThemedText>
        </ThemedView>
        <ThemedView style={styles.informations}>
          <ThemedText>Update ID: {Updates.updateId}</ThemedText>
          <ThemedText>Runtime version: {Updates.runtimeVersion}</ThemedText>
          <ThemedText>
            Release channel: {process.env.RELEASE_CHANNEL}
          </ThemedText>
          <ThemedText>
            Update server url : {Constants.expoConfig?.updates?.url || ''}
          </ThemedText>
        </ThemedView>
        <ThemedView>
          <Picker
            selectedValue={Updates.channel || undefined}
            onValueChange={(val: string) => {
              if (!val) return
              return onSelectReleaseChannel(val)
            }}
          >
            {RELEASE_CHANNELS.map(channel => (
              <Picker.Item
                key={channel}
                label={channel}
                value={channel}
                testID={`release-channel-${channel}`}
              />
            ))}
          </Picker>
          <Button
            title="Check for updates"
            onPress={() => checkUpdates()}
            disabled={loading}
          />
        </ThemedView>
      </ScrollView>
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  titleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  informations: {
    gap: 8,
    marginBottom: 8,
  },
  safeAreaView: {
    flex: 1,
    backgroundColor: '#fff',
  },
  scrollView: {
    flexGrow: 1,
    paddingVertical: 16,
    paddingHorizontal: 16,
  },
})
