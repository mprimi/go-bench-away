package core

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestJobSerialization(t *testing.T) {

	j := NewJob(JobParameters{
		GitRemote:       "https://example.com/foo/bar",
		GitRef:          "v1.0.0",
		TestsSubDir:     "tests",
		TestsFilterExpr: ".*",
		Reps:            5,
		TestMinRuntime:  1 * time.Second,
		Timeout:         1 * time.Hour,
		SkipCleanup:     true,
		Username:        "Alice",
		GoPath:          "/usr/local/go1.18",
		CleanupCmd:      "rm /tmp/foo_test",
	})

	checkSerializeAndLoad := func() {
		loadedJob, err := LoadJob(j.Bytes())
		if err != nil {
			t.Errorf("Failed to load job: %v", err)
		} else if !reflect.DeepEqual(j, loadedJob) {
			t.Errorf("Jobs mismatch: \nJ1: %v\nJ2: %v", j, loadedJob)
		}
	}

	if j.Status != Submitted {
		t.Errorf("Unexpected status: %v", j.Status)
	}

	if len(j.Id) != 36 {
		t.Errorf("Unexpected Id length: %d", len(j.Id))
	}

	if j.Created == *new(time.Time) {
		t.Errorf("Creation timestamp not set")
	} else if j.Started != *new(time.Time) {
		t.Errorf("Unexpected Started value: %v", j.Started)
	} else if j.Completed != *new(time.Time) {
		t.Errorf("Unexpected Completed value: %v", j.Completed)
	}

	checkSerializeAndLoad()

	j.SetRunningStatus()
	j.WorkerInfo = WorkerInfo{
		Hostname: "runner.example.com",
		Uname:    "runner",
		Version:  "1.2.3",
	}

	if j.Created == *new(time.Time) {
		t.Errorf("Creation timestamp not set")
	} else if j.Started == *new(time.Time) {
		t.Errorf("Unexpected Started value: %v", j.Started)
	} else if j.Completed != *new(time.Time) {
		t.Errorf("Unexpected Completed value: %v", j.Completed)
	}

	checkSerializeAndLoad()

	j.SetFinalStatus(Succeeded)
	j.SHA = "XXXXXX"
	j.GoVersion = "go1.18"
	j.Log = "jobs/blah/log.txt"
	j.Results = "jobs/blah/results.txt"
	j.Script = "jobs/blah/run.sh"

	checkSerializeAndLoad()
}

func TestJobRecord_States(t *testing.T) {

	j := NewJob(JobParameters{
		GitRemote:       "https://example.com/foo/bar",
		GitRef:          "v1.0.0",
		TestsSubDir:     "tests",
		TestsFilterExpr: ".*",
		Reps:            5,
		TestMinRuntime:  1 * time.Second,
		Timeout:         1 * time.Hour,
		Username:        "Alice",
	})

	if j.RunTime() != "" {
		t.Fatalf("Unexpected job run time")
	}

	if j.IsCompleted() {
		t.Fatalf("Unexpected job completed status")
	}

	j.SetRunningStatus()

	if j.IsCompleted() {
		t.Fatalf("Unexpected job completed status")
	}

	if j.RunTime() == "" {
		t.Fatalf("Missing job run time while running")
	}

	j.SetFinalStatus(Failed)

	if j.RunTime() == "" {
		t.Fatalf("Missing job run time after completion")
	}

	if !j.IsCompleted() {
		t.Fatalf("Unexpected job not completed status")
	}
}

func TestJobStatus_IconAndString(t *testing.T) {
	knownStates := []JobStatus{
		Submitted,
		Running,
		Failed,
		Succeeded,
		Cancelled,
	}

	expectedStrings := []string{
		"⚪️ SUBMITTED",
		"🟣 RUNNING",
		"🔴 FAILED",
		"🟢 SUCCEEDED",
		"❌ CANCELLED",
	}

	for i, state := range knownStates {
		jobStateDescription := fmt.Sprintf("%s %s", state.Icon(), state.String())

		if jobStateDescription != expectedStrings[i] {
			t.Fatalf("Expected: %s, actual: %s", expectedStrings[i], jobStateDescription)
		}
	}
}
