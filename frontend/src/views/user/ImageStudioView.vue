<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('imageStudio.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('imageStudio.description') }}</p>
        </div>
        <button class="btn btn-secondary" :disabled="loadingKeys || loadingTasks" @click="refreshAll">
          <Icon name="refresh" size="md" :class="loadingKeys || loadingTasks ? 'animate-spin' : ''" />
          <span>{{ t('imageStudio.refresh') }}</span>
        </button>
      </div>

      <div class="grid gap-6 xl:grid-cols-[420px_minmax(0,1fr)]">
        <section class="card p-5">
          <form class="space-y-5" @submit.prevent="handleSubmit">
            <div>
              <label class="input-label" for="image-studio-key">{{ t('imageStudio.key') }}</label>
              <select
                id="image-studio-key"
                v-model.number="selectedKeyID"
                class="input"
                :disabled="loadingKeys || keys.length === 0"
              >
                <option :value="0" disabled>{{ t('imageStudio.selectKey') }}</option>
                <option v-for="key in keys" :key="key.id" :value="key.id">
                  {{ keyOptionLabel(key) }}
                </option>
              </select>
              <div
                v-if="!loadingKeys && keys.length === 0"
                class="mt-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800/50 dark:bg-amber-900/20 dark:text-amber-200"
              >
                <p>{{ t('imageStudio.noKeys') }}</p>
                <RouterLink class="btn btn-secondary btn-sm mt-3" to="/keys">
                  <Icon name="key" size="sm" />
                  <span>{{ t('imageStudio.goToKeys') }}</span>
                </RouterLink>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-2 rounded-xl bg-gray-100 p-1 dark:bg-dark-900">
              <button
                v-for="option in modeOptions"
                :key="option.value"
                type="button"
                class="inline-flex items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition"
                :class="mode === option.value
                  ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-200'
                  : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
                @click="mode = option.value"
              >
                <Icon :name="option.icon" size="sm" />
                <span>{{ t(option.labelKey) }}</span>
              </button>
            </div>

            <div>
              <label class="input-label" for="image-studio-model">{{ t('imageStudio.model') }}</label>
              <input id="image-studio-model" v-model.trim="model" class="input" type="text" />
            </div>

            <div>
              <div class="flex items-center justify-between gap-3">
                <label class="input-label" for="image-studio-prompt">{{ t('imageStudio.prompt') }}</label>
                <button
                  type="button"
                  class="btn btn-secondary btn-sm shrink-0"
                  :disabled="optimizingPrompt || !prompt.trim() || !selectedKeyID"
                  @click="handleOptimizePrompt(false)"
                >
                  <Icon name="sparkles" size="sm" :class="optimizingPrompt ? 'animate-pulse' : ''" />
                  <span>{{ optimizingPrompt ? t('imageStudio.optimizingPrompt') : t('imageStudio.optimizePrompt') }}</span>
                </button>
              </div>
              <textarea
                id="image-studio-prompt"
                v-model="prompt"
                class="input min-h-32 resize-y"
                :placeholder="t('imageStudio.promptPlaceholder')"
              />
            </div>

            <div v-if="mode === 'image_to_image'" class="space-y-3">
              <label class="input-label">{{ t('imageStudio.referenceImage') }}</label>
              <label
                class="flex cursor-pointer flex-col items-center justify-center rounded-xl border border-dashed border-gray-300 bg-gray-50 px-4 py-5 text-center transition hover:border-primary-400 hover:bg-primary-50 dark:border-dark-600 dark:bg-dark-900/60 dark:hover:border-primary-500/70 dark:hover:bg-primary-900/20"
              >
                <input
                  ref="referenceInput"
                  class="sr-only"
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  @change="handleReferenceChange"
                />
                <Icon name="upload" size="lg" class="text-gray-400 dark:text-dark-400" />
                <span class="mt-2 text-sm font-medium text-gray-700 dark:text-dark-200">
                  {{ t('imageStudio.upload') }}
                </span>
                <span class="mt-1 text-xs text-gray-500 dark:text-dark-400">PNG, JPG, WEBP</span>
              </label>
              <div
                v-if="referenceFile"
                class="overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800"
              >
                <div class="flex items-center justify-between gap-3 px-3 py-2">
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ referenceFile.name }}</p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">{{ formatBytes(referenceFile.size) }}</p>
                  </div>
                  <button type="button" class="btn btn-ghost btn-icon" :title="t('imageStudio.removeImage')" @click="clearReference">
                    <Icon name="x" size="sm" />
                  </button>
                </div>
                <img
                  v-if="referencePreviewUrl"
                  :src="referencePreviewUrl"
                  :alt="referenceFile.name"
                  class="max-h-64 w-full border-t border-gray-100 object-contain dark:border-dark-700"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label" for="image-studio-ratio">{{ t('imageStudio.ratio') }}</label>
                <select id="image-studio-ratio" v-model="ratio" class="input">
                  <option v-for="item in ratios" :key="item" :value="item">{{ item }}</option>
                </select>
              </div>
              <div>
                <label class="input-label" for="image-studio-resolution">{{ t('imageStudio.resolution') }}</label>
                <select id="image-studio-resolution" v-model="resolution" class="input">
                  <option v-for="item in resolutions" :key="item" :value="item">{{ item }}</option>
                </select>
              </div>
              <div>
                <label class="input-label" for="image-studio-quality">{{ t('imageStudio.quality') }}</label>
                <select id="image-studio-quality" v-model="quality" class="input">
                  <option v-for="item in qualities" :key="item" :value="item">{{ qualityLabel(item) }}</option>
                </select>
              </div>
              <div>
                <label class="input-label" for="image-studio-count">{{ t('imageStudio.count') }}</label>
                <select id="image-studio-count" v-model.number="count" class="input">
                  <option v-for="item in counts" :key="item" :value="item">{{ item }}</option>
                </select>
              </div>
            </div>

            <p
              v-if="formError"
              class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-800/50 dark:bg-red-900/20 dark:text-red-200"
            >
              {{ formError }}
            </p>

            <button class="btn btn-primary w-full" type="submit" :disabled="submitting || optimizingPrompt || keys.length === 0">
              <Icon name="sparkles" size="md" :class="submitting ? 'animate-pulse' : ''" />
              <span>{{ submitting ? t('imageStudio.generating') : t('imageStudio.generate') }}</span>
            </button>
          </form>
        </section>

        <section class="card overflow-hidden">
          <div class="flex items-center justify-between gap-3 border-b border-gray-100 px-5 py-4 dark:border-dark-700">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('imageStudio.history') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('imageStudio.historySubtitle') }}</p>
            </div>
            <button class="btn btn-secondary btn-icon" :disabled="loadingTasks" :title="t('imageStudio.refresh')" @click="loadTasks(true)">
              <Icon name="refresh" size="md" :class="loadingTasks ? 'animate-spin' : ''" />
            </button>
          </div>

          <div v-if="loadingTasks && tasks.length === 0" class="flex min-h-80 items-center justify-center">
            <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
          </div>

          <div v-else-if="tasks.length === 0" class="flex min-h-80 items-center justify-center px-6 text-center">
            <div>
              <Icon name="sparkles" size="xl" class="mx-auto text-gray-300 dark:text-dark-500" />
              <p class="mt-3 text-sm text-gray-500 dark:text-dark-400">{{ t('imageStudio.empty') }}</p>
            </div>
          </div>

          <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
            <article v-for="task in tasks" :key="task.task_id" class="p-5">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-2">
                    <span :class="statusBadgeClass(task.status)">{{ statusLabel(task.status) }}</span>
                    <span class="rounded-full bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                      {{ modeLabel(task.mode) }}
                    </span>
                    <span class="rounded-full bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                      {{ task.size }}
                    </span>
                    <span class="rounded-full bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                      {{ task.quality || 'auto' }}
                    </span>
                    <span v-if="task.count > 1" class="rounded-full bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                      x{{ task.count }}
                    </span>
                  </div>
                  <p class="mt-2 max-h-12 overflow-hidden text-sm leading-6 text-gray-700 dark:text-dark-200">
                    {{ task.prompt }}
                  </p>
                  <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">{{ formatDate(task.created_at) }}</p>
                </div>
                <button
                  class="btn btn-ghost btn-icon text-gray-500 hover:text-red-600 dark:text-dark-300 dark:hover:text-red-300"
                  :disabled="deletingTaskIDs[task.task_id]"
                  :title="t('imageStudio.delete')"
                  @click="handleDelete(task)"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>

              <div v-if="isActiveStatus(task.status)" class="mt-4">
                <div class="mb-1 flex items-center justify-between text-xs text-gray-500 dark:text-dark-400">
                  <span>{{ t('imageStudio.progress') }}</span>
                  <span>{{ task.progress }}%</span>
                </div>
                <div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
                  <div class="h-full rounded-full bg-primary-500 transition-all" :style="{ width: `${task.progress}%` }"></div>
                </div>
              </div>

              <p
                v-if="task.status === 'failed' && task.error"
                class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-800/50 dark:bg-red-900/20 dark:text-red-200"
              >
                {{ task.error }}
              </p>

              <div v-if="outputAssets(task).length > 0" class="mt-4 grid gap-3 sm:grid-cols-2 2xl:grid-cols-3">
                <figure
                  v-for="asset in outputAssets(task)"
                  :key="asset.id"
                  class="overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800"
                >
                  <div class="flex aspect-square items-center justify-center bg-gray-100 dark:bg-dark-900">
                    <button
                      v-if="assetObjectUrls[asset.id]"
                      type="button"
                      class="flex h-full w-full cursor-zoom-in items-center justify-center"
                      :title="t('imageStudio.viewOriginal')"
                      @click="openImageViewer(task, asset)"
                    >
                      <img
                        :src="assetObjectUrls[asset.id]"
                        :alt="asset.revised_prompt || task.prompt"
                        class="h-full w-full object-contain transition duration-200 hover:scale-[1.01]"
                      />
                    </button>
                    <Icon v-else name="refresh" size="lg" class="animate-spin text-gray-400" />
                  </div>
                  <figcaption class="space-y-2 p-3">
                    <div class="flex items-center justify-between gap-3 text-xs text-gray-500 dark:text-dark-400">
                      <span>{{ assetSizeLabel(asset, task) }}</span>
                      <div class="flex items-center gap-1">
                        <button
                          type="button"
                          class="btn btn-ghost btn-icon h-8 w-8 text-gray-500 hover:text-primary-600 dark:text-dark-300 dark:hover:text-primary-200"
                          :disabled="!assetObjectUrls[asset.id]"
                          :title="t('imageStudio.viewOriginal')"
                          @click="openImageViewer(task, asset)"
                        >
                          <Icon name="eye" size="sm" />
                        </button>
                        <button
                          type="button"
                          class="btn btn-ghost btn-icon h-8 w-8 text-gray-500 hover:text-primary-600 dark:text-dark-300 dark:hover:text-primary-200"
                          :disabled="!assetObjectUrls[asset.id]"
                          :title="t('imageStudio.download')"
                          @click="downloadAsset(task, asset)"
                        >
                          <Icon name="download" size="sm" />
                        </button>
                      </div>
                    </div>
                    <p v-if="asset.revised_prompt" class="max-h-10 overflow-hidden text-xs leading-5 text-gray-500 dark:text-dark-400">
                      {{ asset.revised_prompt }}
                    </p>
                  </figcaption>
                </figure>
              </div>
            </article>

            <div v-if="canLoadMore" class="p-5 text-center">
              <button class="btn btn-secondary" :disabled="loadingMore" @click="loadTasks(false)">
                <Icon name="refresh" size="sm" :class="loadingMore ? 'animate-spin' : ''" />
                <span>{{ t('imageStudio.loadMore') }}</span>
              </button>
            </div>
          </div>
        </section>
      </div>

      <Teleport to="body">
        <div
          v-if="promptOptimization.open"
          class="fixed inset-0 z-[75] flex items-center justify-center bg-gray-950/70 px-4 py-6 backdrop-blur-sm"
          @click.self="closePromptOptimization"
        >
          <div class="w-full max-w-2xl rounded-xl bg-white shadow-2xl dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3 border-b border-gray-100 px-5 py-4 dark:border-dark-700">
              <div class="min-w-0">
                <h3 class="truncate text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('imageStudio.promptOptimizationTitle') }}
                </h3>
                <p class="mt-1 truncate text-xs text-gray-500 dark:text-dark-400">
                  {{ promptOptimization.model }}
                </p>
              </div>
              <button
                type="button"
                class="btn btn-ghost btn-icon text-gray-500 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
                :title="t('imageStudio.close')"
                @click="closePromptOptimization"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
            <div class="p-5">
              <textarea
                class="input min-h-44 resize-y text-sm leading-6"
                :value="promptOptimization.prompt"
                readonly
              />
              <div class="mt-5 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                <button type="button" class="btn btn-secondary" :disabled="optimizingPrompt" @click="closePromptOptimization">
                  <Icon name="x" size="sm" />
                  <span>{{ t('imageStudio.cancelPromptOptimization') }}</span>
                </button>
                <button type="button" class="btn btn-secondary" :disabled="optimizingPrompt" @click="regenerateOptimizedPrompt">
                  <Icon name="refresh" size="sm" :class="optimizingPrompt ? 'animate-spin' : ''" />
                  <span>{{ optimizingPrompt ? t('imageStudio.optimizingPrompt') : t('imageStudio.regeneratePrompt') }}</span>
                </button>
                <button type="button" class="btn btn-primary" :disabled="optimizingPrompt" @click="useOptimizedPrompt">
                  <Icon name="check" size="sm" />
                  <span>{{ t('imageStudio.useOptimizedPrompt') }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div
          v-if="imageViewer.asset && imageViewer.url"
          class="fixed inset-0 z-[70] bg-gray-950/95 text-white"
          @click.self="closeImageViewer"
        >
          <div class="absolute left-4 right-4 top-4 z-10 flex items-center justify-between gap-3">
            <div class="min-w-0">
              <p class="truncate text-sm font-medium">{{ imageViewerTitle }}</p>
              <p v-if="imageViewerPrompt" class="truncate text-xs text-white/60">{{ imageViewerPrompt }}</p>
            </div>
            <div class="flex shrink-0 items-center gap-2 rounded-lg bg-white/10 p-1 backdrop-blur">
              <button
                type="button"
                class="btn btn-ghost btn-icon h-9 w-9 text-white hover:bg-white/10"
                :title="t('imageStudio.zoomOut')"
                :disabled="imageViewer.scale <= viewerMinScale"
                @click="zoomImageViewer(-0.25)"
              >
                <Icon name="minus" size="sm" />
              </button>
              <span class="min-w-14 text-center text-xs font-medium text-white/80">{{ imageViewerScaleLabel }}</span>
              <button
                type="button"
                class="btn btn-ghost btn-icon h-9 w-9 text-white hover:bg-white/10"
                :title="t('imageStudio.zoomIn')"
                :disabled="imageViewer.scale >= viewerMaxScale"
                @click="zoomImageViewer(0.25)"
              >
                <Icon name="plus" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-icon h-9 w-9 text-white hover:bg-white/10"
                :title="t('imageStudio.resetView')"
                @click="resetImageViewer"
              >
                <Icon name="refresh" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-icon h-9 w-9 text-white hover:bg-white/10"
                :title="t('imageStudio.download')"
                @click="downloadImageViewerAsset"
              >
                <Icon name="download" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-icon h-9 w-9 text-white hover:bg-white/10"
                :title="t('imageStudio.close')"
                @click="closeImageViewer"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          </div>

          <div
            class="flex h-full w-full items-center justify-center overflow-hidden px-4 pt-20"
            :class="imageViewer.dragging ? 'cursor-grabbing' : 'cursor-grab'"
            style="touch-action: none;"
            @wheel.prevent="handleImageViewerWheel"
            @pointerdown="startImageViewerDrag"
            @pointermove="moveImageViewerDrag"
            @pointerup="stopImageViewerDrag"
            @pointercancel="stopImageViewerDrag"
          >
            <img
              :src="imageViewer.url"
              :alt="imageViewerPrompt || imageViewerTitle"
              class="max-h-[86vh] max-w-[94vw] select-none object-contain will-change-transform"
              :style="imageViewerImageStyle"
              draggable="false"
            />
          </div>
        </div>
      </Teleport>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import imageStudioAPI, {
  type ImageStudioAsset,
  type ImageStudioKey,
  type ImageStudioMode,
  type ImageStudioStatus,
  type ImageStudioTask
} from '@/api/imageStudio'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const ratios = ['1:1', '3:2', '2:3', '4:3', '3:4', '5:4', '4:5', '16:9', '9:16', '21:9'] as const
const resolutions = ['1K', '2K', '4K'] as const
const qualities = ['high', 'medium', 'low', 'auto'] as const
const counts = [1, 2, 3, 4] as const

const modeOptions: Array<{ value: ImageStudioMode; labelKey: string; icon: 'sparkles' | 'upload' }> = [
  { value: 'text_to_image', labelKey: 'imageStudio.modes.textToImage', icon: 'sparkles' },
  { value: 'image_to_image', labelKey: 'imageStudio.modes.imageToImage', icon: 'upload' }
]

const keys = ref<ImageStudioKey[]>([])
const selectedKeyID = ref(0)
const mode = ref<ImageStudioMode>('text_to_image')
const model = ref('gpt-image-2')
const prompt = ref('')
const ratio = ref<(typeof ratios)[number]>('1:1')
const resolution = ref<(typeof resolutions)[number]>('1K')
const quality = ref<(typeof qualities)[number]>('high')
const count = ref<(typeof counts)[number]>(1)
const referenceFile = ref<File | null>(null)
const referencePreviewUrl = ref('')
const referenceInput = ref<HTMLInputElement | null>(null)
const formError = ref('')

const tasks = ref<ImageStudioTask[]>([])
const currentPage = ref(1)
const totalPages = ref(1)
const loadingKeys = ref(false)
const loadingTasks = ref(false)
const loadingMore = ref(false)
const submitting = ref(false)
const optimizingPrompt = ref(false)
const deletingTaskIDs = ref<Record<string, boolean>>({})
const assetObjectUrls = ref<Record<number, string>>({})
const promptOptimization = ref({
  open: false,
  prompt: '',
  sourcePrompt: '',
  model: '',
  variant: 0
})
const imageViewer = ref({
  task: null as ImageStudioTask | null,
  asset: null as ImageStudioAsset | null,
  url: '',
  scale: 1,
  translateX: 0,
  translateY: 0,
  dragging: false,
  dragStartX: 0,
  dragStartY: 0,
  startTranslateX: 0,
  startTranslateY: 0
})

const pageSize = 20
const loadingAssetIDs = new Set<number>()
const viewerMinScale = 0.25
const viewerMaxScale = 5
let pollTimer: number | undefined

const canLoadMore = computed(() => currentPage.value < totalPages.value)
const imageViewerScaleLabel = computed(() => `${Math.round(imageViewer.value.scale * 100)}%`)
const imageViewerTitle = computed(() => {
  const asset = imageViewer.value.asset
  if (!asset) return ''
  return `${asset.width || '-'}x${asset.height || '-'} · ${formatBytes(asset.size_bytes)}`
})
const imageViewerPrompt = computed(() => imageViewer.value.asset?.revised_prompt || imageViewer.value.task?.prompt || '')
const imageViewerImageStyle = computed(() => ({
  transform: `translate(${imageViewer.value.translateX}px, ${imageViewer.value.translateY}px) scale(${imageViewer.value.scale})`
}))

async function refreshAll() {
  await Promise.all([loadKeys(), loadTasks(true)])
}

async function loadKeys() {
  loadingKeys.value = true
  try {
    keys.value = await imageStudioAPI.listKeys()
    if (keys.value.length === 0) {
      selectedKeyID.value = 0
    } else if (!keys.value.some((key) => key.id === selectedKeyID.value)) {
      selectedKeyID.value = keys.value[0].id
    }
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('imageStudio.loadKeysFailed')))
  } finally {
    loadingKeys.value = false
  }
}

async function loadTasks(reset = true) {
  if (reset) {
    loadingTasks.value = true
  } else {
    loadingMore.value = true
  }
  try {
    const nextPage = reset ? 1 : currentPage.value + 1
    const result = await imageStudioAPI.listTasks(nextPage, pageSize)
    if (reset) {
      tasks.value = result.items
    } else {
      const existingIDs = new Set(tasks.value.map((task) => task.task_id))
      tasks.value = [...tasks.value, ...result.items.filter((task) => !existingIDs.has(task.task_id))]
    }
    currentPage.value = result.page
    totalPages.value = result.pages || 1
    hydrateAssetsForTasks(result.items)
    syncPolling()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('imageStudio.loadFailed')))
  } finally {
    loadingTasks.value = false
    loadingMore.value = false
  }
}

async function handleSubmit() {
  formError.value = ''
  if (optimizingPrompt.value) {
    return
  }
  if (!selectedKeyID.value) {
    formError.value = t('imageStudio.selectKeyFirst')
    return
  }
  if (!prompt.value.trim()) {
    formError.value = t('imageStudio.promptRequired')
    return
  }
  if (mode.value === 'image_to_image' && !referenceFile.value) {
    formError.value = t('imageStudio.referenceRequired')
    return
  }

  const formData = new FormData()
  formData.append('api_key_id', String(selectedKeyID.value))
  formData.append('model', model.value.trim() || 'gpt-image-2')
  formData.append('prompt', prompt.value.trim())
  formData.append('ratio', ratio.value)
  formData.append('resolution', resolution.value)
  formData.append('quality', quality.value)
  formData.append('count', String(count.value))
  if (mode.value === 'image_to_image' && referenceFile.value) {
    formData.append('image', referenceFile.value)
  }

  submitting.value = true
  try {
    const task = await imageStudioAPI.createTask(formData)
    upsertTask(task)
    appStore.showSuccess(t('imageStudio.created'))
  } catch (err: unknown) {
    formError.value = extractApiErrorMessage(err, t('imageStudio.submitFailed'))
    appStore.showError(formError.value)
  } finally {
    submitting.value = false
  }
}

async function handleOptimizePrompt(regenerate: boolean) {
  formError.value = ''
  if (!selectedKeyID.value) {
    formError.value = t('imageStudio.selectKeyFirst')
    return
  }
  if (!prompt.value.trim()) {
    formError.value = t('imageStudio.promptRequired')
    return
  }
  const nextVariant = regenerate ? promptOptimization.value.variant + 1 : 1
  optimizingPrompt.value = true
  try {
    const result = await imageStudioAPI.optimizePrompt({
      api_key_id: selectedKeyID.value,
      prompt: prompt.value.trim(),
      ratio: ratio.value,
      resolution: resolution.value,
      quality: quality.value,
      previous_prompt: regenerate ? promptOptimization.value.prompt : '',
      variant: nextVariant
    })
    promptOptimization.value = {
      open: true,
      prompt: result.prompt,
      sourcePrompt: result.source_prompt,
      model: result.model,
      variant: nextVariant
    }
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('imageStudio.promptOptimizationFailed')))
  } finally {
    optimizingPrompt.value = false
  }
}

function useOptimizedPrompt() {
  if (!promptOptimization.value.prompt) return
  prompt.value = promptOptimization.value.prompt
  closePromptOptimization()
  appStore.showSuccess(t('imageStudio.promptOptimizationApplied'))
}

function closePromptOptimization() {
  promptOptimization.value = {
    open: false,
    prompt: '',
    sourcePrompt: '',
    model: '',
    variant: 0
  }
}

function regenerateOptimizedPrompt() {
  void handleOptimizePrompt(true)
}

function handleReferenceChange(event: Event) {
  const input = event.target as HTMLInputElement
  setReferenceFile(input.files?.[0] ?? null)
}

function setReferenceFile(file: File | null) {
  if (referencePreviewUrl.value) {
    window.URL.revokeObjectURL(referencePreviewUrl.value)
  }
  referenceFile.value = file
  referencePreviewUrl.value = file ? window.URL.createObjectURL(file) : ''
}

function clearReference() {
  setReferenceFile(null)
  if (referenceInput.value) {
    referenceInput.value.value = ''
  }
}

function upsertTask(task: ImageStudioTask) {
  const index = tasks.value.findIndex((item) => item.task_id === task.task_id)
  if (index >= 0) {
    tasks.value.splice(index, 1, task)
  } else {
    tasks.value.unshift(task)
  }
  hydrateAssetsForTasks([task])
  syncPolling()
}

async function pollActiveTasks() {
  const activeIDs = tasks.value.filter((task) => isActiveStatus(task.status)).map((task) => task.task_id)
  if (activeIDs.length === 0) {
    stopPolling()
    return
  }

  const updates = await Promise.all(
    activeIDs.map((taskID) =>
      imageStudioAPI.getTask(taskID).catch((err: unknown) => {
        console.error('Failed to poll image task:', err)
        return null
      })
    )
  )
  updates.forEach((task) => {
    if (task) upsertTask(task)
  })
  syncPolling()
}

function syncPolling() {
  if (tasks.value.some((task) => isActiveStatus(task.status))) {
    startPolling()
  } else {
    stopPolling()
  }
}

function startPolling() {
  if (pollTimer !== undefined) return
  pollTimer = window.setInterval(() => {
    void pollActiveTasks()
  }, 1500)
}

function stopPolling() {
  if (pollTimer === undefined) return
  window.clearInterval(pollTimer)
  pollTimer = undefined
}

async function handleDelete(task: ImageStudioTask) {
  if (!window.confirm(t('imageStudio.deleteConfirm'))) return
  deletingTaskIDs.value = { ...deletingTaskIDs.value, [task.task_id]: true }
  try {
    await imageStudioAPI.deleteTask(task.task_id)
    if (imageViewer.value.task?.task_id === task.task_id) {
      closeImageViewer()
    }
    revokeAssetUrls(task.assets)
    tasks.value = tasks.value.filter((item) => item.task_id !== task.task_id)
    syncPolling()
    appStore.showSuccess(t('imageStudio.deleted'))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('imageStudio.deleteFailed')))
  } finally {
    const next = { ...deletingTaskIDs.value }
    delete next[task.task_id]
    deletingTaskIDs.value = next
  }
}

function hydrateAssetsForTasks(sourceTasks: ImageStudioTask[]) {
  for (const task of sourceTasks) {
    for (const asset of task.assets || []) {
      void loadAssetUrl(asset)
    }
  }
}

async function loadAssetUrl(asset: ImageStudioAsset) {
  if (assetObjectUrls.value[asset.id] || loadingAssetIDs.has(asset.id)) return
  loadingAssetIDs.add(asset.id)
  try {
    const blob = await imageStudioAPI.fetchAssetBlob(asset.id)
    assetObjectUrls.value = {
      ...assetObjectUrls.value,
      [asset.id]: window.URL.createObjectURL(blob)
    }
  } catch (err: unknown) {
    console.error('Failed to load image asset:', err)
  } finally {
    loadingAssetIDs.delete(asset.id)
  }
}

function revokeAssetUrls(assets: ImageStudioAsset[]) {
  const next = { ...assetObjectUrls.value }
  for (const asset of assets) {
    const objectUrl = next[asset.id]
    if (objectUrl) {
      window.URL.revokeObjectURL(objectUrl)
      delete next[asset.id]
    }
  }
  assetObjectUrls.value = next
}

function openImageViewer(task: ImageStudioTask, asset: ImageStudioAsset) {
  const url = assetObjectUrls.value[asset.id]
  if (!url) return
  imageViewer.value = {
    task,
    asset,
    url,
    scale: 1,
    translateX: 0,
    translateY: 0,
    dragging: false,
    dragStartX: 0,
    dragStartY: 0,
    startTranslateX: 0,
    startTranslateY: 0
  }
}

function closeImageViewer() {
  imageViewer.value = {
    task: null,
    asset: null,
    url: '',
    scale: 1,
    translateX: 0,
    translateY: 0,
    dragging: false,
    dragStartX: 0,
    dragStartY: 0,
    startTranslateX: 0,
    startTranslateY: 0
  }
}

function resetImageViewer() {
  imageViewer.value = {
    ...imageViewer.value,
    scale: 1,
    translateX: 0,
    translateY: 0,
    dragging: false
  }
}

function zoomImageViewer(delta: number) {
  const nextScale = clamp(imageViewer.value.scale + delta, viewerMinScale, viewerMaxScale)
  imageViewer.value = {
    ...imageViewer.value,
    scale: nextScale,
    translateX: nextScale === 1 ? 0 : imageViewer.value.translateX,
    translateY: nextScale === 1 ? 0 : imageViewer.value.translateY
  }
}

function handleImageViewerWheel(event: WheelEvent) {
  const delta = event.deltaY > 0 ? -0.15 : 0.15
  zoomImageViewer(delta)
}

function startImageViewerDrag(event: PointerEvent) {
  if (!imageViewer.value.asset) return
  const target = event.currentTarget as HTMLElement
  target.setPointerCapture?.(event.pointerId)
  imageViewer.value = {
    ...imageViewer.value,
    dragging: true,
    dragStartX: event.clientX,
    dragStartY: event.clientY,
    startTranslateX: imageViewer.value.translateX,
    startTranslateY: imageViewer.value.translateY
  }
}

function moveImageViewerDrag(event: PointerEvent) {
  if (!imageViewer.value.dragging) return
  imageViewer.value = {
    ...imageViewer.value,
    translateX: imageViewer.value.startTranslateX + event.clientX - imageViewer.value.dragStartX,
    translateY: imageViewer.value.startTranslateY + event.clientY - imageViewer.value.dragStartY
  }
}

function stopImageViewerDrag(event?: PointerEvent) {
  if (event?.currentTarget instanceof HTMLElement) {
    event.currentTarget.releasePointerCapture?.(event.pointerId)
  }
  imageViewer.value = {
    ...imageViewer.value,
    dragging: false
  }
}

function downloadImageViewerAsset() {
  if (!imageViewer.value.task || !imageViewer.value.asset) return
  downloadAsset(imageViewer.value.task, imageViewer.value.asset)
}

function downloadAsset(task: ImageStudioTask, asset: ImageStudioAsset) {
  const url = assetObjectUrls.value[asset.id]
  if (!url) return
  const link = document.createElement('a')
  link.href = url
  link.download = imageStudioDownloadName(task, asset)
  document.body.appendChild(link)
  link.click()
  link.remove()
}

function imageStudioDownloadName(task: ImageStudioTask, asset: ImageStudioAsset) {
  const ext = imageStudioExtensionFromMime(asset.mime_type)
  return `${task.task_id}_${String(asset.seq + 1).padStart(2, '0')}_${task.size}${ext}`
}

function imageStudioExtensionFromMime(mimeType: string) {
  const normalized = (mimeType || '').toLowerCase()
  if (normalized.includes('jpeg') || normalized.includes('jpg')) return '.jpg'
  if (normalized.includes('webp')) return '.webp'
  if (normalized.includes('gif')) return '.gif'
  return '.png'
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function keyOptionLabel(key: ImageStudioKey) {
  return key.group_name ? `${key.name} - ${key.group_name}` : key.name
}

function outputAssets(task: ImageStudioTask) {
  return (task.assets || []).slice().sort((a, b) => a.seq - b.seq)
}

function isActiveStatus(status: ImageStudioStatus) {
  return status === 'pending' || status === 'running'
}

function statusLabel(status: ImageStudioStatus) {
  return t(`imageStudio.status.${status}`)
}

function modeLabel(value: ImageStudioMode) {
  return value === 'image_to_image' ? t('imageStudio.modes.imageToImage') : t('imageStudio.modes.textToImage')
}

function qualityLabel(value: string) {
  return t(`imageStudio.qualities.${value}`)
}

function statusBadgeClass(status: ImageStudioStatus) {
  const base = 'inline-flex items-center rounded-full px-2 py-1 text-xs font-medium'
  if (status === 'succeeded') return `${base} bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200`
  if (status === 'failed') return `${base} bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-200`
  if (status === 'running') return `${base} bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-200`
  if (status === 'canceled') return `${base} bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300`
  return `${base} bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-200`
}

function formatBytes(bytes: number) {
  if (!Number.isFinite(bytes) || bytes <= 0) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

function assetSizeLabel(asset: ImageStudioAsset, task: ImageStudioTask) {
  if (asset.width > 0 && asset.height > 0) return `${asset.width}x${asset.height}`
  return task.size
}

function formatDate(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

onMounted(() => {
  void refreshAll()
})

onUnmounted(() => {
  stopPolling()
  if (referencePreviewUrl.value) {
    window.URL.revokeObjectURL(referencePreviewUrl.value)
  }
  Object.values(assetObjectUrls.value).forEach((objectUrl) => window.URL.revokeObjectURL(objectUrl))
})
</script>
