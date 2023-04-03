package enum

import (
	"time"

	"github.com/mprimi/go-bench-away/internal/core"
)

type JobStatus int

const (
	Submitted JobStatus = iota
	Running
	Failed
	Succeeded
)

type JobRecord struct {
	Id         string
	Status     JobStatus
	Parameters core.JobParameters

	Created   time.Time
	Started   time.Time
	Completed time.Time

	SHA       string
	GoVersion string

	// Artifacts from job execution
	Log     string
	Results string
	Script  string

	WorkerInfo core.WorkerInfo
}

type JobParameters struct {
	GitRemote       string
	GitRef          string
	TestsSubDir     string
	TestsFilterExpr string
	Reps            uint
	TestMinRuntime  time.Duration
	Timeout         time.Duration
	SkipCleanup     bool
	Username        string
	GoPath          string
	CleanupCmd      string
}

type WorkerInfo struct {
	Hostname string
	Uname    string
	Version  string
}
