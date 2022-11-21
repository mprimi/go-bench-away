package reports

import (
	"github.com/mprimi/go-bench-away/internal/core"
)

type JobRecordClient interface {
	LoadJob(jobId string) (*core.JobRecord, uint64, error)
	DownloadLogArtifact(*core.JobRecord, string) error
	DownloadResultsArtifact(*core.JobRecord, string) error
	DownloadScriptArtifact(*core.JobRecord, string) error
	LoadResultsArtifact(*core.JobRecord) ([]byte, error)
	LoadLogArtifact(*core.JobRecord) ([]byte, error)
}
