package reports

import (
	"fmt"
	"golang.org/x/perf/benchstat"
	"github.com/montanaflynn/stats"
	"github.com/mprimi/go-bench-away/internal/core"
)

type trendChartSeries struct {
	BenchmarkName   string
	JobIds          []string
	Values          []float64
	Deviation       []float64
	HoverLabels     []string
}

type trendChartSection struct {
	baseSection
	Metric          Metric
	ChartId         string
	NumBenchmarks   int
	JobLabels       []string
	JobIds          []string
	Series          []trendChartSeries
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
	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	s.NumBenchmarks = len(table.Rows)
	s.JobLabels = dt.jobLabels
	s.Series = make([]trendChartSeries, s.NumBenchmarks)

	s.JobIds = dt.mapJobs(func (job *core.JobRecord)(string){return job.Id})

	for i, row := range table.Rows {
		sr := &s.Series[i]
		sr.BenchmarkName = row.Benchmark
		sr.JobIds = s.JobIds

		sr.Values = make([]float64, len(s.JobIds))
		sr.Deviation = make([]float64, len(s.JobIds))
		sr.HoverLabels = make([]string, len(s.JobIds))

		for j, m := range row.Metrics {

			sr.Values[j] = m.Mean
			centile, err := stats.Percentile(m.RValues, kCentilePercent)
			if err != nil {
				return err
			}
			sr.Deviation[j] = centile - m.Mean

			scaler := benchstat.NewScaler(m.Mean, m.Unit)
			sr.HoverLabels[j] = fmt.Sprintf("%s ± %s", scaler(m.Mean), scaler(sr.Deviation[j]))
		}
	}

	return nil
}

func TrendChart(metric Metric) SectionConfig {
	return &trendChartSection{
		baseSection: baseSection{
			Type:  "trend_chart",
			Title: "Trend",
		},
		Metric:  metric,
		ChartId: uniqueChartName(),
	}
}
