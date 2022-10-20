package client

import (
	"testing"

	server "github.com/nats-io/nats-server/v2/test"
)

func TestNewClient(t *testing.T) {

	opts := server.DefaultTestOptions
	opts.Port = -1
	opts.JetStream = true
	opts.StoreDir = t.TempDir()

	s := server.RunServer(&opts)
	defer s.Shutdown()

	namespace := "test"
	credentials := ""
	verbose := true

	steps := []struct {
		description string
		expectError bool
		action      func() (Client, error)
	}{
		{
			"Client with no bindings",
			false,
			func() (Client, error) {
				return NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
				)
			},
		},
		{
			"Delete jobs queue, jobs repository, artifacts store before they exist",
			false,
			func() (Client, error) {
				client, err := NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
				)
				if err != nil {
					return client, err
				}
				if err := client.DeleteJobsQueue(); err != nil {
					return client, err
				}
				if err := client.DeleteJobsRepository(); err != nil {
					return client, err
				}
				if err := client.DeleteArtifactsStore(); err != nil {
					return client, err
				}
				return client, nil
			},
		},
		{
			"Client with JobsQueue before it's initialized",
			true,
			func() (Client, error) {
				return NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
					InitJobsQueue(),
				)
			},
		},
		{
			"Client with JobsRepository before it's initialized",
			true,
			func() (Client, error) {
				return NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
					InitJobsRepository(),
				)
			},
		},
		{
			"Client with ArtifactsStore before it's initialized",
			true,
			func() (Client, error) {
				return NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
					InitArtifactsStore(),
				)
			},
		},
		{
			"Initialize jobs queue, jobs repository, artifacts store",
			false,
			func() (Client, error) {
				client, err := NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
				)
				if err != nil {
					return client, err
				}
				if err := client.CreateJobsQueue(); err != nil {
					return client, err
				}
				if err := client.CreateJobsRepository(); err != nil {
					return client, err
				}
				if err := client.CreateArtifactsStore(); err != nil {
					return client, err
				}
				return client, nil
			},
		},
		{
			"Client JobsQueue, JobsRepository, ArtifactsStore",
			false,
			func() (Client, error) {
				return NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
					InitJobsQueue(),
					InitJobsRepository(),
					InitArtifactsStore(),
				)
			},
		},
		{
			"Initialize jobs queue, jobs repository, artifacts store when they already exist",
			false,
			func() (Client, error) {
				client, err := NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
				)
				if err != nil {
					return client, err
				}
				if err := client.CreateJobsQueue(); err != nil {
					return client, err
				}
				if err := client.CreateJobsRepository(); err != nil {
					return client, err
				}
				if err := client.CreateArtifactsStore(); err != nil {
					return client, err
				}
				return client, nil
			},
		},
		{
			"Delete jobs queue, jobs repository, artifacts store",
			false,
			func() (Client, error) {
				client, err := NewClient(
					s.ClientURL(),
					credentials,
					namespace,
					Verbose(verbose),
				)
				if err != nil {
					return client, err
				}
				if err := client.DeleteJobsQueue(); err != nil {
					return client, err
				}
				if err := client.DeleteJobsRepository(); err != nil {
					return client, err
				}
				if err := client.DeleteArtifactsStore(); err != nil {
					return client, err
				}
				return client, nil
			},
		},
	}

	for i, step := range steps {
		stepClient, stepErr := step.action()
		if stepClient != nil {
			stepClient.Close()
		}

		if (stepErr != nil) != step.expectError {
			t.Fatalf(
				"Error in step %d/%d: '%s' -- expect error: %v, got error: %v",
				i+1,
				len(steps),
				step.description,
				step.expectError,
				stepErr,
			)
		}
	}
}
