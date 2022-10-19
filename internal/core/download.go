package core

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"path/filepath"
)

func Download(js nats.JetStreamContext, outputDirPath string, jobIds []string) error {

	// Create jobs KV store
	kv, err := js.KeyValue(jobRecordsStoreName)
	if err != nil {
		return fmt.Errorf("KV store lookup error: %v", err)
	}

	// Bind ObjectStore
	obs, err := js.ObjectStore(artifactsStoreName)
	if err != nil {
		return fmt.Errorf("Failed to bind Object store: %v", err)
	}

	outputDir, err := os.Open(outputDirPath)
	if err != nil {
		return err
	}
	defer outputDir.Close()

	fileInfo, err := outputDir.Stat()
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("Not a directory: %s", outputDirPath)
	}
	outputDir.Close()

	for _, jobId := range jobIds {
		err := downloadJobAndArtifacts(kv, obs, outputDirPath, jobId)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadJobAndArtifacts(kv nats.KeyValue, obs nats.ObjectStore, outputDirPath, jobId string) error {
	jobKey := fmt.Sprintf(jobRecordKeyTemplate, jobId)
	kve, err := kv.Get(jobKey)
	if err != nil {
		return fmt.Errorf("Failed to retrieve job (%s) record: %v", jobId, err)
	}

	job := loadJob(kve.Value())

	jobRecordPath := filepath.Join(outputDirPath, fmt.Sprintf("job-%s-record.json", jobId))
	jobResultsPath := filepath.Join(outputDirPath, fmt.Sprintf("job-%s-results.txt", jobId))
	jobLogPath := filepath.Join(outputDirPath, fmt.Sprintf("job-%s-log.txt", jobId))

	file, err := os.Create(jobRecordPath)
	if err != nil {
		return err
	}
	defer file.Close()

	jobRecordBytes, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	_, err = file.Write(jobRecordBytes)
	if err != nil {
		return err
	}
	file.Close()

	if job.Results == "" {
		log.Printf("Job %s has no results artifact", jobId)
	} else {
		err = obs.GetFile(job.Results, jobResultsPath)
		if err != nil {
			return err
		}
	}

	if job.Log == "" {
		log.Printf("Job %s has no log artifact", jobId)
	} else {
		err = obs.GetFile(job.Log, jobLogPath)
		if err != nil {
			return err
		}
	}

	return nil
}
