package reports

import (
	"fmt"
	"golang.org/x/perf/benchstat"
)

type horizontalBoxChartBox struct {
	Name   string
	Values []float64
	Labels []string
}

type horizontalBoxChartSection struct {
	baseSection
	Metric        Metric
	ChartId       string
	NumBenchmarks int
	Experiments   []horizontalBoxChartBox
}

func (s *horizontalBoxChartSection) fillData(dt *dataTableImpl) error {
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

	s.Experiments = make([]horizontalBoxChartBox, s.NumBenchmarks)

	for i, row := range rows {
		metrics := row.Metrics[0]
		s.Experiments[i] = horizontalBoxChartBox{
			Name:   row.Benchmark,
			Values: metrics.Values,
			Labels: make([]string, len(metrics.Values)),
		}

		for j, value := range metrics.Values {
			s.Experiments[i].Labels[j] = row.Scaler(value)
		}
	}
	return nil
}

func HorizontalBoxChart(title string, metric Metric, filterExpr string) SectionConfig {
	if title == "" {
		title = fmt.Sprintf("%s results distribution", metric)
	}
	subtext := ""
	if filterExpr != "" {
		subtext = fmt.Sprintf("Filter: '%s'", filterExpr)
	}

	return &horizontalBoxChartSection{
		baseSection: baseSection{
			Type:            "horizontal_box_chart",
			Title:           title,
			SubText:         subtext,
			BenchmarkFilter: compileFilter(filterExpr),
		},
		Metric:  metric,
		ChartId: uniqueChartName(),
	}
}
