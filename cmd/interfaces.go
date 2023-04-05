package cmd

import (
	"github.com/mprimi/go-bench-away/v1/core"
)

type CloseableClient interface {
	Close()
}

type ManagerClient interface {
	CloseableClient
	CreateJobsQueue() error
	CreateJobsRepository() error
	CreateArtifactsStore() error
	DeleteJobsQueue() error
	DeleteJobsRepository() error
	DeleteArtifactsStore() error
}

type JobQueueClient interface {
	CloseableClient
	LoadRecentJobs(int) ([]*core.JobRecord, error)
}

type JobRecordClient interface {
	CloseableClient
	LoadJob(jobId string) (*core.JobRecord, uint64, error)
	DownloadLogArtifact(*core.JobRecord, string) error
	DownloadResultsArtifact(*core.JobRecord, string) error
	DownloadScriptArtifact(*core.JobRecord, string) error
	LoadResultsArtifact(*core.JobRecord) ([]byte, error)
	LoadLogArtifact(*core.JobRecord) ([]byte, error)
}
