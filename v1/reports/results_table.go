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
	Hidden      bool
}

func (s *resultsTableSection) fillData(dt *dataTableImpl) error {

	var table *benchstat.Table
	switch s.Metric {
	case TimeOp:
		table = dt.timeOpTable
	case Speed:
		table = dt.speedTable
	case OpsPerSec:
		fallthrough
	case MsgPerSec:
		table = invertTimeOpTable(dt.timeOpTable, s.Metric)
	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	rows := filterByBenchmarkName(table.Rows, s.BenchmarkFilter)

	s.JobLabels = dt.jobLabels
	s.ResultsRows = make([]resultsRow, len(rows))

	for i, row := range rows {
		tr := &s.ResultsRows[i]
		tr.BenchmarkName = row.Benchmark
		tr.Values = make([]string, len(s.JobLabels))
		for j, m := range row.Metrics {
			_, _, tr.Values[j] = valueDeviationAndScaledString(m)

		}
	}

	return nil
}

func ResultsTable(metric Metric, filterExpr string, hidden bool) SectionConfig {
	return &resultsTableSection{
		baseSection: baseSection{
			Type:            "results_table",
			Title:           "",
			BenchmarkFilter: compileFilter(filterExpr),
		},
		Metric: metric,
		Hidden: hidden,
	}
}
