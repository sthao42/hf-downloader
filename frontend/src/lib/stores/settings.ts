import { writable } from 'svelte/store'
import { config } from '../../../wailsjs/go/models'
import type { Settings } from '../types'
import { GetSettings, SaveSettings, OpenHFTokenPage, UpdateRecentPaths } from '../../../wailsjs/go/main/App'

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
    const safeUpdated = {
      ...updated,
      speedLimitMBps: Math.max(0, Number(updated.speedLimitMBps) || 0),
      maxConcurrentFiles: Math.max(1, Number(updated.maxConcurrentFiles) || 2),
      maxConnectionsPerFile: Math.max(1, Number(updated.maxConnectionsPerFile) || 8)
    }
    const s = config.Settings.createFrom(safeUpdated)
    settingsStore.set(s)
    await SaveSettings(s)
  } catch (err) {
    console.error('Failed to save settings:', err)
    throw err
  }
}

export async function updateSpeedLimit(limitMBps: number) {
  const numericLimit = Math.max(0, Number(limitMBps) || 0)
  let current: Settings | null = null
  settingsStore.update(s => {
    if (s) {
      current = config.Settings.createFrom({ ...s, speedLimitMBps: numericLimit })
      return current
    }
    return s
  })
  if (!current) {
    const fetched = await fetchSettings()
    if (fetched) {
      current = config.Settings.createFrom({ ...fetched, speedLimitMBps: numericLimit })
      settingsStore.set(current)
    }
  }
  if (current) {
    try {
      await SaveSettings(current)
    } catch (err) {
      console.error('Failed to update speed limit:', err)
    }
  }
}

export async function reorderRecentPaths(paths: string[]) {
  settingsStore.update(s => {
    if (!s) return s
    return config.Settings.createFrom({ ...s, recentPaths: paths })
  })
  try {
    await UpdateRecentPaths(paths)
  } catch (err) {
    console.error('Failed to persist reordered recent paths:', err)
  }
}

export async function removeRecentPath(pathToRemove: string) {
  let updatedPaths: string[] = []
  settingsStore.update(s => {
    if (!s) return s
    updatedPaths = (s.recentPaths || []).filter(p => p !== pathToRemove)
    return config.Settings.createFrom({ ...s, recentPaths: updatedPaths })
  })
  try {
    await UpdateRecentPaths(updatedPaths)
  } catch (err) {
    console.error('Failed to remove recent path:', err)
  }
}


export async function openTokenPage() {
  try {
    await OpenHFTokenPage()
  } catch (err) {
    console.error('Failed to open HF token page:', err)
  }
}
