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

type basicReportCmd struct {
	baseCommand
	skipTimeOp          bool
	skipSpeed           bool
	benchmarkFilterExpr string
	reportCfg           reports.ReportConfig
}

func basicReportCommand() subcommands.Command {
	return &basicReportCmd{
		baseCommand: baseCommand{
			name:     "report",
			synopsis: "Creates a report for one or more sets of results (i.e. jobs)",
			usage:    "report [options] jobId [jobId [...]]\n",
		},
	}
}

func (cmd *basicReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.reportCfg.OutputPath, "output", "report.html", "Output report (HTML)")
	f.StringVar(&cmd.reportCfg.Title, "title", "", "Title of the report (auto-generated if empty)")
	f.BoolVar(&cmd.skipTimeOp, "no_timeop", false, "Do not include time/op graph and table")
	f.BoolVar(&cmd.skipSpeed, "no_speed", false, "Do not include speed graph and table")
	f.StringVar(&cmd.benchmarkFilterExpr, "benchmark_filter", "", "Regular expression to filter experiments based on benchmark name")
}

func (cmd *basicReportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	jobIds := f.Args()
	if len(jobIds) < 1 {
		fmt.Fprintf(os.Stderr, "Pass at least one job Id argument\n")
		return subcommands.ExitUsageError
	}

	client, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer client.Close()

	dataTable, err := reports.CreateDataTable(client, jobIds...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	if rootOptions.verbose {
		cmd.reportCfg.Verbose()
	}

	cmd.reportCfg.AddSections(
		reports.JobsTable(),
	)

	if !cmd.skipTimeOp {
		cmd.reportCfg.AddSections(
			reports.HorizontalBarChart(reports.TimeOp, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.TimeOp, cmd.benchmarkFilterExpr),
		)
	}

	if dataTable.HasSpeed() && !cmd.skipSpeed {
		cmd.reportCfg.AddSections(
			reports.HorizontalBarChart(reports.Speed, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.Speed, cmd.benchmarkFilterExpr),
		)
	}

	reportErr := reports.CreateReport(client, &cmd.reportCfg, dataTable)
	if reportErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", reportErr)
		return subcommands.ExitFailure
	}

	fmt.Printf("Created report: %s\n", cmd.reportCfg.OutputPath)
	return subcommands.ExitSuccess
}
