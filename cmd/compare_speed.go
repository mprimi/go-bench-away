package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

type compareSpeedCmd struct {
	baseCommand
	parameters core.CompareSpeedParameters
}

func compareSpeedCommand() subcommands.Command {
	return &compareSpeedCmd{
		baseCommand: baseCommand{
			name:     "compare-speed",
			synopsis: "Creates a report comparing 2 sets of benchmark results",
			usage:    "compare-speed <jobId1> <jobId2>",
		},
	}
}

func (cmd *compareSpeedCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.parameters.OutputPath, "output", "compare-speed.html", "Output report (HTML)")
	f.StringVar(&cmd.parameters.OldJobName, "oldName", "A", "Name for the version tested in the first job")
	f.StringVar(&cmd.parameters.NewJobName, "newName", "B", "Name for the version tested in the second job")
	f.StringVar(&cmd.parameters.Title, "title", "Speed comparison", "Name for the version tested in the second job")
}

func (cmd *compareSpeedCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "Missing job ID argument(s)\n")
		return subcommands.ExitUsageError
	}
	cmd.parameters.OldJobId = f.Args()[0]
	cmd.parameters.NewJobId = f.Args()[1]

	nc, js, connErr := core.Connect(rootOpts.natsServerUrl, "go-bench-away CLI")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", connErr)
		return subcommands.ExitFailure
	}
	defer nc.Close()

	err := core.CompareSpeed(js, cmd.parameters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v", cmd.name, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
