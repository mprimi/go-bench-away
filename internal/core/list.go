package core

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func List(js nats.JetStreamContext, limit int) error {

	// Create jobs KV store
	kv, err := js.KeyValue(jobRecordsStoreName)
	if err != nil {
		return fmt.Errorf("KV store lookup error: %v", err)
	}

	lastSubmitMsg, err := js.GetLastMsg(jobStreamName, fmt.Sprintf(jobSubmitSubjectTemplate, "*"))
	if err != nil {
		return fmt.Errorf("Failed to find last message: %v", err)
	}

	startSeq := lastSubmitMsg.Sequence

	log.Printf("Last submit sequence: %d", startSeq)

	// List job requests from newest to oldest
	for i := startSeq; i > 0; i-- {
		// Stop early if a limit is set
		if limit > 0 && startSeq-i > uint64(limit) {
			break
		}

		rawMsg, err := js.GetMsg(jobStreamName, i)
		if err != nil {
			return fmt.Errorf("Failed retrieve submit request %d: %v", i, err)
		}

		jobId := rawMsg.Header.Get(jobIdHeader)
		if jobId == "" {
			// Missing job id header
			continue
		}

		jobKey := fmt.Sprintf(jobRecordKeyTemplate, jobId)

		kve, err := kv.Get(jobKey)
		if err != nil {
			return fmt.Errorf("Failed to retrieve job record (%s): %v", jobKey, err)
		}

		job := loadJob(kve.Value())
		jp := job.Parameters

		fmt.Printf("[%d] %s %v [%s]\n", i, job.Status, time.Since(job.Created).Truncate(time.Minute), job.Id)
		fmt.Printf("  [%s] %s:%s \n", job.SHA, jp.GitRemote, jp.GitRef)
		fmt.Printf("  [%s] [%s] [%d] [%v]\n", jp.TestsSubDir, jp.TestsFilterExpr, jp.Reps, jp.TestMinRuntime)
		fmt.Printf("\n")
	}

	return nil
}
