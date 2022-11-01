package client

import (
	"context"
	"fmt"
	"os"
	"time"

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

type HandleJobFunc func(*core.JobRecord, uint64) (bool, error)

type Client interface {
	Close()
	CreateJobsQueue() error
	CreateJobsRepository() error
	CreateArtifactsStore() error
	DeleteJobsQueue() error
	DeleteJobsRepository() error
	DeleteArtifactsStore() error
	DispatchJobs(context.Context, HandleJobFunc) error
	LoadJob(jobId string) (*core.JobRecord, uint64, error)
	LoadRecentJobs(int) ([]*core.JobRecord, error)
	SubmitJob(params core.JobParameters) (*core.JobRecord, error)
	UpdateJob(*core.JobRecord, uint64) (uint64, error)
	UploadLogArtifact(string, string) (string, error)
	UploadResultsArtifact(string, string) (string, error)
	UploadScriptArtifact(string, string) (string, error)
	DownloadLogArtifact(*core.JobRecord, string) error
	DownloadResultsArtifact(*core.JobRecord, string) error
	DownloadScriptArtifact(*core.JobRecord, string) error
	LoadResultsArtifact(*core.JobRecord) ([]byte, error)
}

type clientImpl struct {
	options        Options
	nc             *nats.Conn
	js             nats.JetStreamContext
	jobsRepository nats.KeyValue
	artifactsStore nats.ObjectStore
}

func (c *clientImpl) logDebug(format string, args ...interface{}) {
	if c.options.verbose {
		fmt.Printf("[debug] client: "+format+"\n", args...)
	}
}

func (c *clientImpl) logWarn(format string, args ...interface{}) {
	fmt.Printf("[warning] client: "+format+"\n", args...)
}

func (c *clientImpl) CreateJobsQueue() error {
	c.logDebug("Creating jobs queue %s", c.options.jobsQueueName)

	cfg := nats.StreamConfig{
		Name:        c.options.jobsQueueName,
		Description: "Jobs queue", //TODO add namespace
		Subjects:    []string{c.options.jobsSubmitSubject},
	}

	_, err := c.js.AddStream(&cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) CreateJobsRepository() error {
	c.logDebug("Creating jobs repository %s", c.options.jobsRepositoryName)

	cfg := nats.KeyValueConfig{
		Bucket:      c.options.jobsRepositoryName,
		Description: "Job records repository",
	}

	_, err := c.js.CreateKeyValue(&cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) CreateArtifactsStore() error {
	c.logDebug("Creating artifacts store %s", c.options.artifactsStoreName)

	cfg := nats.ObjectStoreConfig{
		Bucket:      c.options.artifactsStoreName,
		Description: "Job artifacts store",
	}

	_, err := c.js.CreateObjectStore(&cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) DeleteJobsQueue() error {
	c.logDebug("Deleting jobs queue %s", c.options.jobsQueueName)

	err := c.js.DeleteStream(c.options.jobsQueueName)
	if err == nats.ErrStreamNotFound {
		// noop
	} else if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) DeleteJobsRepository() error {
	c.logDebug("Deleting jobs repository %s", c.options.jobsRepositoryName)

	err := c.js.DeleteKeyValue(c.options.jobsRepositoryName)
	if err == nats.ErrStreamNotFound {
		//noop
	} else if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) DeleteArtifactsStore() error {
	c.logDebug("Deleting artifacts store %s", c.options.artifactsStoreName)

	err := c.js.DeleteObjectStore(c.options.artifactsStoreName)
	if err == nats.ErrStreamNotFound {
		//noop
	} else if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}

func (c *clientImpl) DispatchJobs(ctx context.Context, handleJob HandleJobFunc) error {

	// Subscribe with durable pull consumer
	consumerName := fmt.Sprintf(kJobsConsumerNameTmpl, c.options.namespace)
	var subOpts = []nats.SubOpt{
		nats.BindStream(c.options.jobsQueueName),
	}
	sub, err := c.js.PullSubscribe(
		"",
		consumerName,
		subOpts...,
	)
	if err != nil {
		return fmt.Errorf("Subscribe error: %v", err)
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			c.logWarn("Failed to unsubscribe: %v", err)
		}
	}()

	var dispatchErr error

dispatchLoop:
	for {
		if ctx.Err() != nil {
			// Stop dispatching if the context was closed
			dispatchErr = fmt.Errorf("Context closed: %v", ctx.Err())
			break dispatchLoop
		}

		// Try to fetch one message
		msgs, err := sub.Fetch(1, nats.MaxWait(1*time.Second))
		if err == nats.ErrTimeout {
			c.logDebug("No pending jobs")
			continue dispatchLoop
		} else if err != nil {
			c.logWarn("Error fetching next job from queue: %v", err)
			// Wait a second before retrying
			time.Sleep(1 * time.Second)
			continue dispatchLoop
		}

		// Got a message
		if len(msgs) != 1 {
			panic(fmt.Sprintf("Expected 1 message, got: %d", len(msgs)))
		}

		msg := msgs[0]
		if err := msg.InProgress(); err != nil {
			c.logWarn("Failed to mark message as in-progress: %v", err)
		}

		c.logDebug("Handling job queue message")

		jobId := msg.Header.Get(kJobIdHeader)
		if jobId == "" {
			c.logWarn("Ignoring message lacking job ID header")
			if err := msg.Ack(); err != nil {
				c.logWarn("Failed to ACK message: %v", err)
			}
			continue dispatchLoop
		}

		job, revision, err := c.LoadJob(jobId)
		if err != nil {
			// Skip over this message and move over
			c.logWarn("Failed to load job %s record: %v", jobId, err)
			if err := msg.Ack(); err != nil {
				c.logWarn("Failed to ACK message: %v", err)
			}
			continue dispatchLoop
		} else if job.Id != jobId {
			dispatchErr = fmt.Errorf("Job ID mismatch in repository: %s vs %s", job.Id, jobId)
			break dispatchLoop
		}

		c.logDebug("Dispatching job %s", jobId)

		// TODO implement retry
		_, handleErr := handleJob(job, revision)
		if handleErr != nil {
			c.logWarn("Failed to process job %s: %v", jobId, handleErr)
		}

		if err := msg.Ack(); err != nil {
			c.logWarn("Failed to ACK message: %v", err)
		}
	}

	return dispatchErr
}

func (c *clientImpl) LoadJob(jobId string) (*core.JobRecord, uint64, error) {

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

func (c *clientImpl) LoadRecentJobs(limit int) ([]*core.JobRecord, error) {
	jobs := []*core.JobRecord{}

	lastSubmitMsg, err := c.js.GetLastMsg(c.options.jobsQueueName, c.options.jobsSubmitSubject)
	if err == nats.ErrMsgNotFound {
		return []*core.JobRecord{}, nil
	} else if err != nil {
		return nil, err
	}

	startSeq := lastSubmitMsg.Sequence

	// List job requests from newest to oldest
	for i := startSeq; i > 0; i-- {
		// Stop early if a limit is set
		if limit > 0 && startSeq-i > uint64(limit-1) {
			break
		}

		rawMsg, err := c.js.GetMsg(c.options.jobsQueueName, i)
		if err != nil {
			return nil, fmt.Errorf("Failed retrieve submit request %d: %v", i, err)
		}

		jobId := rawMsg.Header.Get(kJobIdHeader)
		if jobId == "" {
			// Missing job id header
			continue
		}

		jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, jobId)

		kve, err := c.jobsRepository.Get(jobRecordKey)
		if err != nil {
			return nil, fmt.Errorf("Failed to job %s record: %v", jobId, err)
		}

		job, err := core.LoadJob(kve.Value())
		if err != nil {
			return nil, fmt.Errorf("Failed to load job %s: %v", jobId, err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (c *clientImpl) SubmitJob(params core.JobParameters) (*core.JobRecord, error) {

	// Create a job object from parameters
	job := core.NewJob(params)

	// Create a record in jobs repository
	jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, job.Id)
	_, err := c.jobsRepository.Create(jobRecordKey, job.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Failed to create job record: %v", err)
	}

	// Submit job in the queue
	submitMsg := nats.NewMsg(c.options.jobsSubmitSubject)
	// Message is empty, header points to job record in repository
	submitMsg.Header.Add(kJobIdHeader, job.Id)
	// For deduplication
	submitMsg.Header.Add(nats.MsgIdHdr, job.Id)

	_, pubErr := c.js.PublishMsg(submitMsg)
	if pubErr != nil {
		return nil, fmt.Errorf("Failed to submit job: %v", pubErr)
	}

	return job, nil
}

func (c *clientImpl) UpdateJob(job *core.JobRecord, revision uint64) (uint64, error) {
	jobRecordKey := fmt.Sprintf(kJobRecordKeyTmpl, job.Id)
	return c.jobsRepository.Update(jobRecordKey, job.Bytes(), revision)
}

func (c *clientImpl) UploadLogArtifact(jobId, logFilePath string) (string, error) {
	key := fmt.Sprintf(logArtifactKeyTemplate, jobId)
	description := fmt.Sprintf("Job %s log file", jobId)
	err := c.uploadArtifact(key, description, logFilePath)
	return key, err
}

func (c *clientImpl) UploadResultsArtifact(jobId, resultsFilePath string) (string, error) {
	key := fmt.Sprintf(resultsArtifactKeyTemplate, jobId)
	description := fmt.Sprintf("Job %s results file", jobId)
	err := c.uploadArtifact(key, description, resultsFilePath)
	return key, err
}

func (c *clientImpl) UploadScriptArtifact(jobId, scriptFilePath string) (string, error) {
	key := fmt.Sprintf(scriptArtifactKeyTemplate, jobId)
	description := fmt.Sprintf("Job %s run script file", jobId)
	err := c.uploadArtifact(key, description, scriptFilePath)
	return key, err
}

func (c *clientImpl) uploadArtifact(key, description, filePath string) error {
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

func (c *clientImpl) DownloadLogArtifact(job *core.JobRecord, filePath string) error {
	if job.Log == "" {
		return fmt.Errorf("Job %s has no log artifact", job.Id)
	}
	return c.artifactsStore.GetFile(job.Log, filePath)
}

func (c *clientImpl) DownloadResultsArtifact(job *core.JobRecord, filePath string) error {
	if job.Results == "" {
		return fmt.Errorf("Job %s has no results artifact", job.Id)
	}
	return c.artifactsStore.GetFile(job.Results, filePath)
}

func (c *clientImpl) DownloadScriptArtifact(job *core.JobRecord, filePath string) error {
	if job.Script == "" {
		return fmt.Errorf("Job %s has no script artifact", job.Id)
	}
	return c.artifactsStore.GetFile(job.Script, filePath)
}

func (c *clientImpl) LoadResultsArtifact(job *core.JobRecord) ([]byte, error) {
	if job.Results == "" {
		return nil, fmt.Errorf("Job %s has no results artifact", job.Results)
	}
	return c.artifactsStore.GetBytes(job.Results)
}

func NewClient(serverUrl, credentials, namespace string, opts ...Option) (Client, error) {

	client := clientImpl{
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
	return &client, nil
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
