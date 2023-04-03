package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type JobStatus int

const (
	Submitted JobStatus = iota
	Running
	Failed
	Succeeded
)

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

// TODO: export to pkg
type JobRecord struct {
	Id string
	// TODO: change type to enum.JobStatus (potentially)
	Status     JobStatus
	Parameters JobParameters

	Created   time.Time
	Started   time.Time
	Completed time.Time

	SHA       string
	GoVersion string

	// Artifacts from job execution
	Log     string
	Results string
	Script  string

	WorkerInfo WorkerInfo
}

func (jr JobStatus) String() string {
	switch jr {
	case Submitted:
		return "SUBMITTED"
	case Running:
		return "RUNNING"
	case Failed:
		return "FAILED"
	case Succeeded:
		return "SUCCEEDED"
	default:
		panic(fmt.Sprintf("Unexpected job status: %d", jr))
	}
}

func (jr JobStatus) Icon() string {
	switch jr {
	case Submitted:
		return "⚪️"
	case Running:
		return "🟣"
	case Failed:
		return "🔴"
	case Succeeded:
		return "🟢"
	default:
		return "❓"
	}
}

func (jr *JobRecord) RunTime() string {
	switch jr.Status {
	case Failed:
		fallthrough
	case Succeeded:
		return jr.Completed.Sub(jr.Started).Round(time.Second).String()
	case Running:
		return time.Since(jr.Started).Round(time.Second).String()
	default:
		return ""
	}
}

func NewJob(params JobParameters) *JobRecord {
	jobId := uuid.New().String()
	return &JobRecord{
		Id:         jobId,
		Status:     Submitted,
		Parameters: params,
		Created:    time.Now().Round(1 * time.Second).UTC(),
	}
}

func LoadJob(data []byte) (*JobRecord, error) {
	job := JobRecord{}
	err := json.Unmarshal(data, &job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (jr *JobRecord) Bytes() []byte {
	bytes, err := json.Marshal(jr)
	if err != nil {
		panic(fmt.Sprintf("Failed to serialize job: %v", err))
	}
	return bytes
}

func (jr *JobRecord) SetFinalStatus(s JobStatus) {
	jr.Status = s
	jr.Completed = time.Now().Round(1 * time.Second).UTC()
}

func (jr *JobRecord) SetRunningStatus() {
	jr.Status = Running
	jr.Started = time.Now().Round(1 * time.Second).UTC()
}
