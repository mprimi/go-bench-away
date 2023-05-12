package reports

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ReportSpec struct {
	Title    string              `json:"title"`
	Sections []ReportSectionSpec `json:"sections"`
}

type ReportSectionSpec struct {
	Title               string `json:"title"`
	Metric              string `json:"metric"`
	Type                string `json:"type"`
	BenchmarkFilterExpr string `json:"filter"`
}

func (spec *ReportSpec) LoadFile(specPath string) error {
	f, err := os.Open(specPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return spec.Load(f)
}

func (spec *ReportSpec) Load(r io.Reader) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(spec)
	if err != nil {
		return err
	}
	return nil
}

func (spec *ReportSpec) ConfigureReport(reportCfg *ReportConfig) error {
	// Set title if present in spec
	if spec.Title != "" {
		reportCfg.Title = spec.Title
	}

	// Always add jobs table
	reportCfg.AddSections(
		JobsTable(),
	)

	for _, sectionSpec := range spec.Sections {

		// Parse metric
		var metric Metric
		switch sectionSpec.Metric {
		case string(TimeOp):
			metric = TimeOp
		case string(Speed):
			metric = Speed
		case string(Throughput):
			metric = Throughput
		case string(OpsPerSec):
			metric = OpsPerSec
		case string(MsgPerSec):
			metric = MsgPerSec
		default:
			// TODO: handle custom metrics
			return fmt.Errorf("unknown metric: %s", sectionSpec.Metric)
		}

		// Parse section (plot type)
		var section SectionConfig
		var isDelta bool
		switch sectionSpec.Type {
		case "trend_chart":
			section = TrendChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr)

		case "horizontal_bar_chart":
			section = HorizontalBarChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr)

		case "horizontal_box_chart":
			section = HorizontalBoxChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr)

		case "horizontal_delta_chart":
			section = HorizontalDeltaChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr)
			isDelta = true

		default:
			return fmt.Errorf("unknown section type: %s", sectionSpec.Type)
		}

		// Add plot to report
		reportCfg.AddSections(section)

		// Add table to report (always hidden)
		// TODO allow configuring hidden or not
		const hideResultsTable = true
		if isDelta {
			reportCfg.AddSections(ResultsDeltaTable(metric, sectionSpec.BenchmarkFilterExpr, hideResultsTable))
		} else {
			reportCfg.AddSections(ResultsTable(metric, sectionSpec.BenchmarkFilterExpr, hideResultsTable))
		}
	}
	return nil
}
