import type { config, downloader, hfapi, main } from '../../wailsjs/go/models'

export type DownloadItem = downloader.DownloadItem
export type VerificationResult = downloader.VerificationResult
export type FolderBookmark = config.FolderBookmark
export type RoutingRule = config.RoutingRule
export type Settings = config.Settings
export type FileNode = hfapi.FileNode
export type ParsedTarget = hfapi.ParsedTarget
export type InspectResponse = main.InspectResponse

export interface StagedSelection {
  node: FileNode
  destinationDir: string
  selected: boolean
}
