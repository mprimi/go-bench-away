package worker

import (
	"context"
	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockClient struct {
	StubClose                   func()
	StubCreateJobsQueue         func() error
	StubCreateJobsRepository    func() error
	StubCreateArtifactsStore    func() error
	StubDeleteJobsQueue         func() error
	StubDeleteJobsRepository    func() error
	StubDeleteArtifactsStore    func() error
	StubDispatchJobs            func(context.Context, client.HandleJobFunc) error
	StubLoadJob                 func(jobId string) (*core.JobRecord, uint64, error)
	StubLoadRecentJobs          func(int) ([]*core.JobRecord, error)
	StubSubmitJob               func(params core.JobParameters) (*core.JobRecord, error)
	StubUpdateJob               func(*core.JobRecord, uint64) (uint64, error)
	StubUploadLogArtifact       func(string, string) (string, error)
	StubUploadResultsArtifact   func(string, string) (string, error)
	StubUploadScriptArtifact    func(string, string) (string, error)
	StubDownloadLogArtifact     func(*core.JobRecord, string) error
	StubDownloadResultsArtifact func(*core.JobRecord, string) error
	StubDownloadScriptArtifact  func(*core.JobRecord, string) error
	StubLoadResultsArtifact     func(*core.JobRecord) ([]byte, error)
	StubLoadLogArtifact         func(*core.JobRecord) ([]byte, error)
}

func (c *mockClient) Close() {
	c.StubClose()
}
func (c *mockClient) CreateJobsQueue() error {
	return c.StubCreateJobsQueue()
}
func (c *mockClient) CreateJobsRepository() error {
	return c.StubCreateJobsRepository()
}
func (c *mockClient) CreateArtifactsStore() error {
	return c.StubCreateArtifactsStore()
}
func (c *mockClient) DeleteJobsQueue() error {
	return c.StubDeleteJobsQueue()
}
func (c *mockClient) DeleteJobsRepository() error {
	return c.StubDeleteJobsRepository()
}
func (c *mockClient) DeleteArtifactsStore() error {
	return c.StubDeleteArtifactsStore()
}
func (c *mockClient) DispatchJobs(ctx context.Context, fun client.HandleJobFunc) error {
	return c.StubDispatchJobs(ctx, fun)
}
func (c *mockClient) LoadJob(jobId string) (*core.JobRecord, uint64, error) {
	return c.StubLoadJob(jobId)
}
func (c *mockClient) LoadRecentJobs(limit int) ([]*core.JobRecord, error) {
	return c.StubLoadRecentJobs(limit)
}
func (c *mockClient) SubmitJob(params core.JobParameters) (*core.JobRecord, error) {
	return c.StubSubmitJob(params)
}
func (c *mockClient) UpdateJob(job *core.JobRecord, rev uint64) (uint64, error) {
	return c.StubUpdateJob(job, rev)
}
func (c *mockClient) UploadLogArtifact(jobId string, path string) (string, error) {
	return c.StubUploadLogArtifact(jobId, path)
}
func (c *mockClient) UploadResultsArtifact(jobId string, path string) (string, error) {
	return c.StubUploadResultsArtifact(jobId, path)
}
func (c *mockClient) UploadScriptArtifact(jobId string, path string) (string, error) {
	return c.StubUploadScriptArtifact(jobId, path)
}
func (c *mockClient) DownloadLogArtifact(job *core.JobRecord, path string) error {
	return c.StubDownloadLogArtifact(job, path)
}
func (c *mockClient) DownloadResultsArtifact(job *core.JobRecord, path string) error {
	return c.StubDownloadResultsArtifact(job, path)
}
func (c *mockClient) DownloadScriptArtifact(job *core.JobRecord, path string) error {
	return c.StubDownloadScriptArtifact(job, path)
}
func (c *mockClient) LoadResultsArtifact(job *core.JobRecord) ([]byte, error) {
	return c.StubLoadResultsArtifact(job)
}

func (c *mockClient) LoadLogArtifact(job *core.JobRecord) ([]byte, error) {
	return c.StubLoadLogArtifact(job)
}

func newMockClient() client.Client {
	return &mockClient{
		StubClose:                   func() {},
		StubCreateJobsQueue:         func() error { return nil },
		StubCreateJobsRepository:    func() error { return nil },
		StubCreateArtifactsStore:    func() error { return nil },
		StubDeleteJobsQueue:         func() error { return nil },
		StubDeleteJobsRepository:    func() error { return nil },
		StubDeleteArtifactsStore:    func() error { return nil },
		StubDispatchJobs:            func(context.Context, client.HandleJobFunc) error { return nil },
		StubLoadJob:                 func(jobId string) (*core.JobRecord, uint64, error) { return nil, 0, nil },
		StubLoadRecentJobs:          func(int) ([]*core.JobRecord, error) { return nil, nil },
		StubSubmitJob:               func(params core.JobParameters) (*core.JobRecord, error) { return nil, nil },
		StubUpdateJob:               func(*core.JobRecord, uint64) (uint64, error) { return 0, nil },
		StubUploadLogArtifact:       func(string, string) (string, error) { return "", nil },
		StubUploadResultsArtifact:   func(string, string) (string, error) { return "", nil },
		StubUploadScriptArtifact:    func(string, string) (string, error) { return "", nil },
		StubDownloadLogArtifact:     func(*core.JobRecord, string) error { return nil },
		StubDownloadResultsArtifact: func(*core.JobRecord, string) error { return nil },
		StubDownloadScriptArtifact:  func(*core.JobRecord, string) error { return nil },
		StubLoadResultsArtifact:     func(*core.JobRecord) ([]byte, error) { return nil, nil },
		StubLoadLogArtifact:         func(*core.JobRecord) ([]byte, error) { return nil, nil },
	}
}

func TestProcessJob(t *testing.T) {

	client := newMockClient()
	jobsDir := t.TempDir()

	w, err := NewWorker(client, jobsDir)
	if w == nil {
		t.Fatalf("Client is nil")
	} else if err != nil {
		t.Fatalf("Client init failed: %v", err)
	}
	defer client.Close()

	wi := w.(*workerImpl)

	cleanupHookFilePath := filepath.Join(t.TempDir(), "canary.txt")

	jobParams := core.JobParameters{
		GitRemote:       "https://github.com/mprimi/go-bench-away.git",
		GitRef:          "main",
		TestsSubDir:     "internal/core",
		TestsFilterExpr: ".*",
		Reps:            3,
		TestMinRuntime:  1 * time.Second,
		Timeout:         5 * time.Minute,
		SkipCleanup:     true,
		Username:        "test",
		CleanupCmd:      "touch " + cleanupHookFilePath,
	}

	job := core.NewJob(jobParams)
	retry, err := wi.processJob(job, 1)

	if retry {
		t.Fatalf("Unexpected retry: %v", retry)
	} else if err != nil {
		t.Fatalf("Job processing error: %v", err)
	}

	f, err := os.Open(cleanupHookFilePath)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	f.Close()
}
