import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, put, post } = vi.hoisted(() => ({
  get: vi.fn(),
  put: vi.fn(),
  post: vi.fn()
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    put,
    post
  }
}))

import cpaAPI, {
  getOverview,
  listFiles,
  listJobs,
  listRecords,
  runNow,
  updateConfig
} from '../cpa'

describe('cpa admin api', () => {
  beforeEach(() => {
    get.mockReset()
    put.mockReset()
    post.mockReset()
  })

  it('loads overview data from the admin CPA endpoint', async () => {
    const overview = {
      config: {
        enabled: true,
        source_dir: '/root/cpa/auths',
        group_id: 7,
        interval_seconds: 60,
        auto_pause_on_error: false,
        updated_by: 42
      },
      status: {
        enabled: true,
        running: false,
        interval_seconds: 60,
        source_dir: '/root/cpa/auths',
        group_id: 7
      },
      latest_job: null
    }
    get.mockResolvedValue({ data: overview })

    await expect(getOverview()).resolves.toEqual(overview)
    expect(get).toHaveBeenCalledWith('/admin/cpa/overview')
  })

  it('saves config and unwraps the config payload', async () => {
    const request = {
      enabled: true,
      source_dir: '/srv/cpa',
      group_id: 9,
      interval_seconds: 30,
      auto_pause_on_error: true,
      updated_by: 0
    }
    put.mockResolvedValue({
      data: {
        config: request
      }
    })

    await expect(updateConfig(request)).resolves.toEqual(request)
    expect(put).toHaveBeenCalledWith('/admin/cpa/config', request)
  })

  it('normalizes manual run results from Go struct casing', async () => {
    post.mockResolvedValue({
      data: {
        JobID: 101,
        Stats: {
          FilesScanned: 2,
          CreatedCount: 1,
          UpdatedCount: 0,
          FailedCount: 1
        }
      }
    })

    await expect(runNow()).resolves.toEqual({
      job_id: 101,
      stats: {
        files_scanned: 2,
        created_count: 1,
        updated_count: 0,
        failed_count: 1
      }
    })
    expect(post).toHaveBeenCalledWith('/admin/cpa/run')
  })

  it('lists jobs with pagination params', async () => {
    const response = {
      items: [{ id: 9 }],
      total: 1,
      page: 2,
      page_size: 5,
      pages: 1
    }
    get.mockResolvedValue({ data: response })

    await expect(listJobs({ page: 2, page_size: 5 })).resolves.toEqual(response)
    expect(get).toHaveBeenCalledWith('/admin/cpa/jobs', {
      params: {
        page: 2,
        page_size: 5
      }
    })
  })

  it('lists records for a selected job', async () => {
    const response = {
      items: [{ id: 1, job_id: 9 }],
      total: 1,
      page: 1,
      page_size: 10,
      pages: 1
    }
    get.mockResolvedValue({ data: response })

    await expect(listRecords(9, { page: 1, page_size: 10 })).resolves.toEqual(response)
    expect(get).toHaveBeenCalledWith('/admin/cpa/jobs/9/records', {
      params: {
        page: 1,
        page_size: 10
      }
    })
  })

  it('lists file states with result and search filters', async () => {
    const response = {
      items: [{ id: 7, file_path: '/tmp/broken.json' }],
      total: 1,
      page: 3,
      page_size: 20,
      pages: 1
    }
    get.mockResolvedValue({ data: response })

    await expect(
      listFiles({ page: 3, page_size: 20, result: 'failed', search: 'broken' })
    ).resolves.toEqual(response)
    expect(get).toHaveBeenCalledWith('/admin/cpa/files', {
      params: {
        page: 3,
        page_size: 20,
        result: 'failed',
        search: 'broken'
      }
    })
  })

  it('exports a convenient API object', () => {
    expect(cpaAPI.getOverview).toBe(getOverview)
    expect(cpaAPI.updateConfig).toBe(updateConfig)
    expect(cpaAPI.runNow).toBe(runNow)
    expect(cpaAPI.listJobs).toBe(listJobs)
    expect(cpaAPI.listRecords).toBe(listRecords)
    expect(cpaAPI.listFiles).toBe(listFiles)
  })
})
