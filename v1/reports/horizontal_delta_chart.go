package reports

import (
	"fmt"

	"golang.org/x/perf/benchstat"
)

type horizontalDeltaChartSection struct {
	baseSection
	Metric          Metric
	ChartId         string
	NumBenchmarks   int
	ExperimentNames []string
	Deltas          []float64
	DeltaLabels     []string
	BarColors       []string
}

func (s *horizontalDeltaChartSection) fillData(dt *dataTableImpl) error {
	var table *benchstat.Table
	speedupColor, slowdownColor := "green", "red"
	switch s.Metric {
	case TimeOp:
		table = dt.timeOpTable
		s.XTitle = "Δ% time/op (lower is better)"
	case Speed:
		fallthrough
	case Throughput:
		table = dt.speedTable
		s.XTitle = "Δ% throughput (higher is better)"
		speedupColor, slowdownColor = slowdownColor, speedupColor
	case OpsPerSec:
		fallthrough
	case MsgPerSec:
		table = invertTimeOpTable(dt.timeOpTable, s.Metric)
		s.XTitle = "Δ% op/s (higher is better)"
		speedupColor, slowdownColor = slowdownColor, speedupColor

	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	if !table.OldNewDelta {
		return fmt.Errorf("Input table is not a comparison")
	}

	rows := filterByBenchmarkName(table.Rows, s.BenchmarkFilter)

	s.NumBenchmarks = len(rows)
	s.ExperimentNames = make([]string, s.NumBenchmarks)
	s.Deltas = make([]float64, s.NumBenchmarks)
	s.DeltaLabels = make([]string, s.NumBenchmarks)
	s.BarColors = make([]string, s.NumBenchmarks)

	for i, row := range rows {
		s.ExperimentNames[i] = row.Benchmark
		if row.Delta == "~" {
			s.Deltas[i] = 0
			s.DeltaLabels[i] = "inconclusive"
		} else {
			s.Deltas[i] = row.PctDelta
			s.DeltaLabels[i] = fmt.Sprintf("%+.1f%%", row.PctDelta)
		}
		if row.PctDelta < 0 {
			s.BarColors[i] = speedupColor
		} else {
			s.BarColors[i] = slowdownColor
		}
	}

	return nil
}

func HorizontalDeltaChart(title string, metric Metric, filterExpr string) SectionConfig {
	if title == "" {
		title = fmt.Sprintf("Relative %s comparison", metric)
	}
	return &horizontalDeltaChartSection{
		baseSection: baseSection{
			Type:            "horizontal_delta_chart",
			Title:           title,
			BenchmarkFilter: compileFilter(filterExpr),
		},
		Metric:  metric,
		ChartId: uniqueChartName(),
	}
}
