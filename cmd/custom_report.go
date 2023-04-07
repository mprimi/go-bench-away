package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mprimi/go-bench-away/v1/client"
	"github.com/mprimi/go-bench-away/v1/reports"

	"github.com/google/subcommands"
)

type customReportCmd struct {
	baseCommand
	hiddenResultsTable bool
	outputPath         string
	reportCfg          reports.ReportConfig
}

func customReportCommand() subcommands.Command {
	return &customReportCmd{
		baseCommand: baseCommand{
			name:     "custom-report",
			synopsis: "Creates a report based on a provided JSON specification",
			usage:    "custom-report [options] <spec_file>\n",
		},
	}
}

func (cmd *customReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.outputPath, "output", "report.html", "Output report (HTML)")
	f.BoolVar(&cmd.hiddenResultsTable, "hide_table", false, "Hide the results table by default")
}

func (cmd *customReportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Missing argument: report specification file\n")
		return subcommands.ExitUsageError
	}

	spec, err := parseReportSpec(f.Args()[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
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

	dataTable, err := reports.CreateDataTable(c, spec.JobIds...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	if rootOptions.verbose {
		cmd.reportCfg.Verbose()
	}

	if spec.Title != "" {
		cmd.reportCfg.Title = spec.Title
	}

	cmd.reportCfg.AddSections(
		reports.JobsTable(),
	)

	for _, sectionSpec := range spec.Sections {

		var metric reports.Metric
		switch sectionSpec.Metric {
		case string(reports.TimeOp):
			metric = reports.TimeOp
		case string(reports.Speed):
			metric = reports.Speed
		default:
			fmt.Fprintf(os.Stderr, "Unknown metric: %s\n", sectionSpec.Metric)
			return subcommands.ExitFailure
		}

		var section reports.SectionConfig
		var isDelta bool
		switch sectionSpec.Type {
		case "trend_chart":
			section = reports.TrendChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr)

		case "horizontal_bar_chart":
			section = reports.HorizontalBarChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr)

		case "horizontal_delta_chart":
			section = reports.HorizontalDeltaChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr)
			isDelta = true

		default:
			fmt.Fprintf(os.Stderr, "Unknown section type: %s\n", sectionSpec.Type)
			return subcommands.ExitFailure
		}

		cmd.reportCfg.AddSections(section)

		if sectionSpec.ResultsTable || sectionSpec.HiddenResultsTable {
			if isDelta {
				cmd.reportCfg.AddSections(reports.ResultsDeltaTable(metric, sectionSpec.BenchmarkFilterExpr, sectionSpec.HiddenResultsTable))
			} else {
				cmd.reportCfg.AddSections(reports.ResultsTable(metric, sectionSpec.BenchmarkFilterExpr, sectionSpec.HiddenResultsTable))
			}
		}
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

func parseReportSpec(specPath string) (*reports.ReportSpec, error) {
	file, err := os.Open(specPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	specFileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	spec := &reports.ReportSpec{}
	jsonErr := json.Unmarshal(specFileContent, spec)
	if jsonErr != nil {
		return nil, fmt.Errorf("Failed to parse spec: %v", jsonErr)
	}
	return spec, nil
}
