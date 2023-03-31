package client

import (
	"fmt"

	"github.com/mprimi/go-bench-away/internal/core"
	"github.com/nats-io/nats.go"
)

const (
	kJobsConsumerNameTmpl      = "%s-worker" // Substitute Namespace
	kJobRecordKeyTmpl          = "jobs/%s"   // substitute Job ID
	kJobIdHeader               = "x-job-id"
	logArtifactKeyTemplate     = "jobs/%s/log.txt"
	resultsArtifactKeyTemplate = "jobs/%s/results.txt"
	scriptArtifactKeyTemplate  = "jobs/%s/run.sh"
)

type Options struct {
	serverUrl          string
	credentials        string
	namespace          string
	clientName         string
	jobsQueueName      string
	jobsSubmitSubject  string
	jobsRepositoryName string
	artifactsStoreName string
	initJobsRepository bool
	initArtifactsStore bool
	initJobsQueue      bool
	verbose            bool
}

type Option func(*Options) error

type Client struct {
	options        Options
	nc             *nats.Conn
	js             nats.JetStreamContext
	jobsRepository nats.KeyValue
	artifactsStore nats.ObjectStore
}

func (c *Client) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}

func (c *Client) GetJobRecord(jobId string) (*core.JobRecord, error) {
	jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, jobId)
	kve, err := c.jobsRepository.Get(jobRecordKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to get job %s record: %v", jobRecordKey, err)
	}
	job, err := core.LoadJob(kve.Value())
	if err != nil {
		return nil, fmt.Errorf("Failed to load job %s: %v", jobId, err)
	}
	return job, nil
}

func NewClient(serverUrl, credentials, namespace string, opts ...Option) (*Client, error) {

	client := &Client{
		options: Options{
			serverUrl:          serverUrl,
			credentials:        credentials,
			namespace:          namespace,
			jobsQueueName:      fmt.Sprintf("%s-jobs", namespace),
			jobsSubmitSubject:  fmt.Sprintf("%s.jobs.submit", namespace),
			jobsRepositoryName: fmt.Sprintf("%s-jobs", namespace),
			artifactsStoreName: fmt.Sprintf("%s-artifacts", namespace),
			clientName:         "go-bench-away CLI", //TODO add user@hostname
		},
	}

	options := &client.options

	if options.namespace == "" {
		return nil, fmt.Errorf("Namespace cannot be empty")
	}

	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	client.logDebug("Creating client with options: %v", options)

	// Set trap for shutdown in case of error
	initCompleted := false
	defer func() {
		if !initCompleted && client.nc != nil {
			client.nc.Close()
		} else {
			client.logDebug("Created client successfully")
		}
	}()

	// Connect
	natsOpts := []nats.Option{}
	if options.credentials != "" {
		natsOpts = append(natsOpts, nats.UserCredentials(options.credentials))
	}
	nc, err := nats.Connect(options.serverUrl, natsOpts...)
	if err != nil {
		return nil, err
	}
	client.nc = nc

	client.logDebug("Connected")

	// Init JetStream
	js, err := client.nc.JetStream()
	if err != nil {
		return nil, err
	}
	client.js = js

	client.logDebug("Created JS context")

	// No way to bind a stream (unlike KV and Obj),
	// but at least check it exists.
	if options.initJobsQueue {
		_, err := client.js.StreamInfo(options.jobsQueueName)
		if err == nats.ErrStreamNotFound {
			return nil, fmt.Errorf("stream not found: %s (need to run init-schema?)", options.jobsQueueName)
		} else if err != nil {
			return nil, err
		}
	}

	client.logDebug("Found job queue")

	if options.initJobsRepository {
		kv, err := client.js.KeyValue(options.jobsRepositoryName)
		if err == nats.ErrBucketNotFound {
			return nil, fmt.Errorf("KV bucket not found: %s (need to run init-schema?)", options.jobsRepositoryName)
		} else if err != nil {
			return nil, err
		}
		client.jobsRepository = kv
	}

	client.logDebug("Bound jobs repository")

	if options.initArtifactsStore {
		obs, err := client.js.ObjectStore(options.artifactsStoreName)
		if err == nats.ErrStreamNotFound {
			return nil, fmt.Errorf("Obj store not found: %s (need to run init-schema?)", options.artifactsStoreName)
		} else if err != nil {
			return nil, err
		}
		client.artifactsStore = obs
	}

	client.logDebug("Bound artifacts store")

	// Disengage shutdown trap
	initCompleted = true
	return client, nil
}

func InitJobsQueue() Option {
	return func(o *Options) error {
		o.initJobsQueue = true
		return nil
	}
}
func InitJobsRepository() Option {
	return func(o *Options) error {
		o.initJobsRepository = true
		return nil
	}
}

func InitArtifactsStore() Option {
	return func(o *Options) error {
		o.initArtifactsStore = true
		return nil
	}
}

func WithClientName(clientName string) Option {
	return func(o *Options) error {
		o.clientName = clientName
		return nil
	}
}

func Verbose(verbose bool) Option {
	return func(o *Options) error {
		o.verbose = verbose
		return nil
	}
}
