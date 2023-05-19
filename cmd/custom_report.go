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

type customReportCmd struct {
	baseCommand
	outputPath   string
	reportCfg    reports.ReportConfig
	specPath     string
	customLabels string
}

func customReportCommand() subcommands.Command {
	return &customReportCmd{
		baseCommand: baseCommand{
			name:     "custom-report",
			synopsis: "Creates a report based on a provided JSON specification",
			usage:    "custom-report [options] -spec <spec_file> jobId1 jobId2 ... jobIdN\n",
		},
	}
}

func (cmd *customReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.outputPath, "output", "report.html", "Output report (HTML)")
	f.StringVar(&cmd.specPath, "spec", "spec.json", "Report configuration (JSON)")
	f.StringVar(&cmd.customLabels, "labels", "", "Use custom labels (comma separated, no spaces, e.g.: \"a,b,c\")")
}

func (cmd *customReportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Must specify at least one job id\n")
		return subcommands.ExitUsageError
	}

	spec := &reports.ReportSpec{}
	err := spec.LoadFile(cmd.specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load report spec: %s\n", err)
		return subcommands.ExitFailure
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

	dataTable, err := reports.CreateDataTable(c, f.Args()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	if rootOptions.verbose {
		cmd.reportCfg.Verbose()
	}

	err = spec.ConfigureReport(&cmd.reportCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure report: %v\n", err)
		return subcommands.ExitFailure
	}

	if cmd.customLabels != "" {
		cmd.reportCfg.SetCustomLabels(strings.Split(cmd.customLabels, ","))
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
