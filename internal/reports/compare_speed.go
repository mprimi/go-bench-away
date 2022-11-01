package reports

import (
	_ "embed"
	"fmt"
	"github.com/montanaflynn/stats"
	"golang.org/x/perf/benchstat"
	"html/template"
	"os"

	"github.com/mprimi/go-bench-away/internal/client"
)

//go:embed html/compare-speed.html.tmpl
var compareSpeedHtmlTmpl string

type CompareSpeedConfig struct {
	Title      string
	OldJobId   string
	NewJobId   string
	OldJobName string
	NewJobName string
	OutputPath string
}

func CreateCompareSpeedReport(client client.Client, cfg *CompareSpeedConfig) error {

	// Create collection of results
	c := &benchstat.Collection{
		Alpha:      0.1,
		AddGeoMean: false,
		DeltaTest:  benchstat.UTest,
		Order:      nil, // Preserve file add order
	}

	j1, j1Results, err := loadJobAndResults(client, cfg.OldJobId)
	if err != nil {
		return err
	}

	j2, j2Results, err := loadJobAndResults(client, cfg.NewJobId)
	if err != nil {
		return err
	}

	// Add 2 sets of results
	c.AddConfig(j1.Id, j1Results)
	c.AddConfig(j2.Id, j2Results)

	// Look for 'speed' table (only present if benchmarks are reporting throughput)
	tables := c.Tables()
	if len(tables) == 0 {
		return fmt.Errorf("No comparison tables, the jobs may not overlap in tests executed")
	} else if len(tables) < 2 || tables[1].Metric != "speed" {
		// Benchmarks must report speed using https://pkg.go.dev/testing#B.SetBytes
		// For this report to generate
		return fmt.Errorf("Speed table not found, the benchmarks may not be reporting throughput")
	}

	speedTable := tables[1]
	if !speedTable.OldNewDelta {
		return fmt.Errorf("Speed table is of type old/new/delta")
	} else if len(speedTable.Configs) != 2 {
		return fmt.Errorf("Unexpected number of configurations: %d", len(speedTable.Configs))
	}

	// Shorthands
	jp1, jp2 := j1.Parameters, j2.Parameters

	if cfg.OldJobName == "" {
		cfg.OldJobName = jp1.GitRef
	}

	if cfg.NewJobName == "" {
		cfg.NewJobName = jp2.GitRef
	}

	fmt.Printf("Creating speed report comparing:\n")
	fmt.Printf(
		"[%s] rev: %s (SHA:%s from %s)\n"+
			"[%s] rev: %s (SHA:%s from %s)\n"+
			"",
		cfg.OldJobId,
		jp1.GitRef,
		j1.SHA,
		jp1.GitRemote,
		cfg.NewJobId,
		jp1.GitRef,
		j2.SHA,
		jp2.GitRemote,
	)

	// Template values for jobs summary table
	type InfoRow struct {
		RowName  string
		OldValue string
		NewValue string
	}

	infoRows := []InfoRow{
		{"Remote", jp1.GitRemote, jp2.GitRemote},
		{"Ref", jp1.GitRef, jp2.GitRef},
		{"SHA", j1.SHA, j2.SHA},
		{"Filter", jp1.TestsFilterExpr, jp2.TestsFilterExpr},
		{"Reps", fmt.Sprintf("%d x %v", jp1.Reps, jp1.TestMinRuntime), fmt.Sprintf("%d x %v", jp2.Reps, jp2.TestMinRuntime)},
		{"Job", j1.Id, j2.Id},
	}

	// Prepare table of value

	type ExperimentRow struct {
		ExperimentName string
		OldSpeed       string
		NewSpeed       string
		Delta          string
	}

	numBenchmarks := len(speedTable.Rows)

	testNames := make([]string, numBenchmarks)
	oldSpeeds := make([]float64, numBenchmarks)
	oldSpeedErrs := make([]float64, numBenchmarks)
	oldSpeedLabels := make([]string, numBenchmarks)
	newSpeeds := make([]float64, numBenchmarks)
	newSpeedErrs := make([]float64, numBenchmarks)
	newSpeedLabels := make([]string, numBenchmarks)
	deltas := make([]float64, numBenchmarks)
	deltaLabels := make([]string, numBenchmarks)
	deltaColors := make([]string, numBenchmarks)
	experimentRows := make([]ExperimentRow, numBenchmarks)

	for i, row := range speedTable.Rows {
		testNames[i] = fmt.Sprintf("[#%d] %s", i+1, row.Benchmark)
		oldSpeeds[i] = row.Metrics[0].Mean
		newSpeeds[i] = row.Metrics[1].Mean

		oldSpeedErrs[i] = 0
		if len(row.Metrics[0].RValues) > 1 {
			oldVariance, err := stats.SampleVariance(row.Metrics[0].RValues)
			if err != nil {
				return err
			}
			oldSpeedErrs[i] = oldVariance
		}

		newSpeedErrs[i] = 0
		if len(row.Metrics[1].RValues) > 1 {
			newVariance, err := stats.SampleVariance(row.Metrics[1].RValues)
			if err != nil {
				return err
			}
			newSpeedErrs[i] = newVariance
		}

		oldSpeedLabels[i] = fmt.Sprintf("%.1fMB/s", oldSpeeds[i])
		newSpeedLabels[i] = fmt.Sprintf("%.1fMB/s", newSpeeds[i])

		deltas[i] = row.PctDelta
		deltaLabels[i] = fmt.Sprintf("%.1f%%", row.PctDelta)
		deltaColors[i] = "green"
		if row.PctDelta < 0 {
			deltaColors[i] = "red"
		}

		experimentRows[i] = ExperimentRow{
			ExperimentName: testNames[i],
			OldSpeed:       fmt.Sprintf("%.1fMB/s ± %.1f", oldSpeeds[i], oldSpeedErrs[i]),
			NewSpeed:       fmt.Sprintf("%.1fMB/s ± %.1f", newSpeeds[i], newSpeedErrs[i]),
			Delta:          row.Delta,
		}

		if experimentRows[i].Delta == "~" {
			experimentRows[i].Delta = "Inconclusive"
		}
	}

	// Pivot values to feed into report template
	tv := struct {
		Title            template.HTML
		ExperimentNames  template.JS
		OldName          template.HTML
		OldSpeeds        template.JS
		OldSpeedErrors   template.JS
		OldSpeedLabels   template.JS
		NewName          template.HTML
		NewSpeeds        template.JS
		NewSpeedErrors   template.JS
		NewSpeedLabels   template.JS
		Deltas           template.JS
		DeltaLabels      template.JS
		DeltaColors      template.JS
		ExperimentsTable []ExperimentRow
		InfoTable        []InfoRow
	}{
		Title:            template.HTML(cfg.Title),
		ExperimentNames:  mustMarshal(testNames),
		OldName:          template.HTML(cfg.OldJobName),
		OldSpeeds:        mustMarshal(oldSpeeds),
		OldSpeedErrors:   mustMarshal(oldSpeedErrs),
		OldSpeedLabels:   mustMarshal(oldSpeedLabels),
		NewName:          template.HTML(cfg.NewJobName),
		NewSpeeds:        mustMarshal(newSpeeds),
		NewSpeedErrors:   mustMarshal(newSpeedErrs),
		NewSpeedLabels:   mustMarshal(newSpeedLabels),
		Deltas:           mustMarshal(deltas),
		DeltaLabels:      mustMarshal(deltaLabels),
		DeltaColors:      mustMarshal(deltaColors),
		ExperimentsTable: experimentRows,
		InfoTable:        infoRows,
	}

	f, err := os.Create(cfg.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reportTemplate := template.Must(template.New("compare_speed").Parse(compareSpeedHtmlTmpl))
	err = reportTemplate.Execute(f, tv)
	if err != nil {
		return err
	}

	return nil
}
