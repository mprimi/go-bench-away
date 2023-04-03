package client

import (
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/pkg/core"

	"github.com/nats-io/nats.go"
)

func (c *Client) LoadJob(jobId string) (*core.JobRecord, uint64, error) {

	c.logDebug("Loading job '%s'", jobId)

	jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, jobId)

	kve, err := c.jobsRepository.Get(jobRecordKey)
	if err == nats.ErrKeyNotFound {
		return nil, 0, fmt.Errorf("Job not found: '%s'", jobId)
	} else if err != nil {
		return nil, 0, err
	}

	job, err := core.LoadJob(kve.Value())
	if err != nil {
		return nil, 0, err
	}

	revision := kve.Revision()

	c.logDebug("Loaded job %s revision %d", jobId, revision)

	return job, revision, nil
}

func (c *Client) DownloadLogArtifact(job *core.JobRecord, filePath string) error {
	if job.Log == "" {
		return fmt.Errorf("Job %s has no log artifact", job.Id)
	}
	return c.artifactsStore.GetFile(job.Log, filePath)
}

func (c *Client) DownloadResultsArtifact(job *core.JobRecord, filePath string) error {
	if job.Results == "" {
		return fmt.Errorf("Job %s has no results artifact", job.Id)
	}
	return c.artifactsStore.GetFile(job.Results, filePath)
}

func (c *Client) DownloadScriptArtifact(job *core.JobRecord, filePath string) error {
	if job.Script == "" {
		return fmt.Errorf("Job %s has no script artifact", job.Id)
	}
	return c.artifactsStore.GetFile(job.Script, filePath)
}

func (c *Client) LoadResultsArtifact(job *core.JobRecord) ([]byte, error) {
	if job.Results == "" {
		return nil, fmt.Errorf("Job %s has no results artifact", job.Id)
	}
	return c.artifactsStore.GetBytes(job.Results)
}

func (c *Client) LoadLogArtifact(job *core.JobRecord) ([]byte, error) {
	if job.Log == "" {
		return nil, fmt.Errorf("Job %s has no log artifact", job.Id)
	}
	return c.artifactsStore.GetBytes(job.Log)
}

func (c *Client) LoadScriptArtifact(job *core.JobRecord) ([]byte, error) {
	if job.Script == "" {
		return nil, fmt.Errorf("Job %s has no script artifact", job.Id)
	}
	return c.artifactsStore.GetBytes(job.Script)
}

func (c *Client) UploadLogArtifact(jobId, logFilePath string) (string, error) {
	key := fmt.Sprintf(logArtifactKeyTemplate, jobId)
	description := fmt.Sprintf("Job %s log file", jobId)
	err := c.uploadArtifact(key, description, logFilePath)
	return key, err
}

func (c *Client) UploadResultsArtifact(jobId, resultsFilePath string) (string, error) {
	key := fmt.Sprintf(resultsArtifactKeyTemplate, jobId)
	description := fmt.Sprintf("Job %s results file", jobId)
	err := c.uploadArtifact(key, description, resultsFilePath)
	return key, err
}

func (c *Client) UploadScriptArtifact(jobId, scriptFilePath string) (string, error) {
	key := fmt.Sprintf(scriptArtifactKeyTemplate, jobId)
	description := fmt.Sprintf("Job %s run script file", jobId)
	err := c.uploadArtifact(key, description, scriptFilePath)
	return key, err
}

func (c *Client) uploadArtifact(key, description, filePath string) error {
	objMeta := nats.ObjectMeta{
		Name:        key,
		Description: description,
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, putErr := c.artifactsStore.Put(&objMeta, file)
	return putErr
}
