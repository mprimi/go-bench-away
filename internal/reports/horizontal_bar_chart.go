package reports

import (
	"fmt"
	"github.com/montanaflynn/stats"
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
		s.SubText = "Higher is better"
		s.XTitle = "Throughput (higher is better)"
	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	s.NumBenchmarks = len(table.Rows)
	experimentNames := make([]string, s.NumBenchmarks)

	for i, row := range table.Rows {
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

		for j, row := range table.Rows {
			m := row.Metrics[i]
			g.Averages[j] = m.Mean
			centile, err := stats.Percentile(m.RValues, kCentilePercent)
			if err != nil {
				return err
			}
			g.Deviation[j] = centile - m.Mean
			scaler := benchstat.NewScaler(m.Mean, m.Unit)
			g.BarLabels[j] = scaler(m.Mean)
			g.HoverLabels[j] = fmt.Sprintf("%s ± %s", scaler(m.Mean), scaler(g.Deviation[j]))
		}
	}
	return nil
}

func HorizontalBarChart(metric Metric) SectionConfig {
	return &horizontalBarChartSection{
		baseSection: baseSection{
			Type:  "horizontal_bar_chart",
			Title: "Side by side",
		},
		Metric:  metric,
		ChartId: uniqueChartName(),
	}
}
