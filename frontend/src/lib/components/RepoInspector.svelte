<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import type { FileNode, ParsedTarget } from '../types'
  import { formatBytes, detectQuantBadge } from '../utils'
  import { fetchDiskSpace } from '../stores/disk'
  import { settingsStore } from '../stores/settings'
  import type { platform } from '../../../wailsjs/go/models'
  import {
    FileText,
    Download,
    FolderSync,
    SlidersHorizontal,
    Search,
    HardDrive,
    ChevronRight,
    ArrowUp,
    ArrowDown,
    ArrowUpDown
  } from 'lucide-svelte'

  export let target: ParsedTarget | null = null
  export let files: FileNode[] = []
  export let defaultDestinations: Record<string, string> = {}

  const dispatch = createEventDispatcher<{
    stage: { selectedFiles: { file: FileNode; dest: string }[]; autoStart: boolean }
    openRouter: { selectedFiles: { file: FileNode; dest: string }[] }
    changeDest: { filePath: string }
  }>()

  let selectedMap: Record<string, boolean> = {}
  let searchFilter: string = ''
  let activeCategory: 'all' | 'safetensors' | 'gguf' | 'text_encoders' | 'vae' = 'all'

  let diskInfo: platform.DiskSpaceInfo | null = null
  let lastCheckedDir: string = ''

  $: primaryDestDir = (selectedFiles[0] && selectedFiles[0].dest) || (files[0] && defaultDestinations[files[0].path]) || $settingsStore?.defaultDownloadDir || ''

  $: if (primaryDestDir && primaryDestDir !== lastCheckedDir) {
    lastCheckedDir = primaryDestDir
    fetchDiskSpace(primaryDestDir).then(info => {
      diskInfo = info
    })
  }

  // Initialize all files as selected on load
  $: if (files) {
    const next: Record<string, boolean> = {}
    for (const f of files) {
      next[f.path] = selectedMap[f.path] !== undefined ? selectedMap[f.path] : true
    }
    selectedMap = next
  }

  $: filteredFiles = files.filter(f => {
    const lower = f.path.toLowerCase()
    if (searchFilter && !lower.includes(searchFilter.toLowerCase())) {
      return false
    }
    if (activeCategory === 'safetensors') return lower.endsWith('.safetensors')
    if (activeCategory === 'gguf') return lower.endsWith('.gguf')
    if (activeCategory === 'text_encoders') return lower.includes('clip') || lower.includes('t5') || lower.includes('text_encoder')
    if (activeCategory === 'vae') return lower.includes('vae')
    return true
  })

  let sortField: 'name' | 'size' | null = null
  let sortDirection: 'asc' | 'desc' = 'asc'

  function handleSort(field: 'name' | 'size') {
    if (sortField === field) {
      sortDirection = sortDirection === 'asc' ? 'desc' : 'asc'
    } else {
      sortField = field
      sortDirection = field === 'size' ? 'desc' : 'asc'
    }
  }

  $: sortedFiles = [...filteredFiles].sort((a, b) => {
    if (sortField === 'name') {
      const nameA = a.path.toLowerCase()
      const nameB = b.path.toLowerCase()
      const cmp = nameA.localeCompare(nameB)
      return sortDirection === 'asc' ? cmp : -cmp
    }
    if (sortField === 'size') {
      const sizeA = a.size || 0
      const sizeB = b.size || 0
      const cmp = sizeA - sizeB
      return sortDirection === 'asc' ? cmp : -cmp
    }
    return 0
  })

  $: selectedFiles = files
    .filter(f => selectedMap[f.path])
    .map(f => ({ file: f, dest: defaultDestinations[f.path] || '' }))

  $: totalSelectedBytes = selectedFiles.reduce((acc, curr) => acc + (curr.file.size || 0), 0)

  $: isAllFilteredSelected = filteredFiles.length > 0 && filteredFiles.every(f => !!selectedMap[f.path])
  $: isSomeFilteredSelected = filteredFiles.some(f => !!selectedMap[f.path]) && !isAllFilteredSelected

  function toggleAll(selectAll: boolean) {
    const next: Record<string, boolean> = { ...selectedMap }
    for (const f of filteredFiles) {
      next[f.path] = selectAll
    }
    selectedMap = next
  }

  function handleMasterCheckboxToggle() {
    const targetState = !isAllFilteredSelected
    toggleAll(targetState)
  }

  function handleStage(autoStart: boolean) {
    if (selectedFiles.length === 0) return

    // Storage safety check: prompt if available disk space is insufficient
    if (diskInfo && totalSelectedBytes > 0) {
      const buffer = 100 * 1024 * 1024 // 100MB reserve buffer
      if (diskInfo.availableBytes < totalSelectedBytes + buffer) {
        const neededStr = formatBytes(totalSelectedBytes)
        const availStr = formatBytes(diskInfo.availableBytes)
        const proceed = confirm(
          `Storage Drive Warning:\n\n` +
          `Destination drive (${diskInfo.path}) only has ${availStr} available, but selected files require ${neededStr}.\n\n` +
          `Downloading may stall or fail if storage runs out.\n\n` +
          `Do you want to proceed anyway?`
        )
        if (!proceed) return
      }
    }

    dispatch('stage', { selectedFiles, autoStart })
  }

  function handleOpenRouter() {
    if (selectedFiles.length === 0) return
    dispatch('openRouter', { selectedFiles })
  }
</script>

{#if target}
  <div class="bg-dark-850 border border-dark-700/60 rounded-2xl p-5 shadow-xl flex flex-col gap-4">
    <!-- Header: Repo Metadata -->
    <div class="flex items-start justify-between flex-wrap gap-3 pb-4 border-b border-dark-700/50">
      <div>
        <div class="flex items-center gap-2 text-xs font-mono text-slate-400 mb-1">
          <span class="px-2 py-0.5 rounded bg-dark-800 text-accent-indigo border border-dark-700 uppercase font-semibold">
            {target.type}
          </span>
          <ChevronRight class="w-3.5 h-3.5 text-slate-600" />
          <span class="text-slate-300">{target.repoId}</span>
          <span class="text-slate-600">@</span>
          <span class="text-accent-cyan">{target.revision}</span>
          {#if target.subpath}
            <ChevronRight class="w-3.5 h-3.5 text-slate-600" />
            <span class="text-amber-400">{target.subpath}</span>
          {/if}
        </div>
        <h2 class="text-lg font-bold text-white tracking-tight flex items-center gap-2">
          <span>{target.repoId.split('/')[1] || target.repoId}</span>
        </h2>
      </div>

      <!-- Action Buttons & Storage Space Safety Indicator -->
      <div class="flex items-center gap-2 flex-wrap">
        {#if diskInfo}
          {@const isSpaceShort = diskInfo.availableBytes < totalSelectedBytes + 100 * 1024 * 1024}
          <div
            class="px-3 py-2 rounded-xl text-xs font-medium flex items-center gap-1.5 transition-colors border {isSpaceShort ? 'bg-rose-500/15 border-rose-500/40 text-rose-300' : 'bg-dark-800/80 border-dark-700 text-slate-300'}"
            title="Available storage on {diskInfo.path}"
          >
            <HardDrive class="w-3.5 h-3.5 {isSpaceShort ? 'text-rose-400 animate-pulse' : 'text-accent-cyan'}" />
            <span>Drive:</span>
            <span class="font-bold {isSpaceShort ? 'text-rose-400' : 'text-white'}">
              {formatBytes(diskInfo.availableBytes)} free
            </span>
            {#if isSpaceShort && totalSelectedBytes > 0}
              <span class="px-1.5 py-0.5 rounded bg-rose-500/30 text-rose-200 text-[10px] uppercase font-bold tracking-wider">
                Low
              </span>
            {/if}
          </div>
        {/if}

        <button
          type="button"
          on:click={handleOpenRouter}
          disabled={selectedFiles.length === 0}
          class="px-3 py-2 rounded-xl bg-dark-800 hover:bg-dark-700 text-slate-300 border border-dark-700 hover:border-slate-600 text-xs font-medium flex items-center gap-1.5 transition-colors disabled:opacity-40"
        >
          <SlidersHorizontal class="w-3.5 h-3.5 text-accent-cyan" />
          <span>Route Destinations ({selectedFiles.length})</span>
        </button>

        <button
          type="button"
          on:click={() => handleStage(false)}
          disabled={selectedFiles.length === 0}
          class="px-3.5 py-2 rounded-xl bg-dark-800 hover:bg-dark-700 text-slate-200 border border-dark-700 hover:border-slate-500 text-xs font-medium flex items-center gap-1.5 transition-colors disabled:opacity-40"
        >
          <FolderSync class="w-3.5 h-3.5 text-accent-amber" />
          <span>Queue Selected</span>
        </button>

        <button
          type="button"
          on:click={() => handleStage(true)}
          disabled={selectedFiles.length === 0}
          class="px-4 py-2 rounded-xl bg-gradient-to-r from-accent-indigo to-accent-purple hover:from-indigo-500 hover:to-purple-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 flex items-center gap-1.5 transition-all disabled:opacity-40"
        >
          <Download class="w-3.5 h-3.5" />
          <span>Download Now ({formatBytes(totalSelectedBytes)})</span>
        </button>
      </div>
    </div>

    <!-- Filters & Search Toolbar -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <!-- Category Pills -->
      <div class="flex items-center gap-1.5 flex-wrap">
        <button
          type="button"
          on:click={() => (activeCategory = 'all')}
          class="px-2.5 py-1 rounded-lg text-xs font-medium transition-colors {activeCategory === 'all' ? 'bg-accent-indigo text-white' : 'bg-dark-800 text-slate-400 hover:text-slate-200'}"
        >
          All ({files.length})
        </button>
        <button
          type="button"
          on:click={() => (activeCategory = 'safetensors')}
          class="px-2.5 py-1 rounded-lg text-xs font-medium transition-colors {activeCategory === 'safetensors' ? 'bg-accent-indigo text-white' : 'bg-dark-800 text-slate-400 hover:text-slate-200'}"
        >
          Safetensors
        </button>
        <button
          type="button"
          on:click={() => (activeCategory = 'gguf')}
          class="px-2.5 py-1 rounded-lg text-xs font-medium transition-colors {activeCategory === 'gguf' ? 'bg-accent-indigo text-white' : 'bg-dark-800 text-slate-400 hover:text-slate-200'}"
        >
          GGUF Quants
        </button>
        <button
          type="button"
          on:click={() => (activeCategory = 'text_encoders')}
          class="px-2.5 py-1 rounded-lg text-xs font-medium transition-colors {activeCategory === 'text_encoders' ? 'bg-accent-indigo text-white' : 'bg-dark-800 text-slate-400 hover:text-slate-200'}"
        >
          Text Encoders
        </button>
        <button
          type="button"
          on:click={() => (activeCategory = 'vae')}
          class="px-2.5 py-1 rounded-lg text-xs font-medium transition-colors {activeCategory === 'vae' ? 'bg-accent-indigo text-white' : 'bg-dark-800 text-slate-400 hover:text-slate-200'}"
        >
          VAEs
        </button>
      </div>

      <!-- Search & Select All controls -->
      <div class="flex items-center gap-3">
        <div class="relative w-48">
          <Search class="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-slate-500" />
          <input
            type="text"
            bind:value={searchFilter}
            placeholder="Filter files..."
            class="w-full pl-8 pr-2.5 py-1.5 bg-dark-950/80 border border-dark-700 rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-accent-indigo"
          />
        </div>

        <label
          class="flex items-center gap-1.5 text-xs cursor-pointer text-slate-300 hover:text-white select-none px-2.5 py-1.5 rounded-lg bg-dark-900/80 border border-dark-700 hover:border-slate-500 transition-colors flex-shrink-0"
          title="Click to select or deselect all filtered files"
        >
          <input
            type="checkbox"
            checked={isAllFilteredSelected}
            indeterminate={isSomeFilteredSelected}
            on:change={handleMasterCheckboxToggle}
            class="rounded border-dark-700 text-accent-indigo focus:ring-0 bg-dark-950 w-4 h-4 cursor-pointer"
          />
          <span class="font-medium text-xs">Select All</span>
        </label>
      </div>
    </div>

    <!-- File Tree List Container with Top Header Row -->
    <div class="border border-dark-700/60 rounded-xl bg-dark-950/40 overflow-hidden flex flex-col">
      <!-- Top Row: Column Sort Headers (Name & Size) -->
      <div class="flex items-center justify-between px-3 py-2.5 bg-dark-900/90 border-b border-dark-700/60 text-xs font-semibold select-none">
        <!-- Left: Checkbox + Name Sort Button -->
        <div class="flex items-center gap-3 min-w-0 flex-1">
          <input
            type="checkbox"
            checked={isAllFilteredSelected}
            indeterminate={isSomeFilteredSelected}
            on:change={handleMasterCheckboxToggle}
            class="rounded border-dark-700 text-accent-indigo focus:ring-0 bg-dark-950 w-4 h-4 cursor-pointer"
            title="Select or deselect all filtered files"
          />
          <button
            type="button"
            on:click={() => handleSort('name')}
            class="inline-flex items-center gap-1.5 transition-colors group focus:outline-none {sortField === 'name' ? 'text-accent-cyan font-bold' : 'text-slate-300 hover:text-white'}"
            title="Sort by file name ({sortField === 'name' ? (sortDirection === 'asc' ? 'A to Z (click for Z to A)' : 'Z to A (click for A to Z)') : 'Click to sort A to Z'})"
          >
            <span>Name</span>
            {#if sortField === 'name'}
              {#if sortDirection === 'asc'}
                <ArrowUp class="w-3.5 h-3.5 text-accent-cyan" />
              {:else}
                <ArrowDown class="w-3.5 h-3.5 text-accent-cyan" />
              {/if}
            {:else}
              <ArrowUpDown class="w-3.5 h-3.5 text-slate-500 opacity-50 group-hover:opacity-100 transition-opacity" />
            {/if}
          </button>
        </div>

        <!-- Right: Size Sort Button + Destination Spacer -->
        <div class="flex items-center gap-3 flex-shrink-0">
          <button
            type="button"
            on:click={() => handleSort('size')}
            class="inline-flex items-center gap-1.5 transition-colors group focus:outline-none {sortField === 'size' ? 'text-accent-cyan font-bold' : 'text-slate-300 hover:text-white'}"
            title="Sort by file size ({sortField === 'size' ? (sortDirection === 'desc' ? 'Largest first (click for Smallest first)' : 'Smallest first (click for Largest first)') : 'Click to sort Largest first'})"
          >
            <span>Size</span>
            {#if sortField === 'size'}
              {#if sortDirection === 'asc'}
                <ArrowUp class="w-3.5 h-3.5 text-accent-cyan" />
              {:else}
                <ArrowDown class="w-3.5 h-3.5 text-accent-cyan" />
              {/if}
            {:else}
              <ArrowUpDown class="w-3.5 h-3.5 text-slate-500 opacity-50 group-hover:opacity-100 transition-opacity" />
            {/if}
          </button>
          <span class="w-[74px] text-right text-[11px] text-slate-500 font-normal">
            Destination
          </span>
        </div>
      </div>

      <!-- File Tree List (Scrollable) -->
      <div class="max-h-[340px] overflow-y-auto divide-y divide-dark-800/80">
        {#if sortedFiles.length === 0}
          <div class="p-8 text-center text-slate-500 text-sm">
            No files match the current filter.
          </div>
        {:else}
          {#each sortedFiles as file (file.path)}
            {@const quant = detectQuantBadge(file.path)}
            {@const isSelected = !!selectedMap[file.path]}
            <div
              class="flex items-center justify-between p-3 hover:bg-dark-800/40 transition-colors gap-3 {isSelected ? 'bg-indigo-950/10' : 'opacity-70'}"
            >
              <label class="flex items-center gap-3 min-w-0 flex-1 cursor-pointer">
                <input
                  type="checkbox"
                  bind:checked={selectedMap[file.path]}
                  class="rounded border-dark-700 text-accent-indigo focus:ring-0 bg-dark-900 w-4 h-4 cursor-pointer"
                />
                <FileText class="w-4 h-4 text-slate-500 flex-shrink-0" />
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="text-xs font-mono font-medium text-slate-200 truncate">{file.path}</span>
                    {#if quant}
                      <span class="px-2 py-0.5 rounded text-[10px] font-semibold border {quant.color}">
                        {quant.label}
                      </span>
                    {/if}
                  </div>
                  <!-- Pre-routed destination indicator -->
                  <div class="text-[11px] text-slate-500 flex items-center gap-1 mt-0.5 truncate">
                    <HardDrive class="w-3 h-3 text-slate-600 flex-shrink-0" />
                    <span class="truncate">{defaultDestinations[file.path] || $settingsStore?.defaultDownloadDir || 'Default Folder'}</span>
                  </div>
                </div>
              </label>

              <!-- File Size & Destination switch -->
              <div class="flex items-center gap-3 flex-shrink-0">
                <span class="text-xs font-mono font-semibold text-slate-300">
                  {formatBytes(file.size)}
                </span>
                <button
                  type="button"
                  on:click={() => dispatch('changeDest', { filePath: file.path })}
                  class="w-[74px] px-2 py-1 text-[11px] text-slate-400 hover:text-accent-cyan bg-dark-800 hover:bg-dark-700 rounded border border-dark-700/60 transition-colors text-center"
                >
                  Change Dir
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Summary Footer -->
    <div class="flex items-center justify-between text-xs text-slate-400 px-1 pt-1">
      <div>
        Showing <span class="text-slate-200 font-semibold">{filteredFiles.length}</span> of <span class="text-slate-200">{files.length}</span> files
      </div>
      <div>
        Selected: <span class="text-accent-cyan font-bold font-mono">{selectedFiles.length}</span> files (<span class="text-accent-indigo font-bold font-mono">{formatBytes(totalSelectedBytes)}</span>)
      </div>
    </div>
  </div>
{/if}
