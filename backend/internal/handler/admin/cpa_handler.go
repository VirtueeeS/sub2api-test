package admin

import (
	"context"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/cpa"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

type cpaConfigStore interface {
	Get(ctx context.Context) (*cpa.Config, error)
	Save(ctx context.Context, cfg cpa.Config) (*cpa.Config, error)
}

type cpaRuntime interface {
	Start(ctx context.Context) error
	Stop()
	RunNow(ctx context.Context, triggeredBy int64) (cpa.SyncResult, error)
	Status() cpa.Status
}

type cpaStore interface {
	LatestJob(ctx context.Context) (*ent.CPASyncJob, error)
	ListJobs(ctx context.Context, page, pageSize int) ([]*ent.CPASyncJob, int, error)
	ListRecords(ctx context.Context, jobID int64, page, pageSize int) ([]*ent.CPASyncRecord, int, error)
	ListFileStates(ctx context.Context, page, pageSize int, result, search string) ([]*ent.CPAFileState, int, error)
}

type CPAHandler struct {
	configStore cpaConfigStore
	runtime     cpaRuntime
	store       cpaStore
}

func NewCPAHandler(configStore cpaConfigStore, runtime cpaRuntime, store cpaStore) *CPAHandler {
	return &CPAHandler{
		configStore: configStore,
		runtime:     runtime,
		store:       store,
	}
}

func (h *CPAHandler) GetOverview(c *gin.Context) {
	cfg, err := h.configStore.Get(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if cfg == nil {
		defaultCfg := cpa.DefaultConfig()
		cfg = &defaultCfg
	}

	latestJob, err := h.store.LatestJob(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	status := cpa.Status{}
	if h.runtime != nil {
		status = h.runtime.Status()
	}

	response.Success(c, gin.H{
		"config":     cfg,
		"status":     status,
		"latest_job": latestJob,
	})
}

func (h *CPAHandler) UpdateConfig(c *gin.Context) {
	var req cpa.Config
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if subject, ok := middleware2.GetAuthSubjectFromContext(c); ok {
		req.UpdatedBy = subject.UserID
	}

	saved, err := h.configStore.Save(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if h.runtime != nil {
		h.runtime.Stop()
		if err := h.runtime.Start(context.Background()); err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}

	response.Success(c, gin.H{"config": saved})
}

func (h *CPAHandler) RunNow(c *gin.Context) {
	triggeredBy := int64(0)
	if subject, ok := middleware2.GetAuthSubjectFromContext(c); ok {
		triggeredBy = subject.UserID
	}

	result, err := h.runtime.RunNow(c.Request.Context(), triggeredBy)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *CPAHandler) ListJobs(c *gin.Context) {
	page, pageSize, ok := parseCPAPagination(c)
	if !ok {
		return
	}

	items, total, err := h.store.ListJobs(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     cpaTotalPages(total, pageSize),
	})
}

func (h *CPAHandler) ListJobRecords(c *gin.Context) {
	jobID, err := strconv.ParseInt(strings.TrimSpace(c.Param("job_id")), 10, 64)
	if err != nil || jobID < 1 {
		response.BadRequest(c, "Invalid job_id")
		return
	}

	page, pageSize, ok := parseCPAPagination(c)
	if !ok {
		return
	}

	items, total, err := h.store.ListRecords(c.Request.Context(), jobID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     cpaTotalPages(total, pageSize),
	})
}

func (h *CPAHandler) ListFileStates(c *gin.Context) {
	page, pageSize, ok := parseCPAPagination(c)
	if !ok {
		return
	}

	items, total, err := h.store.ListFileStates(
		c.Request.Context(),
		page,
		pageSize,
		strings.TrimSpace(c.Query("result")),
		strings.TrimSpace(c.Query("search")),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     cpaTotalPages(total, pageSize),
	})
}

func parseCPAPagination(c *gin.Context) (int, int, bool) {
	page := 1
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			response.BadRequest(c, "Invalid page")
			return 0, 0, false
		}
		page = v
	}

	pageSize := 20
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			response.BadRequest(c, "Invalid page_size")
			return 0, 0, false
		}
		pageSize = v
	}

	return page, pageSize, true
}

func cpaTotalPages(total, pageSize int) int {
	if total == 0 || pageSize < 1 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}
