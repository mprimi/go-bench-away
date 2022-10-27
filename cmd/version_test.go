package cmd

import (
	"context"
	"flag"
	"testing"
)

func TestVersionCommand(t *testing.T) {

	versionCommand().Execute(
		context.Background(),
		flag.NewFlagSet("test", flag.ContinueOnError),
	)

}
