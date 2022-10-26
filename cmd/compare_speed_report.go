package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/reports"
)

type compareSpeedReportCmd struct {
	baseCommand
	reportCfg reports.CompareSpeedConfig
}

func compareSpeedReportCommand() subcommands.Command {
	return &compareSpeedReportCmd{
		baseCommand: baseCommand{
			name:     "compare-speed",
			synopsis: "Creates a report comparing 2 sets of results (must overlap in benchmarks executed)",
			usage:    "compare-speed <jobId1> <jobId2>\n",
		},
	}
}

func (cmd *compareSpeedReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.reportCfg.OutputPath, "output", "compare-speed.html", "Output report (HTML)")
	f.StringVar(&cmd.reportCfg.OldJobName, "oldName", "", "Name for the first job's set of result (defaults to git ref)")
	f.StringVar(&cmd.reportCfg.NewJobName, "newName", "", "Name for the second job's set of result (defaults to git ref)")
	f.StringVar(&cmd.reportCfg.Title, "title", "Speed comparison", "Title of the report")
}

func (cmd *compareSpeedReportCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "Missing job ID argument(s)\n")
		return subcommands.ExitUsageError
	}
	cmd.reportCfg.OldJobId = f.Args()[0]
	cmd.reportCfg.NewJobId = f.Args()[1]

	if cmd.reportCfg.OldJobId == "" || cmd.reportCfg.NewJobId == "" {
		fmt.Fprintf(os.Stderr, "Empty job ID argument(s)\n")
		return subcommands.ExitUsageError
	}

	client, err := client.NewClient(
		rootOpts.natsServerUrl,
		rootOpts.credentials,
		rootOpts.namespace,
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.Verbose(rootOpts.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer client.Close()

	reportErr := reports.CreateCompareSpeedReport(client, &cmd.reportCfg)
	if reportErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", reportErr)
		return subcommands.ExitFailure
	}

	fmt.Printf("Created report: %s\n", cmd.reportCfg.OutputPath)
	return subcommands.ExitSuccess
}
