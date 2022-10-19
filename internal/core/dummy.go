package core

import (
	"github.com/nats-io/nats.go"
)

func Dummy(_ nats.JetStream) error {
	return nil
}
