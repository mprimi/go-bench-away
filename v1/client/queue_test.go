package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mprimi/go-bench-away/v1/core"
	server "github.com/nats-io/nats-server/v2/test"
)

func TestSubmit(t *testing.T) {

	// Configure local server and start it
	opts := server.DefaultTestOptions
	opts.Port = -1
	opts.JetStream = true
	opts.StoreDir = t.TempDir()

	s := server.RunServer(&opts)
	defer s.Shutdown()

	namespace := "test"
	credentials := ""
	verbose := true

	// Create client and initialize database schema
	bareClient, err := NewClient(
		s.ClientURL(),
		credentials,
		namespace,
		Verbose(verbose),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer bareClient.Close()

	if err := bareClient.CreateJobsQueue(); err != nil {
		t.Fatal(err)
	}

	if err := bareClient.CreateJobsRepository(); err != nil {
		t.Fatal(err)
	}

	// Create client for test
	client, err := NewClient(
		s.ClientURL(),
		credentials,
		namespace,
		Verbose(verbose),
		InitJobsQueue(),
		InitJobsRepository(),
	)

	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	// Dummy job parameters (test really cares about state)
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
	}

	// Submit 4 jobs
	const numJobs = 4
	jobs := make([]*core.JobRecord, numJobs)
	for i := 0; i < numJobs; i++ {
		jobRecord, err := client.SubmitJob(jobParams)
		if err != nil {
			t.Fatal(err)
		}
		jobs[i] = jobRecord
		t.Logf("Submitted job: %s", jobRecord.Id)
	}

	// Cancel the second job before it's processed
	err = client.CancelJob(jobs[1].Id)
	if err != nil {
		t.Fatal(err)
	}

	// Declare expected statuses after processing queue
	expectedStatuses := []core.JobStatus{
		core.Failed,
		core.Cancelled,
		core.Succeeded,
		core.Submitted,
	}

	// Dispatch jobs in queue, simulating a worker processing them
	ctx, cancel := context.WithCancel(context.Background())
	err = client.DispatchJobs(
		ctx,
		func(record *core.JobRecord, revision uint64) (bool, error) {

			if record.Id == jobs[0].Id {
				// Fail the first job
				record.SetFinalStatus(expectedStatuses[0])
				if _, err := client.UpdateJob(record, revision); err != nil {
					t.Fatal(err)
				}
			} else if record.Id == jobs[1].Id {
				// This job was cancelled and should not be dispatched
				t.Fatal("Cancelled job was dispatched")
			} else if record.Id == jobs[2].Id {
				// Succeed the last job
				record.SetFinalStatus(expectedStatuses[2])
				if _, err := client.UpdateJob(record, revision); err != nil {
					t.Fatal(err)
				}

				// Also stop dispatching
				cancel()
			} else if record.Id == jobs[3].Id {
				// Dispatching should stop before this job is processed
				t.Fatal("Unexpected dispatched job")
			}

			return false, nil
		},
	)

	// Dispatch should return error because context was cancelled
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}

	// Verify status of job records in database
	for i, job := range jobs {
		jobRecord, _, err := client.LoadJob(job.Id)
		if err != nil {
			t.Fatalf("Failed to load job[%d]: %s", i, err)
		}

		if jobRecord.Status != expectedStatuses[i] {
			t.Fatalf("Unexpected status of job[%d]: %s (expected: %s)", i, jobRecord.Status, expectedStatuses[i])
		}
	}

	// Try to cancel jobs that were already processed
	expectCancelSuccessful := []bool{false, false, false, true}
	for i, job := range jobs {
		err := client.CancelJob(job.Id)
		if expectCancelSuccessful[i] && err != nil {
			t.Fatalf("failed to cancel jobs[%d]: %s", i, err)
		} else if !expectCancelSuccessful[i] && err == nil {
			t.Fatalf("expected error cancelling jobs[%d]", i)
		}
	}
}
