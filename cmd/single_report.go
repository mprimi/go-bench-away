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

type singleReportCmd struct {
	baseCommand
	skipTimeOp          bool
	skipSpeed           bool
	benchmarkFilterExpr string
	hiddenResultsTable  bool
	reportCfg           reports.ReportConfig
}

func singleReportCommand() subcommands.Command {
	return &singleReportCmd{
		baseCommand: baseCommand{
			name:     "single-report",
			synopsis: "Creates a report for a single set of results",
			usage:    "report [options] jobId\n",
		},
	}
}

func (cmd *singleReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.reportCfg.OutputPath, "output", "report.html", "Output report (HTML)")
	f.StringVar(&cmd.reportCfg.Title, "title", "", "Title of the report (auto-generated if empty)")
	f.BoolVar(&cmd.skipTimeOp, "no_timeop", false, "Do not include time/op graph and table")
	f.BoolVar(&cmd.skipSpeed, "no_speed", false, "Do not include speed graph and table")
	f.StringVar(&cmd.benchmarkFilterExpr, "benchmark_filter", "", "Regular expression to filter experiments based on benchmark name")
	f.BoolVar(&cmd.hiddenResultsTable, "hide_table", true, "Hide the results table by default")
}

func (cmd *singleReportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Pass one job Id argument\n")
		return subcommands.ExitUsageError
	}

	jobId := f.Args()[0]

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

	dataTable, err := reports.CreateDataTable(c, jobId)
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
			reports.HorizontalBoxChart("", reports.TimeOp, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.TimeOp, cmd.benchmarkFilterExpr, cmd.hiddenResultsTable),
		)
	}

	if dataTable.HasSpeed() && !cmd.skipSpeed {
		cmd.reportCfg.AddSections(
			reports.HorizontalBoxChart("", reports.Speed, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.Speed, cmd.benchmarkFilterExpr, cmd.hiddenResultsTable),
		)
	}

	if cmd.reportCfg.Title == "" {
		cmd.reportCfg.Title = fmt.Sprintf("Job report: %s", jobId)
	}

	reportErr := reports.CreateReport(c, &cmd.reportCfg, dataTable)
	if reportErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", reportErr)
		return subcommands.ExitFailure
	}

	fmt.Printf("Created report: %s\n", cmd.reportCfg.OutputPath)
	return subcommands.ExitSuccess
}
