package client

import (
	"fmt"

	"github.com/mprimi/go-bench-away/v1/core"
)

func (c *Client) UpdateJob(job *core.JobRecord, revision uint64) (uint64, error) {
	jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, job.Id)
	return c.jobsRepository.Update(jobRecordKey, job.Bytes(), revision)
}
