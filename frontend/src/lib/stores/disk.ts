import { CheckDiskSpace } from '../../../wailsjs/go/main/App'
import type { platform } from '../../../wailsjs/go/models'

export async function fetchDiskSpace(targetDir: string): Promise<platform.DiskSpaceInfo | null> {
  try {
    const res = await CheckDiskSpace(targetDir || '.')
    return res
  } catch (err) {
    console.warn('Failed to query disk space for', targetDir, err)
    return null
  }
}
