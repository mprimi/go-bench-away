package reports

import (
	_ "embed"
	"fmt"
	"github.com/montanaflynn/stats"
	"golang.org/x/perf/benchstat"
	"html/template"
	"os"

	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
)

//go:embed html/trend.html.tmpl
var trendHtmlTmpl string

type TrendConfig struct {
	Title      string
	JobIds     []string
	OutputPath string
}

func CreateTrendReport(client client.Client, cfg *TrendConfig) error {

	// Create collection of results
	c := &benchstat.Collection{
		Alpha:      kDeltaTestAlpha,
		AddGeoMean: false,
		DeltaTest:  benchstat.UTest,
		Order:      nil, // Preserve file add order
	}

	jobIds := filterDuplicates(cfg.JobIds)

	jobs := make([]*core.JobRecord, len(jobIds))
	for i, jobId := range jobIds {
		job, results, err := loadJobAndResults(client, jobId)
		if err != nil {
			return err
		}
		jobs[i] = job
		c.AddConfig(jobId, results)
	}

	jobLabels := createJobLabels(jobs)

	type SerieVariance struct {
		Type      string    `json:"type"`
		Values    []float64 `json:"array"`
		Visible   bool      `json:"visible"`
		Symmetric bool      `json:"symmetric"`
	}

	type ExperimentSerie struct {
		Name      string        `json:"name"`
		JobIds    []string      `json:"x"`
		Values    []float64     `json:"y"`
		Mode      string        `json:"mode"`
		Type      string        `json:"type"`
		Variances SerieVariance `json:"error_y"`
	}

	type SummaryRow struct {
		BenchmarkName string
		Values        []string
	}

	processTable := func(table *benchstat.Table, expectedUnit string) ([]ExperimentSerie, []SummaryRow, error) {
		timeOpSeries := make([]ExperimentSerie, len(table.Rows))
		timeOpSummaryTable := make([]SummaryRow, len(table.Rows))

		for i, row := range table.Rows {

			averages := make([]float64, len(jobs))
			formattedValues := make([]string, len(jobs))
			variances := make([]float64, len(jobs))

			if len(row.Metrics) != len(jobs) {
				return nil, nil, fmt.Errorf("Unexpected number of values %d for %d jobs", len(row.Metrics), len(jobs))
			}

			for j, metric := range row.Metrics {
				if metric.Unit != expectedUnit {
					return nil, nil, fmt.Errorf("Unexpected unit: %s", metric.Unit)
				}
				averages[j] = metric.Mean

				variances[j] = 0
				if len(metric.RValues) > 1 {
					centile, err := stats.Percentile(metric.RValues, kCentilePercent)
					if err != nil {
						return nil, nil, fmt.Errorf("Failed to calculate %.0f%% centile: %v", kCentilePercent, err)
					}
					variances[j] = centile - metric.Mean
				}

				scaler := benchstat.NewScaler(metric.Mean, metric.Unit)
				formattedValues[j] = fmt.Sprintf("%s ± %s", scaler(metric.Mean), scaler(variances[j]))
			}

			timeOpSeries[i] = ExperimentSerie{
				Name:   row.Benchmark,
				JobIds: jobIds,
				Values: averages,
				Mode:   "lines+markers",
				Type:   "scatter",
				Variances: SerieVariance{
					Type:      "data",
					Values:    variances,
					Visible:   true,
					Symmetric: true,
				},
			}

			timeOpSummaryTable[i] = SummaryRow{
				BenchmarkName: row.Benchmark,
				Values:        formattedValues,
			}
		}

		return timeOpSeries, timeOpSummaryTable, nil
	}

	if len(c.Tables()) == 0 {
		return fmt.Errorf("No tables, the results may not overlap in tests executed")
	}

	type View struct {
		Title     string
		ChartId   template.JS
		Jobs      []*core.JobRecord
		AxisLabel string
		Data      template.JS
		JobLabels []string
		Summary   []SummaryRow
	}

	views := []View{}

	for _, table := range c.Tables() {

		skip := false
		title := ""
		unit := ""
		chartId := ""
		switch table.Metric {
		case kTimeOpTable:
			skip = false
			title = "Time/op trend (lower is better)"
			unit = "ns/op"
			chartId = "time_op"
		case kSpeedTable:
			skip = false
			title = "Throghput trend (higher is better)"
			unit = "MB/s"
			chartId = "speed"
		default:
			skip = true
		}

		if !skip {
			series, summary, err := processTable(table, unit)
			if err != nil {
				return fmt.Errorf("Error processing table %s: %v", table.Metric, err)
			}

			view := View{
				Title:     title,
				ChartId:   template.JS(chartId),
				Jobs:      jobs,
				AxisLabel: unit,
				Data:      mustMarshalIndent(series, 2, 6),
				JobLabels: jobLabels,
				Summary:   summary,
			}

			views = append(views, view)
		}
	}

	// Template values
	tv := struct {
		Title template.HTML
		Jobs  []*core.JobRecord
		Views []View
	}{
		Title: template.HTML(cfg.Title),
		Jobs:  jobs,
		Views: views,
	}

	f, err := os.Create(cfg.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reportTemplate := template.Must(template.New("trend").Parse(trendHtmlTmpl))
	err = reportTemplate.Execute(f, tv)
	if err != nil {
		return err
	}

	return nil
}
