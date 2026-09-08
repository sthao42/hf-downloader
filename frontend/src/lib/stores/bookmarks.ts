import { writable } from 'svelte/store'
import type { FolderBookmark } from '../types'
import {
  GetBookmarks,
  AddBookmark,
  DeleteBookmark,
  UpdateBookmarks,
  SelectDirectoryDialog
} from '../../../wailsjs/go/main/App'
import { config } from '../../../wailsjs/go/models'

export const bookmarksStore = writable<FolderBookmark[]>([])

export async function fetchBookmarks() {
  try {
    const list = await GetBookmarks()
    bookmarksStore.set(list || [])
  } catch (err) {
    console.error('Failed to fetch bookmarks:', err)
  }
}

export async function createBookmark(label: string, path: string) {
  try {
    const bm = await AddBookmark(label, path)
    bookmarksStore.update(list => [...list, bm])
    return bm
  } catch (err) {
    console.error('Failed to add bookmark:', err)
    throw err
  }
}

export async function removeBookmark(id: string) {
  try {
    await DeleteBookmark(id)
    bookmarksStore.update(list => list.filter(b => b.id !== id))
  } catch (err) {
    console.error('Failed to delete bookmark:', err)
    throw err
  }
}

export async function browseDirectory(defaultPath: string = ''): Promise<string> {
  try {
    return await SelectDirectoryDialog(defaultPath)
  } catch (err) {
    console.error('Failed to open directory dialog:', err)
    return ''
  }
}

export async function reorderBookmarks(reordered: FolderBookmark[]) {
  bookmarksStore.set(reordered)
  try {
    const models = reordered.map(b => config.FolderBookmark.createFrom(b))
    await UpdateBookmarks(models)
  } catch (err) {
    console.error('Failed to persist reordered bookmarks:', err)
  }
}
