import type { ImageStudioAsset, ImageStudioTask } from '@/api/imageStudio'

const HISTORY_STORAGE_KEY_PREFIX = 'image_studio_local_history_v1'
const DB_NAME = 'sub2api-image-studio'
const DB_VERSION = 1
const ASSET_STORE = 'assets'
const MAX_HISTORY_ITEMS = 100

export type StoredImageStudioAsset = ImageStudioAsset & {
  local_blob_key?: string
}

export type LocalImageStudioTask = ImageStudioTask & {
  assets: StoredImageStudioAsset[]
}

interface AssetBlobRecord {
  key: string
  blob: Blob
  updated_at: string
}

let dbPromise: Promise<IDBDatabase> | null = null

export function loadLocalImageStudioHistory(ownerID?: string | number | null): LocalImageStudioTask[] {
  if (typeof window === 'undefined') return []
  try {
    const raw = window.localStorage.getItem(historyStorageKey(ownerID))
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter(isLocalImageStudioTask)
      .slice(0, MAX_HISTORY_ITEMS)
  } catch (error) {
    console.error('Failed to load local image studio history:', error)
    return []
  }
}

export function saveLocalImageStudioHistory(tasks: LocalImageStudioTask[], ownerID?: string | number | null) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(historyStorageKey(ownerID), JSON.stringify(tasks.slice(0, MAX_HISTORY_ITEMS)))
  } catch (error) {
    console.error('Failed to save local image studio history:', error)
  }
}

export function mergeLocalImageStudioTask(
  currentTask: LocalImageStudioTask | undefined,
  nextTask: ImageStudioTask
): LocalImageStudioTask {
  const existingBlobKeys = new Map<number, string>()
  for (const asset of currentTask?.assets || []) {
    if (asset.local_blob_key) {
      existingBlobKeys.set(asset.id, asset.local_blob_key)
    }
  }

  return {
    ...nextTask,
    assets: (nextTask.assets || []).map((asset) => ({
      ...asset,
      local_blob_key: existingBlobKeys.get(asset.id) || localAssetBlobKey(asset.id)
    }))
  }
}

export async function loadLocalImageStudioAssetBlob(asset: StoredImageStudioAsset): Promise<Blob | null> {
  if (!asset.local_blob_key) return null
  try {
    const db = await openImageStudioDB()
    const record = await requestToPromise<AssetBlobRecord | undefined>(
      db.transaction(ASSET_STORE, 'readonly').objectStore(ASSET_STORE).get(asset.local_blob_key)
    )
    return record?.blob instanceof Blob ? record.blob : null
  } catch (error) {
    console.error('Failed to load local image studio asset blob:', error)
    return null
  }
}

export async function saveLocalImageStudioAssetBlob(asset: StoredImageStudioAsset, blob: Blob) {
  if (!asset.local_blob_key) return
  try {
    const db = await openImageStudioDB()
    const record: AssetBlobRecord = {
      key: asset.local_blob_key,
      blob,
      updated_at: new Date().toISOString()
    }
    await requestToPromise(db.transaction(ASSET_STORE, 'readwrite').objectStore(ASSET_STORE).put(record))
  } catch (error) {
    console.error('Failed to save local image studio asset blob:', error)
  }
}

export async function deleteLocalImageStudioTaskAssets(task: LocalImageStudioTask) {
  const localAssets = task.assets as StoredImageStudioAsset[]
  const keys = localAssets
    .map((asset) => asset.local_blob_key)
    .filter((key): key is string => Boolean(key))
  if (keys.length === 0) return
  try {
    const db = await openImageStudioDB()
    const store = db.transaction(ASSET_STORE, 'readwrite').objectStore(ASSET_STORE)
    await Promise.all(keys.map((key) => requestToPromise(store.delete(key))))
  } catch (error) {
    console.error('Failed to delete local image studio asset blobs:', error)
  }
}

function openImageStudioDB(): Promise<IDBDatabase> {
  if (dbPromise) return dbPromise
  dbPromise = new Promise((resolve, reject) => {
    const request = window.indexedDB.open(DB_NAME, DB_VERSION)
    request.onerror = () => reject(request.error)
    request.onsuccess = () => resolve(request.result)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(ASSET_STORE)) {
        db.createObjectStore(ASSET_STORE, { keyPath: 'key' })
      }
    }
  })
  return dbPromise
}

function requestToPromise<T = unknown>(request: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onerror = () => reject(request.error)
    request.onsuccess = () => resolve(request.result)
  })
}

function localAssetBlobKey(assetID: number) {
  return `asset:${assetID}`
}

function historyStorageKey(ownerID?: string | number | null) {
  return ownerID ? `${HISTORY_STORAGE_KEY_PREFIX}:${ownerID}` : HISTORY_STORAGE_KEY_PREFIX
}

function isLocalImageStudioTask(value: unknown): value is LocalImageStudioTask {
  if (!value || typeof value !== 'object') return false
  const task = value as Partial<LocalImageStudioTask>
  return typeof task.task_id === 'string' && typeof task.prompt === 'string' && Array.isArray(task.assets)
}
