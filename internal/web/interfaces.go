package web

import (
	"github.com/mprimi/go-bench-away/internal/core"
)

type WebClient interface {
	LoadJob(jobId string) (*core.JobRecord, uint64, error)
	GetQueueStatus() (*core.QueueStatus, error)
	LoadRecentJobs(limit int) ([]*core.JobRecord, error)
	LoadResultsArtifact(job *core.JobRecord) ([]byte, error)
	LoadLogArtifact(job *core.JobRecord) ([]byte, error)
	LoadScriptArtifact(job *core.JobRecord) ([]byte, error)
}
