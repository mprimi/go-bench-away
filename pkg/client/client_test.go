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

	s := server.RunServer(&opts)
	defer s.Shutdown()

	gbaClientConfig := &GBAClientConfig{
		NatsServerUrl: s.ClientURL(),
		Credentials:   "",
		Namespace:     "test",
	}

	return NewGBA

	// client := client.New() || pkg, returns a connected client

	// start nats server
	// init backend storage resources
	// new pkg.Client
	// submit job
	// getJobStatus
	// listJobs (opt)
}
