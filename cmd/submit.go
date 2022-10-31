package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/client"
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
			usage:    "submit [options]\n",
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
	f.StringVar(&cmd.params.GoPath, "go_path", "", "Run using a custom Go (default looks for `go` in $PATH)")
}

func (cmd *submitCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	client, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer client.Close()

	u, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	cmd.params.Username = u.Username

	job, err := client.SubmitJob(cmd.params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("jobId: %s\n", job.Id)
	return subcommands.ExitSuccess
}
