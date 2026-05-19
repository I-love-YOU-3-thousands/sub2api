/**
 * User-facing asynchronous image generation endpoints.
 */

import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'

export type ImageStudioMode = 'text_to_image' | 'image_to_image'
export type ImageStudioStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'canceled'

export interface ImageStudioKey {
  id: number
  name: string
  group_id?: number
  group_name?: string
  platform: string
  status: string
}

export interface ImageStudioAsset {
  id: number
  seq: number
  kind: 'output'
  mime_type: string
  width: number
  height: number
  size_bytes: number
  revised_prompt?: string
  url: string
  created_at: string
}

export interface ImageStudioTask {
  id: number
  task_id: string
  api_key_id: number
  mode: ImageStudioMode
  model: string
  prompt: string
  ratio: string
  resolution: string
  size: string
  quality: string
  count: number
  status: ImageStudioStatus
  progress: number
  error?: string
  started_at?: string
  finished_at?: string
  created_at: string
  updated_at: string
  assets: ImageStudioAsset[]
}

export interface ImageStudioPromptOptimizationRequest {
  api_key_id: number
  prompt: string
  ratio: string
  resolution: string
  quality: string
  previous_prompt?: string
  variant?: number
}

export interface ImageStudioPromptOptimization {
  prompt: string
  source_prompt: string
  model: string
}

export async function listKeys(): Promise<ImageStudioKey[]> {
  const { data } = await apiClient.get<ImageStudioKey[]>('/image-studio/keys')
  return data
}

export async function createTask(formData: FormData): Promise<ImageStudioTask> {
  const { data } = await apiClient.post<ImageStudioTask>('/image-studio/tasks', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000
  })
  return data
}

export async function getTask(taskID: string): Promise<ImageStudioTask> {
  const { data } = await apiClient.get<ImageStudioTask>(`/image-studio/tasks/${taskID}`)
  return data
}

export async function listTasks(page = 1, pageSize = 20): Promise<PaginatedResponse<ImageStudioTask>> {
  const { data } = await apiClient.get<PaginatedResponse<ImageStudioTask>>('/image-studio/tasks', {
    params: { page, page_size: pageSize }
  })
  return data
}

export async function deleteTask(taskID: string): Promise<{ deleted: boolean }> {
  const { data } = await apiClient.delete<{ deleted: boolean }>(`/image-studio/tasks/${taskID}`)
  return data
}

export async function fetchAssetBlob(assetID: number): Promise<Blob> {
  const { data } = await apiClient.get<Blob>(`/image-studio/assets/${assetID}/content`, {
    responseType: 'blob'
  })
  return data
}

export async function optimizePrompt(
  payload: ImageStudioPromptOptimizationRequest
): Promise<ImageStudioPromptOptimization> {
  const { data } = await apiClient.post<ImageStudioPromptOptimization>('/image-studio/prompt/optimize', payload, {
    timeout: 70000
  })
  return data
}

const imageStudioAPI = {
  listKeys,
  createTask,
  getTask,
  listTasks,
  deleteTask,
  fetchAssetBlob,
  optimizePrompt
}

export default imageStudioAPI
