import { writable } from 'svelte/store'
import { config } from '../../../wailsjs/go/models'
import type { Settings } from '../types'
import { GetSettings, SaveSettings, OpenHFTokenPage } from '../../../wailsjs/go/main/App'

export const settingsStore = writable<Settings | null>(null)

export async function fetchSettings() {
  try {
    const s = await GetSettings()
    settingsStore.set(s)
    return s
  } catch (err) {
    console.error('Failed to fetch settings:', err)
  }
}

export async function persistSettings(updated: any) {
  try {
    const s = config.Settings.createFrom(updated)
    await SaveSettings(s)
    settingsStore.set(s)
  } catch (err) {
    console.error('Failed to save settings:', err)
    throw err
  }
}

export async function openTokenPage() {
  try {
    await OpenHFTokenPage()
  } catch (err) {
    console.error('Failed to open HF token page:', err)
  }
}
