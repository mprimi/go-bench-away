package core

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

//go:embed scripts/benchmark.sh
var benchmarkJobScript string

const fetchTimeout = nats.MaxWait(10 * time.Second)

func RunWorker(js nats.JetStreamContext) error {

	// Bind KV
	kv, err := js.KeyValue(jobRecordsStoreName)
	if err == nats.ErrBucketNotFound {
		return fmt.Errorf("Bucket '%s' not found, run 'init' to set up", jobRecordsStoreName)
	} else if err != nil {
		return fmt.Errorf("KV bind error: %v", err)
	}

	// Bind ObjectStore
	obs, err := js.ObjectStore(artifactsStoreName)
	if err == nats.ErrStreamNotFound {
		return fmt.Errorf("ObjStore '%s' not found, run 'init' to set up", jobRecordsStoreName)
	} else if err != nil {
		return fmt.Errorf("Obj bind error: %v", err)
	}

	// Set up consumer
	var subOpts = []nats.SubOpt{
		nats.BindStream(jobStreamName),
	}
	sub, err := js.PullSubscribe(
		"",
		consumerName,
		subOpts...,
	)
	if err != nil {
		return fmt.Errorf("Subscribe error: %v", err)
	}

	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Failed to unsubscribe: %v", err)
		}
	}()

	log.Printf("Worker ready")

	// Pull, execute, publish results loop
	for {
		msgs, err := sub.Fetch(1, fetchTimeout)
		if err == nats.ErrTimeout {
			// No jobs in queue
			log.Printf("No jobs in queue")
		} else if err != nil {
			log.Printf("Fetch error: %v", err)
		} else {
			for _, msg := range msgs {

				err := msg.InProgress()
				if err != nil {
					log.Printf("Failed to mark message as in-progress: %v", err)
				}

				jobId := msg.Header.Get(jobIdHeader)
				if jobId == "" {
					log.Printf("Ignoring message lacking job id header")
				} else {
					// TODO processing may fail.. e.g. record corrupted.
					// May need to NACK, retry later, skip over, ...?
					processJob(kv, obs, jobId)
				}

				ackErr := msg.AckSync()
				if ackErr != nil {
					return fmt.Errorf("Failed to ACK job message: %v", ackErr)
				}
			}
		}
	}
}

func processJob(kv nats.KeyValue, obs nats.ObjectStore, jobId string) {

	// Update job status to RUNNING
	jobKey := jobRecordKey(jobId)
	kve, err := kv.Get(jobKey)
	if err != nil {
		log.Printf("Failed to retrieve job %s record: %v", jobId, err)
	}
	jobRecordRevision := kve.Revision()
	job := loadJob(kve.Value())
	job.Status = Running
	job.Started = time.Now()

	jobRecordRevision, err = kv.Update(jobKey, job.bytes(), jobRecordRevision)
	if err != nil {
		log.Printf("Failed to set job %s status to %s: %v", jobId, job.Status, err)
		return
	}

	log.Printf("Processing job %s", jobId)

	jobTempDir, runErr := runJob(job)

	// Update job status to final
	job.Completed = time.Now()
	job.Status = Succeeded
	if runErr != nil {
		log.Printf("Job %s failed: %v", jobId, runErr)
		job.Status = Failed
	}

	// Upload artifacts
	uploadErr := uploadArtifacts(obs, job, jobTempDir)
	if uploadErr != nil {
		log.Printf("Job %s artifacts upload failed: %v", jobId, uploadErr)
		job.Status = Failed
	}

	if jobTempDir != "" && !job.Parameters.SkipCleanup {
		// Remove job directory, if it was created
		defer os.RemoveAll(jobTempDir)
	}

	// Update job record (status, artifacts, SHA, ...)
	log.Printf("Completed job %s, updating status to: %s", jobId, job.Status)
	_, err = kv.Update(jobKey, job.bytes(), jobRecordRevision)
	if err != nil {
		log.Printf("Failed to set job %s status to %s: %v", jobId, job.Status, err)
		return
	}
}

func runJob(job *JobRecord) (string, error) {

	jobTempDir, err := os.MkdirTemp("", fmt.Sprintf("basho-job-%s-", job.Id))
	if err != nil {
		return "", fmt.Errorf("Failed to create temp directory: %v", err)
	}
	log.Printf("Created temp directory for job: %s", jobTempDir)

	scriptPath := filepath.Join(jobTempDir, "run.sh")
	logPath := filepath.Join(jobTempDir, logFileName)
	resultsPath := filepath.Join(jobTempDir, resultsFileName)
	shaPath := filepath.Join(jobTempDir, "sha.txt")
	goVersionPath := filepath.Join(jobTempDir, "go_version.txt")

	scriptFile, err := os.Create(scriptPath)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to create script: %v", err)
	}
	n, err := scriptFile.WriteString(benchmarkJobScript)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to write job script: %v", err)
	}
	log.Printf("Created job script: %s (%dB)", scriptPath, n)
	scriptFile.Close()

	err = os.Chmod(scriptPath, 0700)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to make script executable: %v", err)
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to create logfile: %v", err)
	}
	defer logFile.Close()

	// Arguments for benchmark script
	arguments := []string{
		jobTempDir,                             //$1
		resultsPath,                            //$2
		shaPath,                                //$3
		goVersionPath,                          //$4
		job.Parameters.GitRemote,               //$5
		job.Parameters.GitRef,                  //$6
		job.Parameters.TestsSubDir,             //$7
		job.Parameters.TestsFilterExpr,         //$8
		fmt.Sprintf("%d", job.Parameters.Reps), //$9
		fmt.Sprintf("%v", job.Parameters.TestMinRuntime), //$10
		fmt.Sprintf("%v", job.Parameters.Timeout),        //$11
	}

	cmd := exec.CommandContext(context.Background(), scriptPath, arguments...)

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to launch job %s: %w", job.Id, err)
	}

	procState, waitErr := cmd.Process.Wait()
	if waitErr != nil {
		return jobTempDir, fmt.Errorf("Error waiting for termination of job %s: %s", job.Id, waitErr)
	}

	shaBytes, err := os.ReadFile(shaPath)
	if err == nil {
		job.SHA = strings.TrimSpace(string(shaBytes))
		log.Printf("Job checkout SHA: %s", job.SHA)
	} else {
		job.SHA = "?"
		log.Printf("Could not determine SHA")
	}

	goVersionBytes, err := os.ReadFile(goVersionPath)
	if err == nil {
		job.GoVersion = strings.TrimSpace(string(goVersionBytes))
		log.Printf("Job Go version: %s", job.GoVersion)
	} else {
		job.GoVersion = "?"
		log.Printf("Could not determine Go version")
	}

	if procState.ExitCode() != 0 {
		return jobTempDir, fmt.Errorf("Non-zero exit code")
	}

	return jobTempDir, nil
}

func uploadArtifacts(obs nats.ObjectStore, job *JobRecord, jobTempDir string) error {

	errors := []string{}

	logPath := filepath.Join(jobTempDir, logFileName)
	logKey := fmt.Sprintf(logKeyTemplate, job.Id)
	logErr := uploadArtifact(obs, logPath, logKey)

	if logErr != nil {
		errors = append(errors, fmt.Sprintf("%s (%v)", logFileName, logErr))
	} else {
		job.Log = logKey
	}

	resultsPath := filepath.Join(jobTempDir, resultsFileName)
	resultsKey := fmt.Sprintf(resultsKeyTemplate, job.Id)
	resErr := uploadArtifact(obs, resultsPath, resultsKey)

	if resErr != nil {
		errors = append(errors, fmt.Sprintf("%s (%v)", resultsFileName, resErr))
	} else {
		job.Results = resultsKey
	}

	if len(errors) == 0 {
		return nil
	}

	errorMsg := strings.Join(errors, ",")

	return fmt.Errorf("Artifact upload errors: %s", errorMsg)
}

func uploadArtifact(obs nats.ObjectStore, filePath, objectKey string) error {

	objMeta := nats.ObjectMeta{
		Name:        objectKey,
		Description: filePath,
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	objInfo, err := obs.Put(&objMeta, file)
	if err != nil {
		return err
	}

	log.Printf("Uploaded artifact: %s (%dB)", objectKey, objInfo.Size)
	return nil
}
