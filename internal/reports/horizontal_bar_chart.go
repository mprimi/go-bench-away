package reports

import (
	"fmt"
	"golang.org/x/perf/benchstat"
)

type horizontalBarChartGroup struct {
	Name            string
	ExperimentNames []string
	Averages        []float64
	Deviation       []float64
	BarLabels       []string
	HoverLabels     []string
}

type horizontalBarChartSection struct {
	baseSection
	Metric        Metric
	ChartId       string
	NumBenchmarks int
	Groups        []horizontalBarChartGroup
}

func (s *horizontalBarChartSection) fillData(dt *dataTableImpl) error {
	var table *benchstat.Table
	switch s.Metric {
	case TimeOp:
		table = dt.timeOpTable
		s.XTitle = "Time/op (lower is better)"
	case Speed:
		table = dt.speedTable
		s.XTitle = "Throughput (higher is better)"
	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	rows := filterByBenchmarkName(table.Rows, s.BenchmarkFilter)

	s.NumBenchmarks = len(rows)
	experimentNames := make([]string, s.NumBenchmarks)

	for i, row := range rows {
		experimentNames[i] = row.Benchmark
	}

	s.Groups = make([]horizontalBarChartGroup, len(dt.jobs))

	for i := range dt.jobs {
		g := &s.Groups[i]
		g.Name = dt.jobLabels[i]
		g.ExperimentNames = experimentNames
		g.Averages = make([]float64, s.NumBenchmarks)
		g.Deviation = make([]float64, s.NumBenchmarks)

		g.BarLabels = make([]string, s.NumBenchmarks)
		g.HoverLabels = make([]string, s.NumBenchmarks)

		for j, row := range rows {
			m := row.Metrics[i]
			g.Averages[j], g.Deviation[j], g.BarLabels[j] = valueDeviationAndScaledString(m)
			g.HoverLabels[j] = g.BarLabels[j]
		}
	}
	return nil
}

func HorizontalBarChart(title string, metric Metric, filterExpr string) SectionConfig {
	if title == "" {
		title = fmt.Sprintf("%s comparison", metric)
	}
	subtext := fmt.Sprintf("Error bars represent %.0f%% confidence interval", kCentilePercent)
	if filterExpr != "" {
		subtext = fmt.Sprintf("%s, benchmarks filter: '%s'", subtext, filterExpr)
	}
	return &horizontalBarChartSection{
		baseSection: baseSection{
			Type:            "horizontal_bar_chart",
			Title:           title,
			SubText:         subtext,
			BenchmarkFilter: compileFilter(filterExpr),
		},
		Metric:  metric,
		ChartId: uniqueChartName(),
	}
}
