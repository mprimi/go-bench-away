package client

import (
	"context"
	"fmt"
	"time"

	"github.com/mprimi/go-bench-away/internal/core"
	"github.com/nats-io/nats.go"
)

func (c *Client) DispatchJobs(ctx context.Context, handleJob func(*core.JobRecord, uint64) (bool, error)) error {

	// Subscribe with durable pull consumer
	consumerName := fmt.Sprintf(kJobsConsumerNameTmpl, c.options.namespace)
	var subOpts = []nats.SubOpt{
		nats.BindStream(c.options.jobsQueueName),
	}
	sub, err := c.js.PullSubscribe(
		"",
		consumerName,
		subOpts...,
	)
	if err != nil {
		return fmt.Errorf("Subscribe error: %v", err)
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			c.logWarn("Failed to unsubscribe: %v", err)
		}
	}()

	var dispatchErr error

dispatchLoop:
	for {
		if ctx.Err() != nil {
			// Stop dispatching if the context was closed
			dispatchErr = fmt.Errorf("Context closed: %v", ctx.Err())
			break dispatchLoop
		}

		// Try to fetch one message
		msgs, err := sub.Fetch(1, nats.MaxWait(1*time.Second))
		if err == nats.ErrTimeout {
			c.logDebug("No pending jobs")
			continue dispatchLoop
		} else if err != nil {
			c.logWarn("Error fetching next job from queue: %v", err)
			// Wait a second before retrying
			time.Sleep(1 * time.Second)
			continue dispatchLoop
		}

		// Got a message
		if len(msgs) != 1 {
			panic(fmt.Sprintf("Expected 1 message, got: %d", len(msgs)))
		}

		msg := msgs[0]
		if err := msg.InProgress(); err != nil {
			c.logWarn("Failed to mark message as in-progress: %v", err)
		}

		c.logDebug("Handling job queue message")

		jobId := msg.Header.Get(kJobIdHeader)
		if jobId == "" {
			c.logWarn("Ignoring message lacking job ID header")
			if err := msg.Ack(); err != nil {
				c.logWarn("Failed to ACK message: %v", err)
			}
			continue dispatchLoop
		}

		job, revision, err := c.LoadJob(jobId)
		if err != nil {
			// Skip over this message and move over
			c.logWarn("Failed to load job %s record: %v", jobId, err)
			if err := msg.Ack(); err != nil {
				c.logWarn("Failed to ACK message: %v", err)
			}
			continue dispatchLoop
		} else if job.Id != jobId {
			dispatchErr = fmt.Errorf("Job ID mismatch in repository: %s vs %s", job.Id, jobId)
			break dispatchLoop
		}

		c.logDebug("Dispatching job %s", jobId)

		// TODO implement retry
		_, handleErr := handleJob(job, revision)
		if handleErr != nil {
			c.logWarn("Failed to process job %s: %v", jobId, handleErr)
		}

		if err := msg.Ack(); err != nil {
			c.logWarn("Failed to ACK message: %v", err)
		}
	}

	return dispatchErr
}
