package client

import (
	"fmt"
	"io"
	"os"

	"github.com/mprimi/go-bench-away/v1/core"

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

func (c *Client) readArtifact(key string, w io.Writer) error {
	if key == "" {
		return fmt.Errorf("missing artifact")
	}
	o, err := c.artifactsStore.Get(key)
	if err != nil {
		return fmt.Errorf("artifact get: %w", err)
	}
	defer o.Close()
	_, err = io.Copy(w, o)
	if err != nil {
		return fmt.Errorf("artifact copy: %w", err)
	}
	return nil
}

func (c *Client) LoadResultsArtifact(job *core.JobRecord, writer io.Writer) error {
	return c.readArtifact(job.Results, writer)
}

func (c *Client) LoadLogArtifact(job *core.JobRecord, writer io.Writer) error {
	return c.readArtifact(job.Log, writer)
}

func (c *Client) LoadScriptArtifact(job *core.JobRecord, writer io.Writer) error {
	return c.readArtifact(job.Script, writer)
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
