package reports

import (
	"fmt"
	"github.com/montanaflynn/stats"
	"golang.org/x/perf/benchstat"
)

type resultsRow struct {
	BenchmarkName string
	Values        []string
}

type resultsTableSection struct {
	baseSection
	Metric      Metric
	JobLabels   []string
	ResultsRows []resultsRow
}

func (s *resultsTableSection) fillData(dt *dataTableImpl) error {

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

	s.JobLabels = dt.jobLabels
	s.ResultsRows = make([]resultsRow, len(table.Rows))

	for i, row := range table.Rows {
		tr := &s.ResultsRows[i]
		tr.BenchmarkName = row.Benchmark
		tr.Values = make([]string, len(s.JobLabels))
		for j, m := range row.Metrics {
			centile, err := stats.Percentile(m.RValues, kCentilePercent)
			if err != nil {
				return err
			}
			deviation := centile - m.Mean
			scaler := benchstat.NewScaler(m.Mean, m.Unit)
			tr.Values[j] = fmt.Sprintf("%s ± %s", scaler(m.Mean), scaler(deviation))
		}
	}

	return nil
}

func ResultsTable(metric Metric) SectionConfig {
	return &resultsTableSection{
		baseSection: baseSection{
			Type:  "results_table",
			Title: "",
		},
		Metric: metric,
	}
}
