package gba

import (
	"fmt"
	"os/user"
	"time"

	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
)

type Client interface {
	ClientInit(verbose bool) error
	SubmitJob(gitRemote string, gitRef string, testsSubDir string, testsFilterExpr string, repetitions uint, testMinRuntime time.Duration, timeout time.Duration)
	GetJobStatusByID(jobId string)
	GetJobIDs() []string
}

type GBAClient struct {
	NatsServerUrl string
	Credentials   string
	Namespace     string
}

func NewGBAClient(natsServerUrl string, credentials string, namespace string) *GBAClient {
	return &GBAClient{
		NatsServerUrl: natsServerUrl,
		Credentials:   credentials,
		Namespace:     namespace,
	}

}

func (c *GBAClient) ClientInit(verbose bool) error {
	gbaClient, err := client.NewClient(
		c.NatsServerUrl,
		c.Credentials,
		c.Namespace,
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.Verbose(verbose),
	)
	if err != nil {
		return fmt.Errorf("%v\n", err)

	}
	defer gbaClient.Close()

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

func (c *GBAClient) SubmitJob(gitRemote string, gitRef string, testsSubDir string, testsFilterExpr string, repetitions uint, testMinRuntime time.Duration, timeout time.Duration) error {

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

func (c *GBAClient) GetJobStatusByID(jobId string) (*core.JobStatus, error) {

	gbaClient, err := client.NewClient(
		c.NatsServerUrl,
		c.Credentials,
		c.Namespace,
	)
	if err != nil {
		return nil, fmt.Errorf("%v\n", err)
	}
	defer gbaClient.Close()

	record, err := gbaClient.GetJobRecord(jobId)
	if err != nil {
		return nil, err
	}

	return &record.Status, nil
}

func (c *GBAClient) GetJobIDs() []string {
	// TODO
	return nil
}
