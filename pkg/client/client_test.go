package gba

import (
	"testing"

	server "github.com/nats-io/nats-server/v2/test"
)

// give test

type MockClient struct {
}

func TestGBAClientInterface(t *testing.T) {

	opts := server.DefaultTestOptions
	opts.Port = -1
	opts.JetStream = true
	opts.StoreDir = t.TempDir()

	namespace := "test"
	credentials := ""

	s := server.RunServer(&opts)
	defer s.Shutdown()

	steps := []struct {
		description string
		expectError bool
		action      func() (*GBAClient, error)
	}{
		{
			"Standard Client",
			false,
			func() (*GBAClient, error) {
				return New(GBAClientConfig{
					s.ClientURL(),
					credentials,
					namespace,
				})
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
	return

	// start nats server
	// init backend storage resources
	// new pkg.Client
	// submit job
	// getJobStatus
	// listJobs (opt)
}
