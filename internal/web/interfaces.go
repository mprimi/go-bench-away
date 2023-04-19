package web

import (
	"io"

	"github.com/mprimi/go-bench-away/v1/core"
)

type WebClient interface {
	LoadJob(jobId string) (*core.JobRecord, uint64, error)
	GetQueueStatus() (*core.QueueStatus, error)
	LoadRecentJobs(limit int) ([]*core.JobRecord, error)
	LoadResultsArtifact(job *core.JobRecord, w io.Writer) error
	LoadLogArtifact(job *core.JobRecord, w io.Writer) error
	LoadScriptArtifact(job *core.JobRecord, w io.Writer) error
	CancelJob(id string) error
}
