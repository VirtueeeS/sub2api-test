import { apiClient } from '../client'

export interface CPAConfig {
  enabled: boolean
  source_dir: string
  group_id: number
  interval_seconds: number
  auto_pause_on_error: boolean
  updated_by: number
}

export interface CPAStatus {
  enabled: boolean
  running: boolean
  last_tick_at?: string
  last_success_at?: string
  last_error?: string
  current_job_id?: number
  interval_seconds: number
  source_dir: string
  group_id: number
}

export interface CPASyncJob {
  id: number
  trigger_type: string
  status: string
  files_scanned: number
  created_count: number
  updated_count: number
  failed_count: number
  error_summary?: string
  triggered_by?: number
  created_at?: string
  updated_at?: string
  started_at?: string
  finished_at?: string
}

export interface CPASyncRecord {
  id: number
  job_id: number
  file_path: string
  action: string
  account_id?: number
  error_message?: string
  processed_at?: string
}

export interface CPAFileState {
  id: number
  file_path: string
  sha256: string
  size: number
  last_result: string
  linked_account_id?: number
  last_error?: string
  last_seen_at?: string
  updated_at?: string
}

export interface CPAPaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface CPASyncStats {
  files_scanned: number
  created_count: number
  updated_count: number
  failed_count: number
}

export interface CPASyncResult {
  job_id: number
  stats: CPASyncStats
}

export interface CPAOverview {
  config: CPAConfig
  status: CPAStatus
  latest_job: CPASyncJob | null
}

export interface CPAJobsQuery {
  page?: number
  page_size?: number
}

export interface CPAFilesQuery extends CPAJobsQuery {
  result?: string
  search?: string
}

interface RawCPASyncStats {
  files_scanned?: number
  created_count?: number
  updated_count?: number
  failed_count?: number
  FilesScanned?: number
  CreatedCount?: number
  UpdatedCount?: number
  FailedCount?: number
}

interface RawCPASyncResult {
  job_id?: number
  JobID?: number
  stats?: RawCPASyncStats
  Stats?: RawCPASyncStats
}

function unwrapConfigResponse(payload: CPAConfig | { config: CPAConfig }): CPAConfig {
  return 'config' in payload ? payload.config : payload
}

function normalizeStats(stats?: RawCPASyncStats): CPASyncStats {
  return {
    files_scanned: stats?.files_scanned ?? stats?.FilesScanned ?? 0,
    created_count: stats?.created_count ?? stats?.CreatedCount ?? 0,
    updated_count: stats?.updated_count ?? stats?.UpdatedCount ?? 0,
    failed_count: stats?.failed_count ?? stats?.FailedCount ?? 0
  }
}

export async function getOverview(): Promise<CPAOverview> {
  const { data } = await apiClient.get<CPAOverview>('/admin/cpa/overview')
  return data
}

export async function updateConfig(request: CPAConfig): Promise<CPAConfig> {
  const { data } = await apiClient.put<CPAConfig | { config: CPAConfig }>('/admin/cpa/config', request)
  return unwrapConfigResponse(data)
}

export async function runNow(): Promise<CPASyncResult> {
  const { data } = await apiClient.post<RawCPASyncResult>('/admin/cpa/run')

  return {
    job_id: data.job_id ?? data.JobID ?? 0,
    stats: normalizeStats(data.stats ?? data.Stats)
  }
}

export async function listJobs(params?: CPAJobsQuery): Promise<CPAPaginatedResponse<CPASyncJob>> {
  const { data } = await apiClient.get<CPAPaginatedResponse<CPASyncJob>>('/admin/cpa/jobs', {
    params
  })
  return data
}

export async function listRecords(
  jobID: number,
  params?: CPAJobsQuery
): Promise<CPAPaginatedResponse<CPASyncRecord>> {
  const { data } = await apiClient.get<CPAPaginatedResponse<CPASyncRecord>>(
    `/admin/cpa/jobs/${jobID}/records`,
    {
      params
    }
  )
  return data
}

export async function listFiles(params?: CPAFilesQuery): Promise<CPAPaginatedResponse<CPAFileState>> {
  const { data } = await apiClient.get<CPAPaginatedResponse<CPAFileState>>('/admin/cpa/files', {
    params
  })
  return data
}

export const cpaAPI = {
  getOverview,
  updateConfig,
  runNow,
  listJobs,
  listRecords,
  listFiles
}

export default cpaAPI
