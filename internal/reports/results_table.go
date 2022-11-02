package reports

import (
	"fmt"
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
	case Speed:
		table = dt.speedTable
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
			_, _, tr.Values[j] = valueDeviationAndScaledString(m)

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
