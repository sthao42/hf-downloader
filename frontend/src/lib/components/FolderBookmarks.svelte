<script lang="ts">
  import { bookmarksStore, createBookmark, removeBookmark, browseDirectory, reorderBookmarks } from '../stores/bookmarks'
  import { settingsStore, reorderRecentPaths, removeRecentPath } from '../stores/settings'
  import { openTaskFolder } from '../stores/queue'
  import { Bookmark, Plus, Trash2, FolderOpen, ExternalLink, HardDrive, GripVertical, X } from 'lucide-svelte'
  import { flip } from 'svelte/animate'
  import { cubicOut } from 'svelte/easing'
  import type { FolderBookmark } from '../types'

  let newLabel: string = ''
  let newPath: string = ''
  let showAddForm: boolean = false
  let adding: boolean = false

  // Local synced array for bookmarks to enable live interactive displacement
  let localBookmarks: FolderBookmark[] = []
  let draggedBMId: string | null = null
  let lastSwapBMId: string | null = null

  // Keep localBookmarks in sync with store when not actively dragging
  $: if ($bookmarksStore) {
    if (draggedBMId === null) {
      localBookmarks = [...$bookmarksStore]
    }
  }

  // Local synced array for recent paths / quick picks
  let localRecentPaths: string[] = []
  let draggedRecentPath: string | null = null
  let lastSwapRecentPath: string | null = null

  $: if ($settingsStore?.recentPaths) {
    if (draggedRecentPath === null) {
      localRecentPaths = [...$settingsStore.recentPaths]
    }
  }

  // Drag handlers for saved bookmarks with live fluid push animation
  function handleDragStartBM(e: DragEvent, id: string) {
    draggedBMId = id
    lastSwapBMId = null
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/plain', id)
    }
  }

  function handleDragOverBM(e: DragEvent, targetId: string) {
    e.preventDefault()
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move'
    }
    if (!draggedBMId || draggedBMId === targetId || lastSwapBMId === targetId) return

    const fromIndex = localBookmarks.findIndex(b => b.id === draggedBMId)
    const toIndex = localBookmarks.findIndex(b => b.id === targetId)

    if (fromIndex !== -1 && toIndex !== -1 && fromIndex !== toIndex) {
      lastSwapBMId = targetId
      const updated = [...localBookmarks]
      const [moved] = updated.splice(fromIndex, 1)
      updated.splice(toIndex, 0, moved)
      localBookmarks = updated
    }
  }

  async function handleDropBM(e: DragEvent) {
    e.preventDefault()
    if (draggedBMId !== null) {
      draggedBMId = null
      lastSwapBMId = null
      bookmarksStore.set(localBookmarks)
      await reorderBookmarks(localBookmarks)
    }
  }

  async function handleDragEndBM() {
    if (draggedBMId !== null) {
      draggedBMId = null
      lastSwapBMId = null
      bookmarksStore.set(localBookmarks)
      await reorderBookmarks(localBookmarks)
    }
  }

  // Drag handlers for quick picks / recent paths with live fluid push animation
  function handleDragStartRecent(e: DragEvent, path: string) {
    draggedRecentPath = path
    lastSwapRecentPath = null
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/plain', path)
    }
  }

  function handleDragOverRecent(e: DragEvent, targetPath: string) {
    e.preventDefault()
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move'
    }
    if (!draggedRecentPath || draggedRecentPath === targetPath || lastSwapRecentPath === targetPath) return

    const fromIndex = localRecentPaths.indexOf(draggedRecentPath)
    const toIndex = localRecentPaths.indexOf(targetPath)

    if (fromIndex !== -1 && toIndex !== -1 && fromIndex !== toIndex) {
      lastSwapRecentPath = targetPath
      const updated = [...localRecentPaths]
      const [moved] = updated.splice(fromIndex, 1)
      updated.splice(toIndex, 0, moved)
      localRecentPaths = updated
    }
  }

  async function handleDropRecent(e: DragEvent) {
    e.preventDefault()
    if (draggedRecentPath !== null) {
      draggedRecentPath = null
      lastSwapRecentPath = null
      await reorderRecentPaths(localRecentPaths)
    }
  }

  async function handleDragEndRecent() {
    if (draggedRecentPath !== null) {
      draggedRecentPath = null
      lastSwapRecentPath = null
      await reorderRecentPaths(localRecentPaths)
    }
  }

  async function handleRemoveRecent(e: MouseEvent, pathToRemove: string) {
    e.preventDefault()
    e.stopPropagation()
    localRecentPaths = localRecentPaths.filter(p => p !== pathToRemove)
    await removeRecentPath(pathToRemove)
  }

  async function handleBrowse() {
    const chosen = await browseDirectory(newPath || '')
    if (chosen) {
      newPath = chosen
      if (!newLabel) {
        const parts = chosen.split(/[\\/]/)
        newLabel = parts[parts.length - 1] || 'My Models'
      }
    }
  }

  async function handleAddBookmark() {
    if (!newLabel.trim() || !newPath.trim()) return
    adding = true
    try {
      await createBookmark(newLabel.trim(), newPath.trim())
      newLabel = ''
      newPath = ''
      showAddForm = false
    } catch (e) {
      console.error(e)
    } finally {
      adding = false
    }
  }
</script>

<div class="bg-dark-850 border border-dark-700/60 rounded-2xl p-5 shadow-xl flex flex-col gap-4">
  <div class="flex items-center justify-between flex-wrap gap-2">
    <div class="flex items-center gap-2">
      <Bookmark class="w-4 h-4 text-accent-cyan" />
      <h3 class="text-sm font-bold text-white tracking-tight">Folder Bookmarks & Quick Picks</h3>
      <span class="text-[10px] text-slate-500 font-mono hidden sm:inline-block ml-1">
        • Drag cards to reorder live
      </span>
    </div>
    <button
      type="button"
      on:click={() => (showAddForm = !showAddForm)}
      class="px-3 py-1.5 rounded-lg bg-dark-800 hover:bg-dark-700 text-slate-200 border border-dark-700 text-xs font-medium flex items-center gap-1.5 transition-colors"
    >
      <Plus class="w-3.5 h-3.5 text-accent-indigo" />
      <span>{showAddForm ? 'Cancel' : 'Add Bookmark'}</span>
    </button>
  </div>

  <!-- Add Bookmark Form -->
  {#if showAddForm}
    <form on:submit|preventDefault={handleAddBookmark} class="p-4 bg-dark-950/70 border border-dark-700 rounded-xl flex flex-col gap-3 animate-in fade-in duration-150">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div>
          <label for="bm-label" class="text-[11px] text-slate-400 font-medium block mb-1">Bookmark Label</label>
          <input
            id="bm-label"
            type="text"
            bind:value={newLabel}
            placeholder="e.g. ComfyUI Checkpoints"
            class="w-full px-3 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-accent-indigo"
          />
        </div>
        <div>
          <label for="bm-path" class="text-[11px] text-slate-400 font-medium block mb-1">Directory Path</label>
          <div class="flex items-center gap-2">
            <input
              id="bm-path"
              type="text"
              bind:value={newPath}
              placeholder="e.g. D:\ComfyUI\models\checkpoints"
              class="w-full px-3 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-xs text-slate-200 font-mono focus:outline-none focus:border-accent-indigo"
            />
            <button
              type="button"
              on:click={handleBrowse}
              class="px-2.5 py-1.5 bg-dark-800 hover:bg-dark-700 text-slate-300 rounded-lg border border-dark-700 text-xs flex-shrink-0"
            >
              Browse
            </button>
          </div>
        </div>
      </div>
      <div class="flex justify-end pt-1">
        <button
          type="submit"
          disabled={adding || !newLabel.trim() || !newPath.trim()}
          class="px-4 py-1.5 bg-accent-indigo hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold disabled:opacity-50 transition-colors"
        >
          Save Bookmark
        </button>
      </div>
    </form>
  {/if}

  <!-- Bookmarks Cards Grid with Live Push & Slide Animation -->
  {#if localBookmarks.length === 0}
    <div class="py-8 text-center text-slate-500 text-xs border border-dashed border-dark-700/60 rounded-xl">
      No folder bookmarks saved yet. Click "Add Bookmark" above to save and organize your destination folders.
    </div>
  {:else}
    <div
      role="list"
      class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3"
      on:dragover|preventDefault
      on:drop={handleDropBM}
    >
      {#each localBookmarks as bm (bm.id)}
        {@const isDragging = draggedBMId === bm.id}
        <div
          animate:flip={{ duration: 250, easing: cubicOut }}
          role="listitem"
          draggable="true"
          on:dragstart={(e) => handleDragStartBM(e, bm.id)}
          on:dragover={(e) => handleDragOverBM(e, bm.id)}
          on:drop={handleDropBM}
          on:dragend={handleDragEndBM}
          class="p-3 rounded-xl flex flex-col justify-between gap-2 group cursor-grab active:cursor-grabbing transition-all duration-150 {isDragging ? 'opacity-30 border-2 border-dashed border-accent-indigo bg-accent-indigo/10 scale-95 shadow-inner' : 'bg-dark-900/90 border border-dark-700/70 hover:border-slate-500 shadow-md hover:shadow-lg'}"
        >
          <div class="flex items-start justify-between gap-2">
            <!-- Drag Handle & Label -->
            <div class="flex items-start gap-1.5 min-w-0 flex-1">
              <div
                class="p-0.5 -ml-1 text-slate-500 group-hover:text-accent-indigo transition-colors flex-shrink-0 mt-0.5"
                title="Click and hold to drag"
              >
                <GripVertical class="w-4 h-4" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="text-xs font-bold text-slate-200 truncate flex items-center gap-1.5">
                  <HardDrive class="w-3.5 h-3.5 text-accent-indigo flex-shrink-0" />
                  <span class="truncate">{bm.label}</span>
                </div>
                <div class="text-[11px] text-slate-500 font-mono truncate mt-0.5" title={bm.path}>
                  {bm.path}
                </div>
              </div>
            </div>

            <button
              type="button"
              draggable="false"
              on:click={() => removeBookmark(bm.id)}
              class="opacity-0 group-hover:opacity-100 p-1 hover:text-rose-400 text-slate-500 transition-opacity flex-shrink-0"
              title="Delete bookmark"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>

          <button
            type="button"
            draggable="false"
            on:click={() => openTaskFolder(bm.path)}
            class="w-full py-1 text-[11px] bg-dark-800/70 hover:bg-dark-700 text-slate-400 hover:text-slate-200 rounded border border-dark-700/50 flex items-center justify-center gap-1 transition-colors"
          >
            <ExternalLink class="w-3 h-3" />
            <span>Open in Explorer</span>
          </button>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Recent Destination Paths / Quick Picks with Live Push & Slide Animation -->
  {#if localRecentPaths && localRecentPaths.length > 0}
    <div class="pt-3 border-t border-dark-700/50">
      <div class="flex items-center justify-between mb-2">
        <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block">
          Quick Picks & Recent Paths:
        </span>
        <span class="text-[10px] text-slate-500 font-mono">
          Drag pills to reorganize live
        </span>
      </div>
      <div
        role="list"
        class="flex items-center gap-2 flex-wrap"
        on:dragover|preventDefault
        on:drop={handleDropRecent}
      >
        {#each localRecentPaths as recent (recent)}
          {@const isRecentDragging = draggedRecentPath === recent}
          <div
            animate:flip={{ duration: 200, easing: cubicOut }}
            role="listitem"
            draggable="true"
            on:dragstart={(e) => handleDragStartRecent(e, recent)}
            on:dragover={(e) => handleDragOverRecent(e, recent)}
            on:drop={handleDropRecent}
            on:dragend={handleDragEndRecent}
            class="group inline-flex items-center rounded-lg border cursor-grab active:cursor-grabbing transition-all duration-150 {isRecentDragging ? 'opacity-30 border-2 border-dashed border-accent-indigo bg-accent-indigo/10 scale-95 shadow-inner' : 'bg-dark-900 hover:bg-dark-800 border-dark-700 hover:border-slate-500'}"
          >
            <div
              class="pl-2 pr-0.5 py-1 text-slate-500 hover:text-accent-indigo transition-colors"
              title="Drag to reorganize"
            >
              <GripVertical class="w-3 h-3" />
            </div>
            <button
              type="button"
              draggable="false"
              on:click={() => openTaskFolder(recent)}
              class="py-1 pl-1 pr-1.5 text-xs font-mono text-slate-400 hover:text-slate-200 flex items-center gap-1.5 transition-colors min-w-0"
              title="Open {recent}"
            >
              <FolderOpen class="w-3 h-3 text-slate-500" />
              <span class="max-w-[200px] truncate">{recent}</span>
            </button>
            <button
              type="button"
              draggable="false"
              on:click={(e) => handleRemoveRecent(e, recent)}
              class="opacity-0 group-hover:opacity-100 p-1 mr-1 text-rose-500/80 hover:text-rose-400 hover:bg-rose-500/20 active:bg-rose-500/30 rounded transition-all duration-150 flex items-center justify-center flex-shrink-0"
              title="Remove {recent}"
              aria-label="Remove {recent}"
            >
              <X class="w-3 h-3" />
            </button>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

