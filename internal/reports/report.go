package reports

import (
	_ "embed"
	"fmt"
	"github.com/montanaflynn/stats"
	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
	"golang.org/x/perf/benchstat"
	"html/template"
	"os"
)

//go:embed html/report.html.tmpl
var reportHtmlTmpl string

type SectionConfig interface {
	fillData(dt *dataTableImpl) error
}

type SectionType string

type baseSection struct {
	Type    SectionType
	Title   string
	SubText string
	XTitle  string
	YTitle  string
}

type jobsTableSection struct {
	baseSection
	Jobs []*core.JobRecord
}

func (s *jobsTableSection) fillData(dt *dataTableImpl) error {
	s.Jobs = dt.jobs
	return nil
}

func JobsTable() SectionConfig {
	return &jobsTableSection{
		baseSection: baseSection{
			Type:  "jobs_table",
			Title: "Jobs",
		},
	}
}

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

type Metric string

const (
	TimeOp = Metric("time/op")
	Speed  = Metric("speed")
)

type ReportConfig struct {
	Title      string
	OutputPath string
	sections   []SectionConfig
	verbose    bool
}

func (r *ReportConfig) AddSections(sections ...SectionConfig) *ReportConfig {
	r.sections = append(r.sections, sections...)
	return r
}

func (r *ReportConfig) Verbose() *ReportConfig {
	r.verbose = true
	return r
}

func (r *ReportConfig) Log(format string, args ...any) {
	if r.verbose {
		fmt.Printf("[debug] "+format+"\n", args...)
	}
}

func CreateReport(client client.Client, cfg *ReportConfig, dataTable DataTable) error {
	dt := dataTable.(*dataTableImpl)
	title := cfg.Title
	if title == "" {
		title = fmt.Sprintf("Performance report (%d result sets)", len(dt.jobs))
	}

	cfg.Log("Generating report '%s'", title)

	for i, section := range cfg.sections {
		cfg.Log("Generating section %d/%d: %T: %+v", i+1, len(cfg.sections), section, section)
		err := section.fillData(dt)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(cfg.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.New("report")
	t = template.Must(t.Parse(reportHtmlTmpl))

	tv := struct {
		Title    string
		Sections []SectionConfig
	}{
		Title:    title,
		Sections: cfg.sections,
	}

	err = t.Execute(f, tv)
	if err != nil {
		return err
	}

	return nil
}
