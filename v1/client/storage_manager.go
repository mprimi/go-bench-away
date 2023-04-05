package client

import (
	"github.com/nats-io/nats.go"
)

func (c *Client) CreateJobsQueue() error {
	c.logDebug("Creating jobs queue %s", c.options.jobsQueueName)

	cfg := nats.StreamConfig{
		Name:        c.options.jobsQueueName,
		Description: "Jobs queue", //TODO add namespace
		Subjects:    []string{c.options.jobsSubmitSubject},
	}

	_, err := c.js.AddStream(&cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateJobsRepository() error {
	c.logDebug("Creating jobs repository %s", c.options.jobsRepositoryName)

	cfg := nats.KeyValueConfig{
		Bucket:      c.options.jobsRepositoryName,
		Description: "Job records repository",
	}

	_, err := c.js.CreateKeyValue(&cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateArtifactsStore() error {
	c.logDebug("Creating artifacts store %s", c.options.artifactsStoreName)

	cfg := nats.ObjectStoreConfig{
		Bucket:      c.options.artifactsStoreName,
		Description: "Job artifacts store",
	}

	_, err := c.js.CreateObjectStore(&cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteJobsQueue() error {
	c.logDebug("Deleting jobs queue %s", c.options.jobsQueueName)

	err := c.js.DeleteStream(c.options.jobsQueueName)
	if err == nats.ErrStreamNotFound {
		// noop
	} else if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteJobsRepository() error {
	c.logDebug("Deleting jobs repository %s", c.options.jobsRepositoryName)

	err := c.js.DeleteKeyValue(c.options.jobsRepositoryName)
	if err == nats.ErrStreamNotFound {
		//noop
	} else if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteArtifactsStore() error {
	c.logDebug("Deleting artifacts store %s", c.options.artifactsStoreName)

	err := c.js.DeleteObjectStore(c.options.artifactsStoreName)
	if err == nats.ErrStreamNotFound {
		//noop
	} else if err != nil {
		return err
	}
	return nil
}
