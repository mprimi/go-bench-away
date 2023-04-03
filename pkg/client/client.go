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
)

// TODO: move to a more appropriate pkg (e.g. messages, enums, etc.)
type JobStatus int

const (
	Submitted JobStatus = iota
	Running
	Failed
	Succeeded
)

type Client interface {
	// Initializes client
	ClientInit(bool) error
	// Submits go-bench-away job
	SubmitJob(string, string, string, string, uint, time.Duration, time.Duration) error
	// Retrieves the status of a job by ID
	GetJobStatusByID(string) (*core.JobStatus, error)
	// Retrieves IDs of all jobs, regardless of JobStatus
	// TODO: Not priority, can remove till we really need it
	//GetJobIDs() ([]string, error)
	// TODO: create close method
	Close() error
}

type GBAClientConfig struct {
	NatsServerUrl string
	Credentials   string
	Namespace     string
}

type gbaClient struct {
	client *client.Client
}

// Rename: New() (*Client, error) => client.New()
//func NewGBAClient(natsServerUrl string, credentials string, namespace string) *GBAClient {
//return &GBAClient{
//NatsServerUrl: natsServerUrl,
//Credentials:   credentials,
//Namespace:     namespace,
//}

//}

func NewGBA(config GBAClientConfig) (*gbaClient, error) {

	gba_client, err := client.NewClient(
		config.NatsServerUrl,
		config.Credentials,
		config.Namespace,
	)
	if err != nil {
		return nil, err
	}

	return &gbaClient{
		client: gba_client,
	}, nil
}

// doesn't return client, internally creates artifacts/resources, New() will use the created resources after Init()
func Init() {

}

// TODO: init within NewGBAClient, remove this later
func ClientInit(c GBAClientConfig, verbose bool) error {
	gbaClient, err := client.NewClient(
		c.NatsServerUrl,
		c.Credentials,
		c.Namespace,
		// TODO: client might've failed to initalize before jobqueue, repo, etc. was created
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.Verbose(verbose),
	)
	if err != nil {
		return err
	}
	defer gbaClient.Close()

	// TODO: split into its own method, don't want to do this with every init
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

	return nil
}

// TODO: returns (JobID string, error)
func (c *gbaClient) SubmitJob(gitRemote string, gitRef string, testsSubDir string, testsFilterExpr string, repetitions uint, testMinRuntime time.Duration, timeout time.Duration) error {

	// TODO: don't need to recreate client
	gbaClient, err := client.NewClient(
		c.NatsServerUrl,
		c.Credentials,
		c.Namespace,
	)
	if err != nil {
		return fmt.Errorf("%v\n", err)

	}
	defer gbaClient.Close()

	currUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("%v\n", err)
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

	gbaClient.SubmitJob(*jobParameters)

	return nil
}

func (c *GBAClient) GetJobStatusByID(jobId string) (*JobStatus, error) {

	// TODO: don't need to recreate client
	gbaClient, err := client.NewClient(
		c.NatsServerUrl,
		c.Credentials,
		c.Namespace,
	)
	if err != nil {
		return nil, err
	}
	defer gbaClient.Close()

	record, err := gbaClient.GetJobRecord(jobId)
	if err != nil {
		return nil, err
	}

	return &record.Status, nil
}

// TODO: add limit parameter, or change to ReturnRecentJobs()
func (c *GBAClient) GetJobIDs() ([]string, error) {
	// TODO
	gbaClient, err := client.NewClient(
		c.NatsServerUrl,
		c.Credentials,
		c.Namespace,
	)
	if err != nil {
		return nil, fmt.Errorf("%v\n", err)
	}

	jobIDs := []string{}

	defer gbaClient.Close()
	return jobIDs, nil
}
