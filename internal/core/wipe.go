package core

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func WipeSchema(js nats.JetStreamContext) error {

	// Delete stream
	err := js.DeleteStream(jobStreamName)
	if err == nats.ErrStreamNotFound {
		// noop
	} else if err != nil {
		return fmt.Errorf("Stream delete error: %v", err)
	} else {
		log.Printf("Stream deleted")
	}

	// Delete jobs KV store
	err = js.DeleteKeyValue(jobRecordsStoreName)
	if err == nats.ErrStreamNotFound {
		//noop
	} else if err != nil {
		return fmt.Errorf("Bucket delete error: %v", err)
	} else {
		log.Printf("Bucket deleted")
	}

	// Delete jobs artifacts Object store
	err = js.DeleteObjectStore(artifactsStoreName)
	if err == nats.ErrStreamNotFound {
		//noop
	} else if err != nil {
		return fmt.Errorf("Object store delete error: %v", err)
	} else {
		log.Printf("Object store deleted")
	}

	return nil
}
