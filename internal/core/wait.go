package core

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"io"
	"os"

	"github.com/nats-io/nats.go"
)

func WaitJob(js nats.JetStreamContext, jobId string) error {

	kv, err := js.KeyValue(jobRecordsStoreName)
	if err != nil {
		return fmt.Errorf("Failed to bind KV store: %v", err)
	}

	// Bind ObjectStore
	obs, err := js.ObjectStore(artifactsStoreName)
	if err != nil {
		return fmt.Errorf("Failed to bind Object store: %v", err)
	}

	jobKey := fmt.Sprintf(jobRecordKeyTemplate, jobId)
	lastRevision := uint64(0)

poll:
	for {
		kve, err := kv.Get(jobKey)
		if err != nil {
			return fmt.Errorf("Failed to retrieve job record (%s): %v", jobKey, err)
		}

		if kve.Revision() == lastRevision {
			time.Sleep(3 * time.Second)
			continue poll
		}

		lastRevision = kve.Revision()

		job := loadJob(kve.Value())
		log.Printf("Status: %s", job.Status)

		if job.Status == Failed || job.Status == Succeeded {
			log.Printf("Completed in: %v", job.Completed.Sub(job.Created))
			jobBytes, err := json.MarshalIndent(job, "", "  ")
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s\n", jobBytes)

			printArtifactIfPresent(obs, logFileName, job.Log)
			printArtifactIfPresent(obs, resultsFileName, job.Results)

			return nil
		}
	}
}

func printArtifactIfPresent(obs nats.ObjectStore, artifactName, artifactKey string) {
	if artifactKey == "" {
		log.Printf("Artifact %s not present", artifactName)
		return
	}

	obj, err := obs.Get(artifactKey)
	if err != nil {
		log.Printf("Failed to retrieve artifact %s", artifactName)
		return
	}
	fmt.Printf("-------------------------------------------------------------\n")
	fmt.Printf("Artifact: %s\n", artifactName)
	fmt.Printf("-------------------------------------------------------------\n")
	if _, err := io.Copy(os.Stdout, obj); err != nil {
		fmt.Printf("Failed to display artifact: %v\n", err)
	}
	fmt.Printf("-------------------------------------------------------------\n\n")

}
