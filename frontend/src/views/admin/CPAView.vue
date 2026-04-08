<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
              {{ t('admin.cpa.title') }}
            </h1>
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.description') }}
            </p>
          </div>

          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="loadingOverview || loadingJobs || loadingFiles || loadingRecords"
              @click="refreshAll"
            >
              {{ t('admin.cpa.actions.reload') }}
            </button>
            <button
              type="button"
              data-testid="cpa-run-now"
              class="btn btn-primary btn-sm"
              :disabled="runningNow"
              @click="runSyncNow"
            >
              {{ runningNow ? t('common.loading') : t('admin.cpa.actions.runNow') }}
            </button>
          </div>
        </div>

        <div class="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <div class="rounded-2xl border border-gray-200 bg-white/70 p-4 dark:border-dark-700 dark:bg-dark-900/60">
            <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.summary.enabled') }}
            </div>
            <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ configForm.enabled ? t('common.enabled') : t('common.disabled') }}
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white/70 p-4 dark:border-dark-700 dark:bg-dark-900/60">
            <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.summary.running') }}
            </div>
            <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ overview?.status.running ? t('common.yes') : t('common.no') }}
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white/70 p-4 dark:border-dark-700 dark:bg-dark-900/60">
            <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.summary.interval') }}
            </div>
            <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ configForm.interval_seconds }}s
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white/70 p-4 dark:border-dark-700 dark:bg-dark-900/60">
            <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.summary.sourceDir') }}
            </div>
            <div class="mt-2 break-all font-mono text-sm text-gray-900 dark:text-white">
              {{ configForm.source_dir || '-' }}
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white/70 p-4 dark:border-dark-700 dark:bg-dark-900/60">
            <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.summary.group') }}
            </div>
            <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ configForm.group_id || '-' }}
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white/70 p-4 dark:border-dark-700 dark:bg-dark-900/60">
            <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.summary.latestJob') }}
            </div>
            <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
              #{{ overview?.latest_job?.id ?? selectedJobID ?? '-' }}
            </div>
          </div>
        </div>
      </section>

      <section class="card p-6">
        <div class="mb-4">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.cpa.config.title') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.cpa.config.description') }}
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="flex items-center gap-3 rounded-xl border border-gray-200 px-4 py-3 text-sm text-gray-700 dark:border-dark-700 dark:text-gray-300">
            <input v-model="configForm.enabled" type="checkbox" />
            <span>{{ t('admin.cpa.config.enabled') }}</span>
          </label>

          <label class="flex items-center gap-3 rounded-xl border border-gray-200 px-4 py-3 text-sm text-gray-700 dark:border-dark-700 dark:text-gray-300">
            <input v-model="configForm.auto_pause_on_error" type="checkbox" />
            <span>{{ t('admin.cpa.config.autoPauseOnError') }}</span>
          </label>

          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.cpa.config.sourceDir') }}
            </label>
            <input
              v-model="configForm.source_dir"
              data-testid="cpa-source-dir"
              type="text"
              class="input w-full"
              :placeholder="t('admin.cpa.config.sourceDir')"
            />
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.cpa.config.groupId') }}
            </label>
            <input
              v-model.number="configForm.group_id"
              type="number"
              min="0"
              class="input w-full"
              :placeholder="t('admin.cpa.config.groupId')"
            />
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.cpa.config.intervalSeconds') }}
            </label>
            <input
              v-model.number="configForm.interval_seconds"
              type="number"
              min="1"
              class="input w-full"
              :placeholder="t('admin.cpa.config.intervalSeconds')"
            />
          </div>
        </div>

        <div class="mt-4 flex justify-end">
          <button type="button" class="btn btn-primary btn-sm" :disabled="savingConfig" @click="saveConfig">
            {{ savingConfig ? t('common.loading') : t('admin.cpa.actions.saveConfig') }}
          </button>
        </div>
      </section>

      <section class="grid gap-6 xl:grid-cols-2">
        <div class="card overflow-hidden">
          <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.cpa.jobs.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.jobs.description') }}
            </p>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full min-w-[720px] text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-700 dark:text-gray-400">
                  <th class="px-6 py-3">#</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.jobs.triggerType') }}</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.jobs.status') }}</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.jobs.stats') }}</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.jobs.updatedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="job in jobs.items"
                  :key="job.id"
                  class="cursor-pointer border-b border-gray-100 transition hover:bg-gray-50/70 dark:border-dark-800 dark:hover:bg-dark-900/40"
                  :class="selectedJobID === job.id ? 'bg-primary-50/70 dark:bg-primary-900/20' : ''"
                  @click="selectJob(job.id)"
                >
                  <td class="px-6 py-4 font-medium text-gray-900 dark:text-white">#{{ job.id }}</td>
                  <td class="px-6 py-4">{{ job.trigger_type }}</td>
                  <td class="px-6 py-4">
                    <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="jobStatusClass(job.status)">
                      {{ job.status }}
                    </span>
                  </td>
                  <td class="px-6 py-4 text-xs text-gray-600 dark:text-gray-300">
                    {{ job.files_scanned }}/{{ job.created_count }}/{{ job.updated_count }}/{{ job.failed_count }}
                  </td>
                  <td class="px-6 py-4 text-xs text-gray-500 dark:text-gray-400">
                    {{ formatDateTime(job.updated_at || job.finished_at || job.started_at || job.created_at) }}
                  </td>
                </tr>
                <tr v-if="!loadingJobs && jobs.items.length === 0">
                  <td colspan="5" class="px-6 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                    {{ t('admin.cpa.jobs.noData') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <Pagination
            v-if="jobs.total > 0"
            :total="jobs.total"
            :page="jobs.page"
            :page-size="DEFAULT_PAGE_SIZE"
            :show-page-size-selector="false"
            @update:page="handleJobsPageChange"
          />
        </div>

        <div class="card overflow-hidden">
          <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.cpa.records.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.cpa.records.description') }}
            </p>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full min-w-[720px] text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-700 dark:text-gray-400">
                  <th class="px-6 py-3">#</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.records.filePath') }}</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.records.action') }}</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.records.accountId') }}</th>
                  <th class="px-6 py-3">{{ t('admin.cpa.records.processedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="record in records.items" :key="record.id" class="border-b border-gray-100 dark:border-dark-800">
                  <td class="px-6 py-4 font-medium text-gray-900 dark:text-white">#{{ record.id }}</td>
                  <td class="px-6 py-4">
                    <div class="max-w-[320px] break-all font-mono text-xs text-gray-600 dark:text-gray-300">
                      {{ record.file_path }}
                    </div>
                  </td>
                  <td class="px-6 py-4">{{ record.action }}</td>
                  <td class="px-6 py-4">{{ record.account_id ?? '-' }}</td>
                  <td class="px-6 py-4 text-xs text-gray-500 dark:text-gray-400">
                    {{ formatDateTime(record.processed_at) }}
                  </td>
                </tr>
                <tr v-if="!loadingRecords && selectedJobID === null">
                  <td colspan="5" class="px-6 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                    {{ t('admin.cpa.records.noSelection') }}
                  </td>
                </tr>
                <tr v-else-if="!loadingRecords && records.items.length === 0">
                  <td colspan="5" class="px-6 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                    {{ t('admin.cpa.records.noData') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <Pagination
            v-if="records.total > 0"
            :total="records.total"
            :page="records.page"
            :page-size="DEFAULT_PAGE_SIZE"
            :show-page-size-selector="false"
            @update:page="handleRecordsPageChange"
          />
        </div>
      </section>

      <section class="card overflow-hidden">
        <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-700">
          <div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.cpa.files.title') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.cpa.files.description') }}
              </p>
            </div>

            <div class="grid gap-3 md:grid-cols-[minmax(0,240px)_180px_auto]">
              <input
                v-model="fileFilters.search"
                type="text"
                class="input w-full"
                :placeholder="t('admin.cpa.files.searchPlaceholder')"
                @keyup.enter="submitFileFilters"
              />

              <select v-model="fileFilters.result" class="input w-full">
                <option value="">{{ t('admin.cpa.files.allResults') }}</option>
                <option v-for="option in fileResultOptions" :key="option" :value="option">
                  {{ option }}
                </option>
              </select>

              <div class="flex gap-2">
                <button type="button" class="btn btn-secondary btn-sm" @click="submitFileFilters">
                  {{ t('admin.cpa.actions.applyFilters') }}
                </button>
                <button type="button" class="btn btn-secondary btn-sm" @click="resetFileFilters">
                  {{ t('admin.cpa.actions.resetFilters') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full min-w-[920px] text-sm">
            <thead>
              <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-700 dark:text-gray-400">
                <th class="px-6 py-3">#</th>
                <th class="px-6 py-3">{{ t('admin.cpa.files.filePath') }}</th>
                <th class="px-6 py-3">{{ t('admin.cpa.files.lastResult') }}</th>
                <th class="px-6 py-3">{{ t('admin.cpa.files.linkedAccount') }}</th>
                <th class="px-6 py-3">{{ t('admin.cpa.files.size') }}</th>
                <th class="px-6 py-3">{{ t('admin.cpa.files.updatedAt') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="file in files.items" :key="file.id" class="border-b border-gray-100 dark:border-dark-800">
                <td class="px-6 py-4 font-medium text-gray-900 dark:text-white">#{{ file.id }}</td>
                <td class="px-6 py-4">
                  <div class="max-w-[360px] break-all font-mono text-xs text-gray-600 dark:text-gray-300">
                    {{ file.file_path }}
                  </div>
                </td>
                <td class="px-6 py-4">
                  <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="fileResultClass(file.last_result)">
                    {{ file.last_result || '-' }}
                  </span>
                </td>
                <td class="px-6 py-4">{{ file.linked_account_id ?? '-' }}</td>
                <td class="px-6 py-4">{{ formatSize(file.size) }}</td>
                <td class="px-6 py-4 text-xs text-gray-500 dark:text-gray-400">
                  {{ formatDateTime(file.updated_at || file.last_seen_at) }}
                </td>
              </tr>
              <tr v-if="!loadingFiles && files.items.length === 0">
                <td colspan="6" class="px-6 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.cpa.files.noData') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <Pagination
          v-if="files.total > 0"
          :total="files.total"
          :page="files.page"
          :page-size="DEFAULT_PAGE_SIZE"
          :show-page-size-selector="false"
          @update:page="handleFilesPageChange"
        />
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type {
  CPAConfig,
  CPAFileState,
  CPAOverview,
  CPAPaginatedResponse,
  CPASyncJob,
  CPASyncRecord
} from '@/api/admin/cpa'
import Pagination from '@/components/common/Pagination.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'

const DEFAULT_PAGE_SIZE = 10

const { t } = useI18n()
const appStore = useAppStore()

const overview = ref<CPAOverview | null>(null)

const configForm = reactive<CPAConfig>(createDefaultConfig())
const selectedJobID = ref<number | null>(null)

const jobs = ref<CPAPaginatedResponse<CPASyncJob>>(createEmptyPaginated<CPASyncJob>())
const records = ref<CPAPaginatedResponse<CPASyncRecord>>(createEmptyPaginated<CPASyncRecord>())
const files = ref<CPAPaginatedResponse<CPAFileState>>(createEmptyPaginated<CPAFileState>())

const jobsPage = ref(1)
const recordsPage = ref(1)
const filesPage = ref(1)

const fileFilters = reactive({
  result: '',
  search: ''
})

const loadingOverview = ref(false)
const loadingJobs = ref(false)
const loadingRecords = ref(false)
const loadingFiles = ref(false)
const savingConfig = ref(false)
const runningNow = ref(false)

const fileResultOptions = computed(() => ['created', 'updated', 'failed', 'skipped'])

onMounted(() => {
  void initialize()
})

async function initialize() {
  await loadOverview()
  await Promise.all([loadJobs(1), loadFiles(1)])

  if (selectedJobID.value !== null) {
    await loadRecords(selectedJobID.value, 1)
  }
}

async function loadOverview(options: { preserveSelectedJob?: boolean } = {}) {
  loadingOverview.value = true
  try {
    const result = await adminAPI.cpa.getOverview()
    overview.value = result
    applyOverview(result, options.preserveSelectedJob ?? false)
  } catch (error) {
    appStore.showError(getErrorMessage(error))
  } finally {
    loadingOverview.value = false
  }
}

async function loadJobs(page = jobsPage.value) {
  jobsPage.value = page
  loadingJobs.value = true
  try {
    jobs.value = await adminAPI.cpa.listJobs({
      page,
      page_size: DEFAULT_PAGE_SIZE
    })
  } catch (error) {
    appStore.showError(getErrorMessage(error))
  } finally {
    loadingJobs.value = false
  }
}

async function loadRecords(jobID: number, page = recordsPage.value) {
  recordsPage.value = page
  loadingRecords.value = true
  try {
    records.value = await adminAPI.cpa.listRecords(jobID, {
      page,
      page_size: DEFAULT_PAGE_SIZE
    })
  } catch (error) {
    appStore.showError(getErrorMessage(error))
  } finally {
    loadingRecords.value = false
  }
}

async function loadFiles(page = filesPage.value) {
  filesPage.value = page
  loadingFiles.value = true
  try {
    files.value = await adminAPI.cpa.listFiles({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      result: fileFilters.result,
      search: fileFilters.search
    })
  } catch (error) {
    appStore.showError(getErrorMessage(error))
  } finally {
    loadingFiles.value = false
  }
}

async function saveConfig() {
  savingConfig.value = true
  try {
    const saved = await adminAPI.cpa.updateConfig({
      ...configForm
    })
    assignConfig(saved)
    appStore.showSuccess(t('admin.cpa.messages.configSaved'))
    await loadOverview({ preserveSelectedJob: true })
  } catch (error) {
    appStore.showError(getErrorMessage(error))
  } finally {
    savingConfig.value = false
  }
}

async function runSyncNow() {
  runningNow.value = true
  try {
    const result = await adminAPI.cpa.runNow()
    if (result.job_id > 0) {
      selectedJobID.value = result.job_id
      recordsPage.value = 1
    }

    appStore.showSuccess(t('admin.cpa.messages.runNowSuccess'))

    jobsPage.value = 1
    filesPage.value = 1

    await Promise.all([
      loadOverview({ preserveSelectedJob: true }),
      loadJobs(1),
      loadFiles(1)
    ])

    if (result.job_id > 0) {
      await loadRecords(result.job_id, 1)
    }
  } catch (error) {
    appStore.showError(getErrorMessage(error))
  } finally {
    runningNow.value = false
  }
}

async function refreshAll() {
  jobsPage.value = 1
  filesPage.value = 1
  recordsPage.value = 1

  await Promise.all([
    loadOverview({ preserveSelectedJob: true }),
    loadJobs(1),
    loadFiles(1)
  ])

  if (selectedJobID.value !== null) {
    await loadRecords(selectedJobID.value, 1)
  }
}

function selectJob(jobID: number) {
  selectedJobID.value = jobID
  recordsPage.value = 1
  void loadRecords(jobID, 1)
}

function handleJobsPageChange(page: number) {
  void loadJobs(page)
}

function handleRecordsPageChange(page: number) {
  if (selectedJobID.value === null) {
    return
  }

  void loadRecords(selectedJobID.value, page)
}

function handleFilesPageChange(page: number) {
  void loadFiles(page)
}

function submitFileFilters() {
  filesPage.value = 1
  void loadFiles(1)
}

function resetFileFilters() {
  fileFilters.result = ''
  fileFilters.search = ''
  filesPage.value = 1
  void loadFiles(1)
}

function applyOverview(result: CPAOverview, preserveSelectedJob: boolean) {
  assignConfig(result.config)

  if (!preserveSelectedJob || selectedJobID.value === null) {
    selectedJobID.value = pickPreferredJobID(result)
  }
}

function assignConfig(config: CPAConfig) {
  configForm.enabled = config.enabled
  configForm.source_dir = config.source_dir
  configForm.group_id = config.group_id
  configForm.interval_seconds = config.interval_seconds
  configForm.auto_pause_on_error = config.auto_pause_on_error
  configForm.updated_by = config.updated_by
}

function pickPreferredJobID(result: CPAOverview): number | null {
  return result.status.current_job_id ?? result.latest_job?.id ?? null
}

function jobStatusClass(status: string) {
  switch (status) {
    case 'succeeded':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
    case 'running':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
    case 'failed':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-gray-300'
  }
}

function fileResultClass(result: string) {
  switch (result) {
    case 'created':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
    case 'updated':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
    case 'failed':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-gray-300'
  }
}

function formatDateTime(value?: string) {
  if (!value) {
    return '-'
  }

  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function formatSize(bytes: number) {
  if (!Number.isFinite(bytes) || bytes <= 0) {
    return '0 B'
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }

  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function getErrorMessage(error: unknown) {
  return (error as { message?: string })?.message || t('errors.networkError')
}

function createDefaultConfig(): CPAConfig {
  return {
    enabled: false,
    source_dir: '',
    group_id: 0,
    interval_seconds: 60,
    auto_pause_on_error: false,
    updated_by: 0
  }
}

function createEmptyPaginated<T>(): CPAPaginatedResponse<T> {
  return {
    items: [],
    total: 0,
    page: 1,
    page_size: DEFAULT_PAGE_SIZE,
    pages: 0
  }
}
</script>
