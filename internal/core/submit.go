package core

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func SubmitJob(js nats.JetStreamContext, params JobParameters) error {

	kv, err := js.KeyValue(jobRecordsStoreName)
	if err != nil {
		return fmt.Errorf("Failed to bind KV store: %v", err)
	}

	job := newJob(params)

	_, err = kv.Create(jobRecordKey(job.Id), job.bytes())
	if err != nil {
		return fmt.Errorf("Failed to create job record: %v", err)
	}

	submitMsg := nats.NewMsg(jobSubmitSubject(job.Id))
	submitMsg.Header.Add(jobIdHeader, job.Id)
	submitMsg.Header.Add(nats.MsgIdHdr, job.Id) // For deduplication

	pubAck, pubErr := js.PublishMsg(submitMsg)
	if pubErr != nil {
		return fmt.Errorf("Failed to submit job: %v", pubErr)
	}

	log.Printf("Submitted job %s: (%d)", job.Id, pubAck.Sequence)

	return nil
}
