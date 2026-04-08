package cpa

import (
	"context"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/cpafilestate"
	"github.com/Wei-Shaw/sub2api/ent/cpasyncjob"
	"github.com/Wei-Shaw/sub2api/ent/cpasyncrecord"
)

type Store struct {
	client *dbent.Client
}

type RecordInput struct {
	JobID        int64
	FilePath     string
	Action       string
	AccountID    *int64
	ErrorMessage string
	ProcessedAt  time.Time
}

func NewStore(client *dbent.Client) *Store {
	return &Store{client: client}
}

func (s *Store) CreateJob(ctx context.Context, triggerType string, triggeredBy int64) (*dbent.CPASyncJob, error) {
	job, err := s.client.CPASyncJob.Create().
		SetTriggerType(triggerType).
		SetStatus(JobStatusPending).
		SetTriggeredBy(triggeredBy).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create CPA sync job: %w", err)
	}
	return job, nil
}

func (s *Store) MarkJobRunning(ctx context.Context, id int64, startedAt time.Time) error {
	_, err := s.client.CPASyncJob.UpdateOneID(id).
		SetStatus(JobStatusRunning).
		SetStartedAt(startedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark CPA sync job running: %w", err)
	}
	return nil
}

func (s *Store) FinishJob(ctx context.Context, id int64, status string, stats SyncStats, errorSummary string, finishedAt time.Time) error {
	upd := s.client.CPASyncJob.UpdateOneID(id).
		SetStatus(status).
		SetFilesScanned(stats.FilesScanned).
		SetCreatedCount(stats.CreatedCount).
		SetUpdatedCount(stats.UpdatedCount).
		SetFailedCount(stats.FailedCount).
		SetFinishedAt(finishedAt)
		
	if errorSummary == "" {
		upd = upd.ClearErrorSummary()
	} else {
		upd = upd.SetErrorSummary(errorSummary)
	}
	if _, err := upd.Save(ctx); err != nil {
		return fmt.Errorf("finish CPA sync job: %w", err)
	}
	return nil
}

func (s *Store) AppendRecord(ctx context.Context, input RecordInput) error {
	create := s.client.CPASyncRecord.Create().
		SetJobID(input.JobID).
		SetFilePath(input.FilePath).
		SetAction(input.Action).
		SetProcessedAt(input.ProcessedAt)
	if input.AccountID != nil {
		create = create.SetAccountID(*input.AccountID)
	}
	if input.ErrorMessage != "" {
		create = create.SetErrorMessage(input.ErrorMessage)
	}
	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("append CPA sync record: %w", err)
	}
	return nil
}

func (s *Store) UpsertFileState(ctx context.Context, filePath string, fp FileFingerprint, result string, linkedAccountID *int64, errMsg string, seenAt, syncedAt *time.Time) error {
	row, err := s.client.CPAFileState.Query().Where(cpafilestate.FilePathEQ(filePath)).Only(ctx)
	if dbent.IsNotFound(err) {
		create := s.client.CPAFileState.Create().
			SetFilePath(filePath).
			SetSha256(fp.SHA256).
			SetSize(fp.Size).
			SetLastResult(result)
		if seenAt != nil { create = create.SetLastSeenAt(*seenAt) }
		if syncedAt != nil { create = create.SetLastSyncedAt(*syncedAt) }
		if linkedAccountID != nil { create = create.SetLinkedAccountID(*linkedAccountID) }
		if errMsg != "" { create = create.SetLastError(errMsg) }
		_, err = create.Save(ctx)
		if err != nil { return fmt.Errorf("create CPA file state: %w", err) }
		return nil
	}
	if err != nil { return fmt.Errorf("query CPA file state: %w", err) }
	upd := s.client.CPAFileState.UpdateOneID(row.ID).
		SetSha256(fp.SHA256).
		SetSize(fp.Size).
		SetLastResult(result)
	if seenAt != nil { upd = upd.SetLastSeenAt(*seenAt) }
	if syncedAt != nil { upd = upd.SetLastSyncedAt(*syncedAt) }
	if linkedAccountID != nil { upd = upd.SetLinkedAccountID(*linkedAccountID) } else { upd = upd.ClearLinkedAccountID() }
	if errMsg != "" { upd = upd.SetLastError(errMsg) } else { upd = upd.ClearLastError() }
	if _, err := upd.Save(ctx); err != nil { return fmt.Errorf("update CPA file state: %w", err) }
	return nil
}

func (s *Store) LatestJob(ctx context.Context) (*dbent.CPASyncJob, error) {
	job, err := s.client.CPASyncJob.Query().Order(dbent.Desc(cpasyncjob.FieldCreatedAt)).First(ctx)
	if dbent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest CPA sync job: %w", err)
	}
	return job, nil
}

func (s *Store) ListJobs(ctx context.Context, page, pageSize int) ([]*dbent.CPASyncJob, int, error) {
	page, pageSize = normalizeCPAPagination(page, pageSize)

	total, err := s.client.CPASyncJob.Query().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count CPA sync jobs: %w", err)
	}

	items, err := s.client.CPASyncJob.Query().
		Order(dbent.Desc(cpasyncjob.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list CPA sync jobs: %w", err)
	}
	return items, total, nil
}

func (s *Store) ListRecords(ctx context.Context, jobID int64, page, pageSize int) ([]*dbent.CPASyncRecord, int, error) {
	page, pageSize = normalizeCPAPagination(page, pageSize)

	total, err := s.client.CPASyncRecord.Query().
		Where(cpasyncrecord.JobIDEQ(jobID)).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count CPA sync records: %w", err)
	}

	items, err := s.client.CPASyncRecord.Query().
		Where(cpasyncrecord.JobIDEQ(jobID)).
		Order(dbent.Desc(cpasyncrecord.FieldProcessedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list CPA sync records: %w", err)
	}
	return items, total, nil
}

func (s *Store) ListFileStates(ctx context.Context, page, pageSize int, result, search string) ([]*dbent.CPAFileState, int, error) {
	page, pageSize = normalizeCPAPagination(page, pageSize)

	countQuery := s.client.CPAFileState.Query()
	listQuery := s.client.CPAFileState.Query()
	if result != "" {
		countQuery = countQuery.Where(cpafilestate.LastResultEQ(result))
		listQuery = listQuery.Where(cpafilestate.LastResultEQ(result))
	}
	if search != "" {
		countQuery = countQuery.Where(cpafilestate.FilePathContainsFold(search))
		listQuery = listQuery.Where(cpafilestate.FilePathContainsFold(search))
	}

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count CPA file states: %w", err)
	}

	items, err := listQuery.
		Order(dbent.Desc(cpafilestate.FieldUpdatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list CPA file states: %w", err)
	}
	return items, total, nil
}

func normalizeCPAPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}
