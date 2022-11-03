package reports

import (
	_ "embed"
	"fmt"
	"github.com/mprimi/go-bench-away/internal/client"
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
