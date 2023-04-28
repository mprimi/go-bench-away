package reports

import (
	"fmt"

	"github.com/mprimi/go-bench-away/v1/core"
	"golang.org/x/perf/benchstat"
)

type trendChartSeries struct {
	BenchmarkName string
	JobIds        []string
	Values        []float64
	Deviation     []float64
	HoverLabels   []string
}

type trendChartSection struct {
	baseSection
	Metric        Metric
	ChartId       string
	NumBenchmarks int
	JobLabels     []string
	JobIds        []string
	Series        []trendChartSeries
}

func (s *trendChartSection) fillData(dt *dataTableImpl) error {
	var table *benchstat.Table
	switch s.Metric {
	case TimeOp:
		table = dt.timeOpTable
		s.YTitle = "time/op"
		s.XTitle = "(lower is better)"
	case Speed:
		table = dt.speedTable
		s.YTitle = "throughput"
		s.XTitle = "(higher is better)"
	case OpsPerSec:
		table = invertTimeOpTable(dt.timeOpTable, s.Metric)
		s.YTitle = "operations/s"
		s.XTitle = "(higher is better)"
	case MsgPerSec:
		table = invertTimeOpTable(dt.timeOpTable, s.Metric)
		s.YTitle = "messages/s"
		s.XTitle = "(higher is better)"
	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	rows := filterByBenchmarkName(table.Rows, s.BenchmarkFilter)

	s.NumBenchmarks = len(rows)
	s.JobLabels = dt.jobLabels
	s.Series = make([]trendChartSeries, s.NumBenchmarks)

	s.JobIds = dt.mapJobs(func(job *core.JobRecord) string { return job.Id })

	for i, row := range rows {
		sr := &s.Series[i]
		sr.BenchmarkName = row.Benchmark
		sr.JobIds = s.JobIds

		sr.Values = make([]float64, len(s.JobIds))
		sr.Deviation = make([]float64, len(s.JobIds))
		sr.HoverLabels = make([]string, len(s.JobIds))

		for j, m := range row.Metrics {
			sr.Values[j], sr.Deviation[j], sr.HoverLabels[j] = valueDeviationAndScaledString(m)
		}
	}

	return nil
}

func TrendChart(title string, metric Metric, filterExpr string) SectionConfig {
	if title == "" {
		title = fmt.Sprintf("%s trend", metric)
	}
	subtext := fmt.Sprintf("Error bars represent %.0f%% confidence interval", kCentilePercent)
	if filterExpr != "" {
		subtext = fmt.Sprintf("%s, benchmarks filter: '%s'", subtext, filterExpr)
	}
	return &trendChartSection{
		baseSection: baseSection{
			Type:            "trend_chart",
			Title:           title,
			SubText:         subtext,
			BenchmarkFilter: compileFilter(filterExpr),
		},
		Metric:  metric,
		ChartId: uniqueChartName(),
	}
}
