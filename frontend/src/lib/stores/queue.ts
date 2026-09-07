import { writable } from 'svelte/store'
import type { DownloadItem } from '../types'
import {
  GetQueueItems,
  QueueItems,
  StartItem,
  PauseItem,
  ResumeItem,
  CancelItem,
  RemoveItem,
  UpdateItemDestination,
  VerifyLocalFile,
  OpenFolder
} from '../../../wailsjs/go/main/App'

export const queueStore = writable<DownloadItem[]>([])

export async function fetchQueue() {
  try {
    const items = await GetQueueItems()
    queueStore.set(items || [])
  } catch (err) {
    console.error('Failed to fetch queue items:', err)
  }
}

export function handleProgressUpdate(updated: DownloadItem) {
  queueStore.update(items => {
    const idx = items.findIndex(i => i.id === updated.id)
    if (idx !== -1) {
      const copy = [...items]
      copy[idx] = updated
      return copy
    }
    return [updated, ...items]
  })
}

export async function addQueueItems(items: DownloadItem[], autoStart: boolean) {
  try {
    const added = await QueueItems(items, autoStart)
    queueStore.update(current => {
      const ids = new Set(added.map(a => a.id))
      const filtered = current.filter(c => !ids.has(c.id))
      return [...added, ...filtered]
    })
    return added
  } catch (err) {
    console.error('Failed to queue items:', err)
    throw err
  }
}

export async function startTask(id: string) {
  try {
    await StartItem(id)
  } catch (err) {
    console.error('Failed to start task:', err)
  }
}

export async function pauseTask(id: string) {
  try {
    await PauseItem(id)
  } catch (err) {
    console.error('Failed to pause task:', err)
  }
}

export async function resumeTask(id: string) {
  try {
    await ResumeItem(id)
  } catch (err) {
    console.error('Failed to resume task:', err)
  }
}

export async function cancelTask(id: string) {
  try {
    await CancelItem(id)
  } catch (err) {
    console.error('Failed to cancel task:', err)
  }
}

export async function removeTask(id: string) {
  try {
    await RemoveItem(id)
    queueStore.update(items => items.filter(i => i.id !== id))
  } catch (err) {
    console.error('Failed to remove task:', err)
  }
}

export async function updateTaskDestination(id: string, newPath: string) {
  try {
    await UpdateItemDestination(id, newPath)
    queueStore.update(items =>
      items.map(i => (i.id === id ? { ...i, destinationDir: newPath } : i))
    )
  } catch (err) {
    console.error('Failed to update destination:', err)
  }
}

export async function verifyTaskFile(item: DownloadItem) {
  try {
    return await VerifyLocalFile(item)
  } catch (err) {
    console.error('Failed to verify local file:', err)
    throw err
  }
}

export async function openTaskFolder(folderPath: string) {
  try {
    await OpenFolder(folderPath)
  } catch (err) {
    console.error('Failed to open folder:', err)
  }
}
