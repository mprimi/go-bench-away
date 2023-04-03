package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/internal/reports"
	"github.com/mprimi/go-bench-away/pkg/client"

	"github.com/google/subcommands"
)

type trendReportCmd struct {
	baseCommand
	skipTimeOp          bool
	skipSpeed           bool
	benchmarkFilterExpr string
	hiddenResultsTable  bool
	reportCfg           reports.ReportConfig
}

func trendReportCommand() subcommands.Command {
	return &trendReportCmd{
		baseCommand: baseCommand{
			name:     "trend",
			synopsis: "Creates a report trends of benchmark results over time",
			usage:    "report [options] jobId1 jobId2 ... jobIdN\n",
		},
	}
}

func (cmd *trendReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.reportCfg.OutputPath, "output", "report.html", "Output report (HTML)")
	f.StringVar(&cmd.reportCfg.Title, "title", "", "Title of the report (auto-generated if empty)")
	f.BoolVar(&cmd.skipTimeOp, "no_timeop", false, "Do not include time/op graph and table")
	f.BoolVar(&cmd.skipSpeed, "no_speed", false, "Do not include speed graph and table")
	f.StringVar(&cmd.benchmarkFilterExpr, "benchmark_filter", "", "Regular expression to filter experiments based on benchmark name")
	f.BoolVar(&cmd.hiddenResultsTable, "hide_table", false, "Hide the results table by default")
}

func (cmd *trendReportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	jobIds := f.Args()
	if len(jobIds) < 2 {
		fmt.Fprintf(os.Stderr, "Need at least two job Id arguments\n")
		return subcommands.ExitUsageError
	}

	c, err := client.NewClient(
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
	defer c.Close()

	dataTable, err := reports.CreateDataTable(c, jobIds...)
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
			reports.TrendChart("", reports.TimeOp, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.TimeOp, cmd.benchmarkFilterExpr, cmd.hiddenResultsTable),
		)
	}

	if dataTable.HasSpeed() && !cmd.skipSpeed {
		cmd.reportCfg.AddSections(
			reports.TrendChart("", reports.Speed, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.Speed, cmd.benchmarkFilterExpr, cmd.hiddenResultsTable),
		)
	}

	reportErr := reports.CreateReport(c, &cmd.reportCfg, dataTable)
	if reportErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", reportErr)
		return subcommands.ExitFailure
	}

	fmt.Printf("Created report: %s\n", cmd.reportCfg.OutputPath)
	return subcommands.ExitSuccess
}
