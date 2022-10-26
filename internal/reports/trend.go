package reports

import (
	_ "embed"
	"fmt"
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
		Alpha:      0.1,
		AddGeoMean: false,
		DeltaTest:  benchstat.UTest,
		Order:      nil, // Preserve file add order
	}

	jobs := make([]*core.JobRecord, len(cfg.JobIds))
	for i, jobId := range cfg.JobIds {
		job, results, err := loadJobAndResults(client, jobId)
		if err != nil {
			return err
		}
		jobs[i] = job
		c.AddConfig(jobId, results)
	}

	jobRefs := make([]string, len(jobs))
	for i, j := range jobs {
		jobRefs[i] = j.Parameters.GitRef
	}

	for _, t := range c.Tables() {
		fmt.Printf("Table: %s\n", t.Metric)
	}

	type SerieVariance struct {
		Type      string    `json:"type"`
		MaxValues []float64 `json:"array"`
		MinValues []float64 `json:"arrayminus"`
		Visible   bool      `json:"visible"`
		Symmetric bool      `json:"symmetric"`
	}

	type ExperimentSerie struct {
		Name      string        `json:"name"`
		RefNames  []string      `json:"x"`
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
			maxes := make([]float64, len(jobs))
			mins := make([]float64, len(jobs))
			if len(row.Metrics) != len(jobs) {
				return nil, nil, fmt.Errorf("Unexpected number of values %d for %d jobs", len(row.Metrics), len(jobs))
			}

			for j, metric := range row.Metrics {
				if metric.Unit != expectedUnit {
					return nil, nil, fmt.Errorf("Unexpected unit: %s", metric.Unit)
				}
				averages[j] = metric.Mean
				maxes[j] = metric.Max - metric.Mean
				mins[j] = metric.Mean - metric.Min

				formattedValues[j] = benchstat.NewScaler(metric.Mean, metric.Unit)(metric.Mean)
			}

			timeOpSeries[i] = ExperimentSerie{
				Name:     row.Benchmark,
				RefNames: jobRefs,
				Values:   averages,
				Mode:     "lines+markers",
				Type:     "scatter",
				Variances: SerieVariance{
					Type:      "data",
					MaxValues: maxes,
					MinValues: mins,
					Visible:   true,
					Symmetric: false,
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
		Series    template.JS
		Summary   []SummaryRow
	}

	views := []View{}

	for _, table := range c.Tables() {

		skip := false
		title := ""
		unit := ""
		chartId := ""
		switch table.Metric {
		case "time/op":
			skip = false
			title = "Time/op trend (lower is better)"
			unit = "ns/op"
			chartId = "time_op"
		case "speed":
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
				Series:    mustMarshalIndent(series, 2, 6),
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
