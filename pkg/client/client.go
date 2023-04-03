/*
Public wrapper for go-bench-away client internals
*/
package gba

import (
	"fmt"
	"os/user"
	"time"

	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
	"github.com/mprimi/go-bench-away/pkg/enum"
)

// TODO: move to a more appropriate pkg (e.g. messages, enums, etc.)
type JobStatus int

const (
	Submitted JobStatus = iota
	Running
	Failed
	Succeeded
)

type GBAClientInterface interface {
	// Initializes client
	Init() error
	// Closes go-bench-away client connection
	Close()
	// Submits go-bench-away job
	SubmitJob(string, string, string, string, uint, time.Duration, time.Duration) (string, error)
	// Retrieves the status of a job by ID
	GetJobStatusByID(string) (*enum.JobStatus, error)
	// Retrieves IDs of all jobs, regardless of JobStatus
	// TODO: Not priority, can remove till we really need it
	//GetJobIDs() ([]string, error)
}

type GBAClientConfig struct {
	NatsServerUrl string
	Credentials   string
	Namespace     string
}

type GBAClient struct {
	client *client.Client
	config *GBAClientConfig
}

// Rename: New() (*Client, error) [usage]=> client.New()
func New(config GBAClientConfig) (*GBAClient, error) {

	gba_client, err := client.NewClient(
		config.NatsServerUrl,
		config.Credentials,
		config.Namespace,
	)
	if err != nil {
		return nil, err
	}

	return &GBAClient{
		client: gba_client,
		config: &config,
	}, nil
}

// Closes go-bench-away client connection
func (c *GBAClient) Close() {
	c.client.Close()
}

// doesn't return client, internally creates artifacts/resources
func (c *GBAClient) Init() error {
	gbaClient := c.client

	initFuncs := []func() error{
		gbaClient.CreateJobsQueue,
		gbaClient.CreateJobsRepository,
		gbaClient.CreateArtifactsStore,
	}

	for _, fn := range initFuncs {
		if err := fn(); err != nil {
			return fmt.Errorf("%v\n", err)
		}
	}

	gba_client, err := client.NewClient(
		c.config.NatsServerUrl,
		c.config.Credentials,
		c.config.Namespace,
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
	)
	if err != nil {
		return err
	}

	c.client = gba_client

	return nil
}

// TODO: returns (JobID string, error)
func (c *GBAClient) SubmitJob(gitRemote string, gitRef string, testsSubDir string, testsFilterExpr string, repetitions uint, testMinRuntime time.Duration, timeout time.Duration) (string, error) {

	gbaClient := c.client

	currUser, err := user.Current()
	if err != nil {
		return "", err
	}

	jobParameters := &core.JobParameters{
		GitRemote:       gitRemote,
		GitRef:          gitRef,
		TestsSubDir:     testsSubDir,
		TestsFilterExpr: testsFilterExpr,
		Reps:            repetitions,
		TestMinRuntime:  testMinRuntime,
		Timeout:         timeout,
		SkipCleanup:     false,
		Username:        currUser.Username,
	}

	record, err := gbaClient.SubmitJob(*jobParameters)
	if err != nil {
		return "", err
	}
	return record.Id, nil
}

func (c *GBAClient) GetJobStatusByID(jobId string) (*enum.JobStatus, error) {

	gbaClient := c.client

	record, err := gbaClient.GetJobRecord(jobId)
	if err != nil {
		return nil, err
	}

	return (*enum.JobStatus)(&record.Status), nil
}

// TODO: add limit parameter, or change to ReturnRecentJobs()
//func (c *GBAClient) GetJobIDs() ([]string, error) {

//jobIDs := []string{}

//return jobIDs, nil
//}
