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
	Labels   []string            `json:"labels"`
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

	// Set custom labels, if present
	if spec.Labels != nil && len(spec.Labels) > 0 {
		reportCfg.SetCustomLabels(spec.Labels)
	}

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
		var sections []SectionConfig
		var isDelta bool
		switch sectionSpec.Type {
		case "trend_chart":
			sections = append(sections, TrendChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr))

		case "horizontal_bar_chart":
			sections = append(sections, HorizontalBarChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr))

		case "horizontal_bar_chart_with_delta":
			sections = append(sections, HorizontalBarChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr))
			sections = append(sections, HorizontalDeltaChart(" ", metric, sectionSpec.BenchmarkFilterExpr))
			isDelta = true

		case "horizontal_box_chart":
			sections = append(sections, HorizontalBoxChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr))

		case "horizontal_delta_chart":
			sections = append(sections, HorizontalDeltaChart(sectionSpec.Title, metric, sectionSpec.BenchmarkFilterExpr))
			isDelta = true

		default:
			return fmt.Errorf("unknown section type: %s", sectionSpec.Type)
		}

		// Add plot to report
		reportCfg.AddSections(sections...)

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
