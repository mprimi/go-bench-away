package core

import (
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
