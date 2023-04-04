package worker

import (
	"context"

	"github.com/mprimi/go-bench-away/pkg/core"
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
