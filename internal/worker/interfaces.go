package worker

import (
	"context"
	"github.com/mprimi/go-bench-away/internal/core"
)

type HandleJobFunc func(*core.JobRecord, uint64) (bool, error)

type DispatcherClient interface {
	DispatchJobs(context.Context, func(*core.JobRecord, uint64) (bool, error)) error
}

type JobUpdaterClient interface {
	UpdateJob(*core.JobRecord, uint64) (uint64, error)
	UploadLogArtifact(string, string) (string, error)
	UploadResultsArtifact(string, string) (string, error)
	UploadScriptArtifact(string, string) (string, error)
}

type WorkerClient interface {
	DispatcherClient
	JobUpdaterClient
}

// type OmniClient interface {
// 	Close()
// 	CreateJobsQueue() error
// 	CreateJobsRepository() error
// 	CreateArtifactsStore() error
// 	DeleteJobsQueue() error
// 	DeleteJobsRepository() error
// 	DeleteArtifactsStore() error
// 	DispatchJobs(context.Context, HandleJobFunc) error
// 	LoadJob(jobId string) (*core.JobRecord, uint64, error)
// 	LoadRecentJobs(int) ([]*core.JobRecord, error)
// 	SubmitJob(params core.JobParameters) (*core.JobRecord, error)
// 	UpdateJob(*core.JobRecord, uint64) (uint64, error)
// 	UploadLogArtifact(string, string) (string, error)
// 	UploadResultsArtifact(string, string) (string, error)
// 	UploadScriptArtifact(string, string) (string, error)
// 	DownloadLogArtifact(*core.JobRecord, string) error
// 	DownloadResultsArtifact(*core.JobRecord, string) error
// 	DownloadScriptArtifact(*core.JobRecord, string) error
// 	LoadResultsArtifact(*core.JobRecord) ([]byte, error)
// 	LoadLogArtifact(*core.JobRecord) ([]byte, error)
// }
