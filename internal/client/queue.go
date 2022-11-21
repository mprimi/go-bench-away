package client

import (
	"fmt"

	"github.com/mprimi/go-bench-away/internal/core"
	"github.com/nats-io/nats.go"
)

func (c *Client) SubmitJob(params core.JobParameters) (*core.JobRecord, error) {

	// Create a job object from parameters
	job := core.NewJob(params)

	// Create a record in jobs repository
	jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, job.Id)
	_, err := c.jobsRepository.Create(jobRecordKey, job.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Failed to create job record: %v", err)
	}

	// Submit job in the queue
	submitMsg := nats.NewMsg(c.options.jobsSubmitSubject)
	// Message is empty, header points to job record in repository
	submitMsg.Header.Add(kJobIdHeader, job.Id)
	// For deduplication
	submitMsg.Header.Add(nats.MsgIdHdr, job.Id)

	_, pubErr := c.js.PublishMsg(submitMsg)
	if pubErr != nil {
		return nil, fmt.Errorf("Failed to submit job: %v", pubErr)
	}

	return job, nil
}

func (c *Client) LoadRecentJobs(limit int) ([]*core.JobRecord, error) {
	jobs := []*core.JobRecord{}

	lastSubmitMsg, err := c.js.GetLastMsg(c.options.jobsQueueName, c.options.jobsSubmitSubject)
	if err == nats.ErrMsgNotFound {
		return []*core.JobRecord{}, nil
	} else if err != nil {
		return nil, err
	}

	startSeq := lastSubmitMsg.Sequence

	// List job requests from newest to oldest
	for i := startSeq; i > 0; i-- {
		// Stop early if a limit is set
		if limit > 0 && startSeq-i > uint64(limit-1) {
			break
		}

		rawMsg, err := c.js.GetMsg(c.options.jobsQueueName, i)
		if err != nil {
			return nil, fmt.Errorf("Failed retrieve submit request %d: %v", i, err)
		}

		jobId := rawMsg.Header.Get(kJobIdHeader)
		if jobId == "" {
			// Missing job id header
			continue
		}

		jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, jobId)

		kve, err := c.jobsRepository.Get(jobRecordKey)
		if err != nil {
			return nil, fmt.Errorf("Failed to job %s record: %v", jobId, err)
		}

		job, err := core.LoadJob(kve.Value())
		if err != nil {
			return nil, fmt.Errorf("Failed to load job %s: %v", jobId, err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
