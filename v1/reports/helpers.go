package reports

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/mprimi/go-bench-away/v1/core"
	"golang.org/x/perf/benchstat"
)

func loadJobAndResults(client JobRecordClient, jobId string) (*core.JobRecord, []byte, error) {
	job, _, err := client.LoadJob(jobId)
	if err != nil {
		return nil, nil, err
	}

	if job.Status != core.Succeeded && job.Status != core.Failed {
		return nil, nil, fmt.Errorf("Job %s status is %v", job.Id, job.Status)
	}

	fmt.Printf("Loading job %s\n", jobId)
	const initialBufferSize = 1024
	buf := bytes.NewBuffer(make([]byte, 0, initialBufferSize))
	err = client.LoadResultsArtifact(job, buf)
	if err != nil {
		return nil, nil, err
	}

	return job, buf.Bytes(), nil
}

// Count the unique string in the slice
func countUnique(elements []string) int {
	set := make(map[string]struct{}, len(elements))
	for _, element := range elements {
		set[element] = struct{}{}
	}
	return len(set)
}

// Multiple jobs may use the same GitRef (e.g. when comparing two versions of go)
// This makes graphs and table hard to read, since the same ref appears.
// Try to compose a minimum label for each job that makes it unique
func createJobLabels(jobs []*core.JobRecord) []string {

	containsDuplicates := func(labels []string) bool {
		m := make(map[string]struct{}, len(labels))
		for _, l := range labels {
			if _, present := m[l]; present {
				return true
			}
			m[l] = struct{}{}
		}
		return false
	}

	// Function that creates a label from a job
	type LabelFunc func(*core.JobRecord) string

	labelFunctions := []LabelFunc{
		// Try GitRef (or short SHA if ref is the SHA)
		func(job *core.JobRecord) string {
			if job.Parameters.GitRef == job.SHA {
				return job.SHA[0:7]
			}
			return job.Parameters.GitRef
		},
		// Try GitRef + SHA (or just SHA if the GitRef is the SHA)
		func(job *core.JobRecord) string {
			if job.Parameters.GitRef == job.SHA {
				return job.SHA[0:7]
			}
			return fmt.Sprintf("%s [%s]", job.Parameters.GitRef, job.SHA[0:7])
		},
		// Try GitRef + Go version
		func(job *core.JobRecord) string { return fmt.Sprintf("%s [%s]", job.Parameters.GitRef, job.GoVersion) },
		// Last resort, use job ID
		func(job *core.JobRecord) string { return job.Id },
	}

	for _, f := range labelFunctions {
		labels := make([]string, len(jobs))
		for i, job := range jobs {
			labels[i] = f(job)
		}
		if !containsDuplicates(labels) {
			return labels
		}
	}

	panic("Could not construct a set of unique labels")
}

// Return a new unique name for a chart div
var chartCounter int

func uniqueChartName() string {
	chartCounter += 1
	return fmt.Sprintf("chart_%d", chartCounter)
}

func resetChartId() {
	//TODO this is a ugly hack necessary for creating deterministic graphs in tests, find a better way
	chartCounter = 0
}

func valueDeviationAndScaledString(m *benchstat.Metrics) (float64, float64, string) {
	if len(m.RValues) == 0 {
		return 0, 0, "no data"
	}
	mean := m.Mean
	scaler := benchstat.NewScaler(mean, m.Unit)
	centile, err := stats.Percentile(m.RValues, kCentilePercent)
	if err != nil {
		panic(fmt.Sprintf("Failed to calculate percentile for %T %+v: %v", m, m, err))
	}
	deviation := centile - mean
	scaledString := fmt.Sprintf("%s ± %s", scaler(mean), scaler(deviation))
	return mean, deviation, scaledString
}

func filterByBenchmarkName(inputRows []*benchstat.Row, filter *regexp.Regexp) []*benchstat.Row {
	if filter == nil {
		return inputRows
	}

	outputRows := make([]*benchstat.Row, 0, len(inputRows))
	for _, row := range inputRows {
		if filter.MatchString(row.Benchmark) {
			outputRows = append(outputRows, row)
		}
	}
	return outputRows
}

func compileFilter(filterExpr string) *regexp.Regexp {
	if filterExpr == "" {
		return nil
	}
	return regexp.MustCompile(filterExpr)
}

// Given a TimeOp table, construct and return a table with inverse values. e.g. 0.1 s/op -> 10 op/s
// All rows values are expected to be ns/op and converted to op/s.
func invertTimeOpTable(timeOpTable *benchstat.Table, metric Metric) *benchstat.Table {
	if timeOpTable.Metric != string(TimeOp) {
		panic(fmt.Sprintf("unexpected input metric: %s", timeOpTable.Metric))
	}

	nsOpToMsgPerSec := func(v float64) float64 {
		return 1 / v * float64(time.Second)
	}

	// N.B. benchstat calculates geometric mean differently see GeomMean:
	// https://cs.opensource.google/go/x/perf/+/d343f639:internal/stats/sample.go;l=152
	mean := func(vs []float64) float64 {
		if len(vs) == 0 {
			return math.NaN()
		}
		m := 0.0
		for _, x := range vs {
			ix := nsOpToMsgPerSec(x)
			if ix <= 0 {
				return math.NaN()
			}
			m += ix
		}
		return m / float64(len(vs))
	}

	opsPerSecondTable := &benchstat.Table{
		Metric:      string(metric),
		OldNewDelta: timeOpTable.OldNewDelta,
		Configs:     timeOpTable.Configs,
		Groups:      timeOpTable.Groups,
		Rows:        make([]*benchstat.Row, len(timeOpTable.Rows)),
	}

	for i, timeOpRow := range timeOpTable.Rows {
		opsPerSecondRow := &benchstat.Row{
			Benchmark: timeOpRow.Benchmark,
			Group:     timeOpRow.Group,
			Scaler:    nil,
			Metrics:   make([]*benchstat.Metrics, len(timeOpRow.Metrics)),
			Note:      timeOpRow.Note,
		}

		for j, timeOpMetric := range timeOpRow.Metrics {
			if len(timeOpMetric.Values) == 0 {
				// empty row, copy as-is
				opsPerSecondRow.Metrics[j] = timeOpMetric
				continue
			}
			if timeOpMetric.Unit != "ns/op" {
				panic(fmt.Sprintf("unexpected unit: %s", timeOpMetric.Unit))
			}
			opsPerSecondMetric := &benchstat.Metrics{
				Unit:    string(metric),
				Values:  make([]float64, len(timeOpMetric.Values)),
				RValues: make([]float64, len(timeOpMetric.RValues)),
				Min:     nsOpToMsgPerSec(timeOpMetric.Max),
				Mean:    mean(timeOpMetric.RValues),
				Max:     nsOpToMsgPerSec(timeOpMetric.Min),
			}

			for k, value := range timeOpMetric.Values {
				opsPerSecondMetric.Values[k] = nsOpToMsgPerSec(value)
			}
			for k, value := range timeOpMetric.RValues {
				opsPerSecondMetric.RValues[k] = nsOpToMsgPerSec(value)
			}

			opsPerSecondRow.Metrics[j] = opsPerSecondMetric
		}

		// For side-by-side comparison tables (2 result sets), recalculate percentage differences
		if timeOpTable.OldNewDelta {
			if len(opsPerSecondRow.Metrics) != 2 {
				panic(fmt.Sprintf("unexpected number of metrics in comparison table: %d", len(opsPerSecondRow.Metrics)))
			}

			// `Change` values are -1, 0, 1, just flip the sign
			opsPerSecondRow.Change = -timeOpRow.Change

			// PctDelta needs to be re-calculated (can't just flip the sign)
			// This is a simplified calculation using means, and assumes significance testing carries over
			// from the time/op calculation. Not as precise, but it will do.
			meanBefore, meanAfter := opsPerSecondRow.Metrics[0].Mean, opsPerSecondRow.Metrics[1].Mean
			opsPerSecondRow.PctDelta = 100 * (meanAfter - meanBefore) / (meanBefore/2 + meanAfter/2)
			if timeOpRow.Delta == "~" {
				// Delta is not statistically significant
				opsPerSecondRow.Delta = timeOpRow.Delta
			} else {
				opsPerSecondRow.Delta = fmt.Sprintf("%.1f%%", opsPerSecondRow.PctDelta)
			}
		}

		opsPerSecondTable.Rows[i] = opsPerSecondRow
	}

	return opsPerSecondTable
}
