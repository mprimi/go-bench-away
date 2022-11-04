package reports

import (
	"fmt"
	"golang.org/x/perf/benchstat"
)

type resultsDeltaRow struct {
	BenchmarkName string
	Values        []string
}

type resultsDeltaTableSection struct {
	baseSection
	Metric      Metric
	JobLabels   []string
	ResultsRows []resultsDeltaRow
}

func (s *resultsDeltaTableSection) fillData(dt *dataTableImpl) error {

	var table *benchstat.Table
	switch s.Metric {
	case TimeOp:
		table = dt.timeOpTable
	case Speed:
		table = dt.speedTable
	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	if !table.OldNewDelta {
		return fmt.Errorf("Input table is not a comparison")
	}

	s.JobLabels = dt.jobLabels
	s.ResultsRows = make([]resultsDeltaRow, len(table.Rows))

	for i, row := range table.Rows {
		tr := &s.ResultsRows[i]
		tr.BenchmarkName = row.Benchmark
		tr.Values = make([]string, len(s.JobLabels)+1)
		for j, m := range row.Metrics {
			_, _, tr.Values[j] = valueDeviationAndScaledString(m)
		}

		if row.Delta == "~" {
			tr.Values[len(s.JobLabels)] = "Inconclusive"
		} else {
			tr.Values[len(s.JobLabels)] = fmt.Sprintf("%+.1f%%", row.PctDelta)
		}

	}

	return nil
}

func ResultsDeltaTable(metric Metric) SectionConfig {
	return &resultsDeltaTableSection{
		baseSection: baseSection{
			Type:  "results_delta_table",
			Title: "",
		},
		Metric: metric,
	}
}
