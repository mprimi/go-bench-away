package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

type downloadCmd struct {
	baseCommand
	outputDirPath string
}

func downloadCommand() subcommands.Command {
	return &downloadCmd{
		baseCommand: baseCommand{
			name:     "download",
			synopsis: "Download job records and artifacts",
			usage:    "download <jobId> [jobId [...]]",
		},
	}
}

func (cmd *downloadCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.outputDirPath, "output", ".", "Output directory")
}

func (cmd *downloadCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Missing job ID argument(s)\n")
		return subcommands.ExitUsageError
	}

	nc, js, connErr := core.Connect(rootOpts.natsServerUrl, "go-bench-away CLI")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", connErr)
		return subcommands.ExitFailure
	}
	defer nc.Close()

	err := core.Download(js, cmd.outputDirPath, f.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v", cmd.name, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
