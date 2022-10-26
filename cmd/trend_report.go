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

type trendReportCmd struct {
	baseCommand
	reportCfg reports.TrendConfig
}

func trendReportCommand() subcommands.Command {
	return &trendReportCmd{
		baseCommand: baseCommand{
			name:     "trend-report",
			synopsis: "Creates a report comparing N sets of results (must overlap in benchmarks executed)",
			usage:    "compare-speed <jobId> [jobId [...]]\n",
		},
	}
}

func (cmd *trendReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.reportCfg.OutputPath, "output", "trend.html", "Output report (HTML)")
	f.StringVar(&cmd.reportCfg.Title, "title", "Trend", "Title of the report")
}

func (cmd *trendReportCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Missing job IDs arguments\n")
		return subcommands.ExitUsageError
	}
	cmd.reportCfg.JobIds = f.Args()

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

	reportErr := reports.CreateTrendReport(client, &cmd.reportCfg)
	if reportErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", reportErr)
		return subcommands.ExitFailure
	}

	fmt.Printf("Created report: %s\n", cmd.reportCfg.OutputPath)
	return subcommands.ExitSuccess
}
