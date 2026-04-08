import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import CPAView from '../CPAView.vue'

const { getOverview, updateConfig, runNow, listJobs, listRecords, listFiles } = vi.hoisted(() => ({
  getOverview: vi.fn(),
  updateConfig: vi.fn(),
  runNow: vi.fn(),
  listJobs: vi.fn(),
  listRecords: vi.fn(),
  listFiles: vi.fn()
}))

const { showError, showSuccess } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    cpa: {
      getOverview,
      updateConfig,
      runNow,
      listJobs,
      listRecords,
      listFiles
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const latestJob = {
  id: 88,
  trigger_type: 'auto',
  status: 'succeeded',
  files_scanned: 2,
  created_count: 1,
  updated_count: 0,
  failed_count: 0,
  triggered_by: 42,
  created_at: '2026-04-08T01:00:00Z',
  updated_at: '2026-04-08T01:00:00Z',
  started_at: '2026-04-08T01:00:00Z',
  finished_at: '2026-04-08T01:00:05Z'
}

describe('admin CPAView', () => {
  beforeEach(() => {
    getOverview.mockReset()
    updateConfig.mockReset()
    runNow.mockReset()
    listJobs.mockReset()
    listRecords.mockReset()
    listFiles.mockReset()
    showError.mockReset()
    showSuccess.mockReset()

    getOverview.mockResolvedValue({
      config: {
        enabled: true,
        source_dir: '/root/cpa/auths',
        group_id: 7,
        interval_seconds: 60,
        auto_pause_on_error: true,
        updated_by: 42
      },
      status: {
        enabled: true,
        running: false,
        interval_seconds: 60,
        source_dir: '/root/cpa/auths',
        group_id: 7,
        current_job_id: 88
      },
      latest_job: latestJob
    })
    listJobs.mockResolvedValue({
      items: [latestJob],
      total: 1,
      page: 1,
      page_size: 10,
      pages: 1
    })
    listRecords.mockResolvedValue({
      items: [
        {
          id: 1,
          job_id: 88,
          file_path: '/root/cpa/auths/demo.json',
          action: 'created',
          account_id: 123,
          processed_at: '2026-04-08T01:00:05Z'
        }
      ],
      total: 1,
      page: 1,
      page_size: 10,
      pages: 1
    })
    listFiles.mockResolvedValue({
      items: [
        {
          id: 7,
          file_path: '/root/cpa/auths/demo.json',
          sha256: 'abc',
          size: 12,
          last_result: 'created',
          linked_account_id: 123,
          updated_at: '2026-04-08T01:00:05Z'
        }
      ],
      total: 1,
      page: 1,
      page_size: 10,
      pages: 1
    })
    runNow.mockResolvedValue({
      job_id: 99,
      stats: {
        files_scanned: 1,
        created_count: 1,
        updated_count: 0,
        failed_count: 0
      }
    })
  })

  it('loads overview, jobs, files and the selected job records on mount', async () => {
    const wrapper = mount(CPAView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Pagination: true
        }
      }
    })

    await flushPromises()

    expect(getOverview).toHaveBeenCalledTimes(1)
    expect(listJobs).toHaveBeenCalledWith({ page: 1, page_size: 10 })
    expect(listFiles).toHaveBeenCalledWith({ page: 1, page_size: 10, result: '', search: '' })
    expect(listRecords).toHaveBeenCalledWith(88, { page: 1, page_size: 10 })
    expect(wrapper.find('[data-testid=\"cpa-source-dir\"]').element).toHaveProperty('value', '/root/cpa/auths')
    expect(wrapper.text()).toContain('admin.cpa.actions.runNow')
  })

  it('runs a manual sync and refreshes the tables', async () => {
    const wrapper = mount(CPAView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Pagination: true
        }
      }
    })

    await flushPromises()
    await wrapper.get('[data-testid=\"cpa-run-now\"]').trigger('click')
    await flushPromises()

    expect(runNow).toHaveBeenCalledTimes(1)
    expect(showSuccess).toHaveBeenCalledTimes(1)
    expect(listRecords).toHaveBeenLastCalledWith(99, { page: 1, page_size: 10 })
    expect(getOverview).toHaveBeenCalledTimes(2)
    expect(listJobs).toHaveBeenCalledTimes(2)
    expect(listFiles).toHaveBeenCalledTimes(2)
  })
})
