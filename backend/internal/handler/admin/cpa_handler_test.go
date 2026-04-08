package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/cpa"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCPAHandlerOverviewUpdateAndRun(t *testing.T) {
	configStore := &fakeCPAConfigStore{
		cfg: cpa.Config{
			Enabled:         true,
			SourceDir:       "/srv/cpa",
			GroupID:         9,
			IntervalSeconds: 60,
		},
	}
	runtime := &fakeCPARuntime{
		status: cpa.Status{
			Enabled:         true,
			IntervalSeconds: 60,
			SourceDir:       "/srv/cpa",
			GroupID:         9,
			CurrentJobID:    101,
		},
		runResult: cpa.SyncResult{
			JobID: 101,
			Stats: cpa.SyncStats{FilesScanned: 2, UpdatedCount: 1},
		},
	}
	store := &fakeCPAStore{
		latestJob: &ent.CPASyncJob{
			ID:          101,
			TriggerType: cpa.TriggerTypeAuto,
			Status:      cpa.JobStatusSucceeded,
			FilesScanned: 2,
		},
	}
	router := setupCPAHandlerRouter(configStore, runtime, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/cpa/overview", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var overview struct {
		Config    cpa.Config       `json:"config"`
		Status    cpa.Status       `json:"status"`
		LatestJob *ent.CPASyncJob  `json:"latest_job"`
	}
	decodeSuccessData(t, rec, &overview)
	require.Equal(t, "/srv/cpa", overview.Config.SourceDir)
	require.Equal(t, 60, overview.Status.IntervalSeconds)
	require.NotNil(t, overview.LatestJob)
	require.EqualValues(t, 101, overview.LatestJob.ID)

	updateBody := bytes.NewBufferString(`{"enabled":false,"source_dir":"/srv/next-cpa","group_id":3,"interval_seconds":30,"auto_pause_on_error":true}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/cpa/config", updateBody)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var updated struct {
		Config cpa.Config `json:"config"`
	}
	decodeSuccessData(t, rec, &updated)
	require.Equal(t, "/srv/next-cpa", updated.Config.SourceDir)
	require.Equal(t, 30, updated.Config.IntervalSeconds)
	require.NotNil(t, configStore.saved)
	require.EqualValues(t, 42, configStore.saved.UpdatedBy)
	require.Equal(t, 1, runtime.stopCount)
	require.Equal(t, 1, runtime.startCount)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/cpa/run", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var runResp cpa.SyncResult
	decodeSuccessData(t, rec, &runResp)
	require.EqualValues(t, 101, runResp.JobID)
	require.Equal(t, []int64{42}, runtime.runCalls)
}

func TestCPAHandlerListEndpoints(t *testing.T) {
	configStore := &fakeCPAConfigStore{cfg: cpa.DefaultConfig()}
	runtime := &fakeCPARuntime{}
	store := &fakeCPAStore{
		jobs: []*ent.CPASyncJob{
			{ID: 9, Status: cpa.JobStatusSucceeded, TriggerType: cpa.TriggerTypeManual},
		},
		jobsTotal: 2,
		recordsByJob: map[int64][]*ent.CPASyncRecord{
			9: {
				{ID: 1, JobID: 9, FilePath: "/tmp/broken.json", Action: cpa.RecordActionFailed},
			},
		},
		recordsTotalByJob: map[int64]int{9: 1},
		fileStates: []*ent.CPAFileState{
			{ID: 7, FilePath: "/tmp/broken.json", LastResult: cpa.FileResultFailed},
		},
		fileStatesTotal: 1,
	}
	router := setupCPAHandlerRouter(configStore, runtime, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/cpa/jobs?page=1&page_size=1", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var jobsResp struct {
		Items    []*ent.CPASyncJob `json:"items"`
		Total    int               `json:"total"`
		Page     int               `json:"page"`
		PageSize int               `json:"page_size"`
		Pages    int               `json:"pages"`
	}
	decodeSuccessData(t, rec, &jobsResp)
	require.Len(t, jobsResp.Items, 1)
	require.Equal(t, 2, jobsResp.Total)
	require.Equal(t, 2, jobsResp.Pages)
	require.Equal(t, 1, store.lastJobsPage)
	require.Equal(t, 1, store.lastJobsPageSize)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/cpa/jobs/9/records?page=1&page_size=10", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var recordsResp struct {
		Items []*ent.CPASyncRecord `json:"items"`
		Total int                  `json:"total"`
	}
	decodeSuccessData(t, rec, &recordsResp)
	require.Len(t, recordsResp.Items, 1)
	require.Equal(t, "/tmp/broken.json", recordsResp.Items[0].FilePath)
	require.EqualValues(t, 9, store.lastRecordsJobID)
	require.Equal(t, 1, store.lastRecordsPage)
	require.Equal(t, 10, store.lastRecordsPageSize)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/cpa/files?page=1&page_size=20&result=failed&search=broken", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var filesResp struct {
		Items []*ent.CPAFileState `json:"items"`
		Total int                 `json:"total"`
	}
	decodeSuccessData(t, rec, &filesResp)
	require.Len(t, filesResp.Items, 1)
	require.Equal(t, "/tmp/broken.json", filesResp.Items[0].FilePath)
	require.Equal(t, cpa.FileResultFailed, store.lastFileResult)
	require.Equal(t, "broken", store.lastFileSearch)
}

func setupCPAHandlerRouter(configStore *fakeCPAConfigStore, runtime *fakeCPARuntime, store *fakeCPAStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 42, Concurrency: 1})
		c.Set(string(middleware2.ContextKeyUserRole), "admin")
		c.Next()
	})

	handler := NewCPAHandler(configStore, runtime, store)
	group := router.Group("/api/v1/admin/cpa")
	group.GET("/overview", handler.GetOverview)
	group.PUT("/config", handler.UpdateConfig)
	group.POST("/run", handler.RunNow)
	group.GET("/jobs", handler.ListJobs)
	group.GET("/jobs/:job_id/records", handler.ListJobRecords)
	group.GET("/files", handler.ListFileStates)
	return router
}

func decodeSuccessData(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()

	var payload struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.Equal(t, "success", payload.Message)
	require.NoError(t, json.Unmarshal(payload.Data, target))
}

type fakeCPAConfigStore struct {
	cfg   cpa.Config
	saved *cpa.Config
}

func (f *fakeCPAConfigStore) Get(_ context.Context) (*cpa.Config, error) {
	cfg := f.cfg
	return &cfg, nil
}

func (f *fakeCPAConfigStore) Save(_ context.Context, cfg cpa.Config) (*cpa.Config, error) {
	f.cfg = cfg
	f.saved = &cfg
	return &cfg, nil
}

type fakeCPARuntime struct {
	status     cpa.Status
	runResult  cpa.SyncResult
	runCalls   []int64
	startCount int
	stopCount  int
}

func (f *fakeCPARuntime) Start(_ context.Context) error {
	f.startCount++
	return nil
}

func (f *fakeCPARuntime) Stop() {
	f.stopCount++
}

func (f *fakeCPARuntime) RunNow(_ context.Context, triggeredBy int64) (cpa.SyncResult, error) {
	f.runCalls = append(f.runCalls, triggeredBy)
	return f.runResult, nil
}

func (f *fakeCPARuntime) Status() cpa.Status {
	return f.status
}

type fakeCPAStore struct {
	latestJob *ent.CPASyncJob

	jobs             []*ent.CPASyncJob
	jobsTotal        int
	lastJobsPage     int
	lastJobsPageSize int

	recordsByJob      map[int64][]*ent.CPASyncRecord
	recordsTotalByJob map[int64]int
	lastRecordsJobID  int64
	lastRecordsPage   int
	lastRecordsPageSize int

	fileStates        []*ent.CPAFileState
	fileStatesTotal   int
	lastFilesPage     int
	lastFilesPageSize int
	lastFileResult    string
	lastFileSearch    string
}

func (f *fakeCPAStore) LatestJob(context.Context) (*ent.CPASyncJob, error) {
	return f.latestJob, nil
}

func (f *fakeCPAStore) ListJobs(_ context.Context, page, pageSize int) ([]*ent.CPASyncJob, int, error) {
	f.lastJobsPage = page
	f.lastJobsPageSize = pageSize
	return f.jobs, f.jobsTotal, nil
}

func (f *fakeCPAStore) ListRecords(_ context.Context, jobID int64, page, pageSize int) ([]*ent.CPASyncRecord, int, error) {
	f.lastRecordsJobID = jobID
	f.lastRecordsPage = page
	f.lastRecordsPageSize = pageSize
	return f.recordsByJob[jobID], f.recordsTotalByJob[jobID], nil
}

func (f *fakeCPAStore) ListFileStates(_ context.Context, page, pageSize int, result, search string) ([]*ent.CPAFileState, int, error) {
	f.lastFilesPage = page
	f.lastFilesPageSize = pageSize
	f.lastFileResult = result
	f.lastFileSearch = search
	return f.fileStates, f.fileStatesTotal, nil
}
