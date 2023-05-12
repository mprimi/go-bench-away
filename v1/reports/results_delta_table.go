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
	Hidden      bool
}

func (s *resultsDeltaTableSection) fillData(dt *dataTableImpl) error {

	var table *benchstat.Table
	switch s.Metric {
	case TimeOp:
		table = dt.timeOpTable
	case Speed:
		fallthrough
	case Throughput:
		table = dt.speedTable
	case OpsPerSec:
		fallthrough
	case MsgPerSec:
		table = invertTimeOpTable(dt.timeOpTable, s.Metric)
	default:
		return fmt.Errorf("Unknow table metric: %s", s.Metric)
	}

	if !table.OldNewDelta {
		return fmt.Errorf("Input table is not a comparison")
	}

	rows := filterByBenchmarkName(table.Rows, s.BenchmarkFilter)

	s.JobLabels = dt.jobLabels
	s.ResultsRows = make([]resultsDeltaRow, len(rows))

	for i, row := range rows {
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

func ResultsDeltaTable(metric Metric, filterExpr string, hidden bool) SectionConfig {
	return &resultsDeltaTableSection{
		baseSection: baseSection{
			Type:            "results_delta_table",
			Title:           "",
			BenchmarkFilter: compileFilter(filterExpr),
		},
		Metric: metric,
		Hidden: hidden,
	}
}
