package core

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

func Connect(natsURL, clientName string) (*nats.Conn, nats.JetStreamContext, error) {

	// Connect
	var natsOpts = []nats.Option{
		nats.Name(clientName),
	}
	nc, err := nats.Connect(natsURL, natsOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("Connect error: %v", err)
	}

	// Bind JS
	var jsOpts = []nats.JSOpt{}
	js, err := nc.JetStream(jsOpts...)
	if err != nil {
		nc.Close()
		return nil, nil, fmt.Errorf("JS context init error: %v", err)
	}

	return nc, js, nil
}
