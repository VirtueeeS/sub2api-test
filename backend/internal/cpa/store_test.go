package cpa

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStoreCreateFinishJobAndLatest(t *testing.T) {
	client := newCPATestClient(t)
	store := NewStore(client)
	job, err := store.CreateJob(context.Background(), TriggerTypeManual, 12)
	require.NoError(t, err)
	require.NoError(t, store.MarkJobRunning(context.Background(), job.ID, time.Now().UTC()))
	require.NoError(t, store.FinishJob(context.Background(), job.ID, JobStatusSucceeded, SyncStats{FilesScanned: 3, CreatedCount: 1, UpdatedCount: 1, FailedCount: 0}, "", time.Now().UTC()))
	latest, err := store.LatestJob(context.Background())
	require.NoError(t, err)
	require.Equal(t, job.ID, latest.ID)
	require.Equal(t, JobStatusSucceeded, latest.Status)
	require.Equal(t, 3, latest.FilesScanned)
}

func TestStoreAppendRecordAndUpsertFileState(t *testing.T) {
	client := newCPATestClient(t)
	store := NewStore(client)
	job, err := store.CreateJob(context.Background(), TriggerTypeAuto, 0)
	require.NoError(t, err)
	accountID := int64(88)
	now := time.Now().UTC()
	require.NoError(t, store.AppendRecord(context.Background(), RecordInput{JobID: job.ID, FilePath: "/tmp/a.json", Action: RecordActionCreated, AccountID: &accountID, ProcessedAt: now}))
	require.NoError(t, store.UpsertFileState(context.Background(), "/tmp/a.json", FileFingerprint{SHA256: "abc", Size: 12}, FileResultCreated, &accountID, "", &now, &now))
	states, err := client.CPAFileState.Query().All(context.Background())
	require.NoError(t, err)
	require.Len(t, states, 1)
	require.Equal(t, "abc", states[0].Sha256)
	require.Equal(t, FileResultCreated, states[0].LastResult)
	records, err := client.CPASyncRecord.Query().All(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, RecordActionCreated, records[0].Action)
}

func TestStoreListQueries(t *testing.T) {
	client := newCPATestClient(t)
	store := NewStore(client)

	jobOne, err := store.CreateJob(context.Background(), TriggerTypeAuto, 1)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	jobTwo, err := store.CreateJob(context.Background(), TriggerTypeManual, 2)
	require.NoError(t, err)

	now := time.Now().UTC()
	accountID := int64(99)
	require.NoError(t, store.AppendRecord(context.Background(), RecordInput{
		JobID:       jobOne.ID,
		FilePath:    "/tmp/alpha.json",
		Action:      RecordActionCreated,
		AccountID:   &accountID,
		ProcessedAt: now,
	}))
	require.NoError(t, store.AppendRecord(context.Background(), RecordInput{
		JobID:       jobTwo.ID,
		FilePath:    "/tmp/broken.json",
		Action:      RecordActionFailed,
		ProcessedAt: now,
	}))

	require.NoError(t, store.UpsertFileState(context.Background(), "/tmp/alpha.json", FileFingerprint{
		SHA256: "alpha",
		Size:   11,
	}, FileResultCreated, &accountID, "", &now, &now))
	require.NoError(t, store.UpsertFileState(context.Background(), "/tmp/broken.json", FileFingerprint{
		SHA256: "broken",
		Size:   22,
	}, FileResultFailed, nil, "boom", &now, nil))

	jobs, total, err := store.ListJobs(context.Background(), 1, 1)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, jobs, 1)
	require.Equal(t, jobTwo.ID, jobs[0].ID)

	records, total, err := store.ListRecords(context.Background(), jobTwo.ID, 1, 10)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, records, 1)
	require.Equal(t, "/tmp/broken.json", records[0].FilePath)
	require.Equal(t, RecordActionFailed, records[0].Action)

	states, total, err := store.ListFileStates(context.Background(), 1, 10, FileResultFailed, "broken")
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, states, 1)
	require.Equal(t, "/tmp/broken.json", states[0].FilePath)
	require.Equal(t, FileResultFailed, states[0].LastResult)
}

func TestStoreLatestJobReturnsNilWhenEmpty(t *testing.T) {
	client := newCPATestClient(t)
	store := NewStore(client)

	job, err := store.LatestJob(context.Background())
	require.NoError(t, err)
	require.Nil(t, job)
}
