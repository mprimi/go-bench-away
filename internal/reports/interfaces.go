package reports

import (
	"io"

	"github.com/mprimi/go-bench-away/v1/core"
)

type JobRecordClient interface {
	LoadJob(jobId string) (*core.JobRecord, uint64, error)
	LoadResultsArtifact(*core.JobRecord, io.Writer) error
}
