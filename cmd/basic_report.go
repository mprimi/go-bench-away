package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mprimi/go-bench-away/v1/client"
	"github.com/mprimi/go-bench-away/v1/reports"

	"github.com/google/subcommands"
)

type basicReportCmd struct {
	baseCommand
	skipTimeOp          bool
	skipSpeed           bool
	benchmarkFilterExpr string
	hiddenResultsTable  bool
	outputPath          string
	reportCfg           reports.ReportConfig
	customLabels        string
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
	f.StringVar(&cmd.outputPath, "output", "report.html", "Output report (HTML)")
	f.StringVar(&cmd.reportCfg.Title, "title", "", "Title of the report (auto-generated if empty)")
	f.BoolVar(&cmd.skipTimeOp, "no_timeop", false, "Do not include time/op graph and table")
	f.BoolVar(&cmd.skipSpeed, "no_speed", false, "Do not include speed graph and table")
	f.StringVar(&cmd.benchmarkFilterExpr, "benchmark_filter", "", "Regular expression to filter experiments based on benchmark name")
	f.BoolVar(&cmd.hiddenResultsTable, "hide_table", false, "Hide the results table by default")
	f.StringVar(&cmd.customLabels, "labels", "", "Use custom labels (comma separated, no spaces, e.g.: \"a,b,c\")")
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

	if cmd.customLabels != "" {
		cmd.reportCfg.SetCustomLabels(strings.Split(cmd.customLabels, ","))
	}

	cmd.reportCfg.AddSections(
		reports.JobsTable(),
	)

	if !cmd.skipTimeOp {
		cmd.reportCfg.AddSections(
			reports.HorizontalBarChart("", reports.TimeOp, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.TimeOp, cmd.benchmarkFilterExpr, cmd.hiddenResultsTable),
		)
	}

	if dataTable.HasSpeed() && !cmd.skipSpeed {
		cmd.reportCfg.AddSections(
			reports.HorizontalBarChart("", reports.Speed, cmd.benchmarkFilterExpr),
			reports.ResultsTable(reports.Speed, cmd.benchmarkFilterExpr, cmd.hiddenResultsTable),
		)
	}

	file, err := os.Create(cmd.outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer file.Close()

	reportErr := reports.WriteReport(&cmd.reportCfg, dataTable, file)
	if reportErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", reportErr)
		return subcommands.ExitFailure
	}

	fmt.Printf("Created report: %s\n", cmd.outputPath)
	return subcommands.ExitSuccess
}
