package core

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func InitSchema(js nats.JetStreamContext) error {

	// Create stream
	var si *nats.StreamInfo
	si, err := js.StreamInfo(jobStreamName)
	if err == nats.ErrStreamNotFound {

		sconf := nats.StreamConfig{
			Name:        jobStreamName,
			Description: "Queue of pending jobs",
			Subjects:    []string{fmt.Sprintf(jobSubmitSubjectTemplate, "*")},
		}

		si, err = js.AddStream(&sconf)
		if err != nil {
			return fmt.Errorf("Add stream error: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("Stream lookup error: %v", err)
	} else {
		log.Printf("Stream exists")
	}

	log.Printf("Jobs stream: %v", si.Config.Name)

	// Create jobs KV store
	kv, err := js.KeyValue(jobRecordsStoreName)
	if err == nats.ErrBucketNotFound {

		kvconf := nats.KeyValueConfig{
			Bucket:      jobRecordsStoreName,
			Description: "Store job records",
		}

		kv, err = js.CreateKeyValue(&kvconf)

		if err != nil {
			return fmt.Errorf("Create KV store error: %v", err)
		}

	} else if err != nil {
		return fmt.Errorf("KV store lookup error: %v", err)
	} else {
		log.Printf("KV store exists")
	}

	log.Printf("Jobs KV store created: %s", kv.Bucket())

	// Create jobs artifacts Object store
	osconf := nats.ObjectStoreConfig{
		Bucket:      artifactsStoreName,
		Description: "Store job artifacts",
	}

	_, err = js.ObjectStore(artifactsStoreName)
	if err == nats.ErrStreamNotFound {

		_, err = js.CreateObjectStore(&osconf)

		if err != nil {
			return fmt.Errorf("Create Obj store error: %v", err)
		}

	} else if err != nil {
		return fmt.Errorf("Obj store lookup error: %v", err)
	} else {
		log.Printf("Obj store exists")
	}

	log.Printf("Artifacts Obj store: %v", osconf.Bucket)

	return nil
}
