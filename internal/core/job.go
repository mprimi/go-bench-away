package core

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
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
}

type JobRecord struct {
	Id         string
	Status     JobStatus
	Parameters JobParameters

	Created   time.Time
	Started   time.Time
	Completed time.Time

	SHA       string
	GoVersion string
	Log       string
	Results   string
}

func (this JobStatus) String() string {
	switch this {
	case Submitted:
		return "SUBMITTED"
	case Running:
		return "RUNNING"
	case Failed:
		return "FAILED"
	case Succeeded:
		return "SUCCEEDED"
	default:
		panic(fmt.Sprintf("Unexpected job status: %d", this))
	}
}

func (this JobStatus) Icon() string {
	switch this {
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

func NewJob(params JobParameters) *JobRecord {
	jobId := uuid.New().String()
	return &JobRecord{
		Id:         jobId,
		Status:     Submitted,
		Parameters: params,
		Created:    time.Now(),
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

func (this *JobRecord) Bytes() []byte {
	bytes, err := json.Marshal(this)
	if err != nil {
		panic(fmt.Sprintf("Failed to serialize job: %v", err))
	}
	return bytes
}
