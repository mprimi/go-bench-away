package worker

import (
	"context"
	"github.com/mprimi/go-bench-away/pkg/core"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockClient struct {
	StubUpdateJob             func(*core.JobRecord, uint64) (uint64, error)
	StubUploadLogArtifact     func(string, string) (string, error)
	StubUploadResultsArtifact func(string, string) (string, error)
	StubUploadScriptArtifact  func(string, string) (string, error)
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

func (c *mockClient) DispatchJobs(ctx context.Context, handleJob func(*core.JobRecord, uint64) (bool, error)) error {
	return nil
}

func newMockClient() WorkerClient {
	return &mockClient{
		StubUpdateJob:             func(*core.JobRecord, uint64) (uint64, error) { return 0, nil },
		StubUploadLogArtifact:     func(string, string) (string, error) { return "", nil },
		StubUploadResultsArtifact: func(string, string) (string, error) { return "", nil },
		StubUploadScriptArtifact:  func(string, string) (string, error) { return "", nil },
	}
}

func TestProcessJob(t *testing.T) {

	var client WorkerClient = newMockClient()
	jobsDir := t.TempDir()

	w, err := NewWorker(client, jobsDir)
	if w == nil {
		t.Fatalf("Client is nil")
	} else if err != nil {
		t.Fatalf("Client init failed: %v", err)
	}

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
