package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

type submitCmd struct {
	baseCommand
	params core.JobParameters
}

func submitCommand() subcommands.Command {
	return &submitCmd{
		baseCommand: baseCommand{
			name:     "submit",
			synopsis: "Submit a job",
			usage:    "submit [options]",
		},
	}
}

func (cmd *submitCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.params.GitRemote, "remote", "https://github.com/nats-io/nats-server.git", "Git remote URL")
	f.StringVar(&cmd.params.GitRef, "ref", "main", "Git reference (branch, SHA, tag, ...)")
	f.StringVar(&cmd.params.TestsSubDir, "tests_dir", ".", "Name of subdirectory in source where to run tests from")
	f.StringVar(&cmd.params.TestsFilterExpr, "filter", "*", "Filter expression to select what tests are executed")
	f.UintVar(&cmd.params.Reps, "reps", 3, "Number of repetitions for each tests")
	f.DurationVar(&cmd.params.TestMinRuntime, "min_runtime", 1*time.Second, "Minimum duration of each benchmark")
	f.DurationVar(&cmd.params.Timeout, "timeout", 3*time.Hour, "Max time allowed to run all tests")
	f.BoolVar(&cmd.params.SkipCleanup, "skip_cleanup", false, "Do not remove worker temporary directory after execution")
}

func (cmd *submitCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	nc, js, connErr := core.Connect(rootOpts.natsServerUrl, "go-bench-away CLI")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", connErr)
		return subcommands.ExitFailure
	}
	defer nc.Close()

	err := core.SubmitJob(js, cmd.params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v", cmd.name, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
