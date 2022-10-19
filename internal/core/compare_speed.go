package core

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/montanaflynn/stats"
	"github.com/nats-io/nats.go"
	"golang.org/x/perf/benchstat"
	"html/template"
	"io"
	"os"
)

//go:embed html/compare-speed.html.tmpl
var compareSpeedHtmlTmpl string

type reportInputs struct {
	parameters CompareSpeedParameters
	oldJob     JobRecord
	newJob     JobRecord
	oldResults []byte
	newResults []byte
}

func CompareSpeed(js nats.JetStreamContext, params CompareSpeedParameters) error {

	// Create jobs KV store
	kv, err := js.KeyValue(jobRecordsStoreName)
	if err != nil {
		return fmt.Errorf("KV store lookup error: %v", err)
	}

	// Bind ObjectStore
	obs, err := js.ObjectStore(artifactsStoreName)
	if err != nil {
		return fmt.Errorf("Failed to bind Object store: %v", err)
	}

	ri := reportInputs{
		parameters: params,
	}

	var res []byte

	j1, res, err := fetchJobAndResults(kv, obs, params.OldJobId)
	if err != nil {
		return err
	}
	ri.oldJob = *j1
	ri.oldResults = res

	j2, res, err := fetchJobAndResults(kv, obs, params.NewJobId)
	if err != nil {
		return err
	}
	ri.newJob = *j2
	ri.newResults = res

	fmt.Printf("Creating speed report comparing:\n")
	fmt.Printf("[%s] rev: %s (SHA:%s from %s)\n", params.OldJobId, j1.Parameters.GitRef, j1.SHA, j1.Parameters.GitRemote)
	fmt.Printf("[%s] rev: %s (SHA:%s from %s)\n", params.NewJobId, j2.Parameters.GitRef, j2.SHA, j2.Parameters.GitRemote)

	err = createReport(ri)
	if err != nil {
		return err
	}

	return nil
}

func fetchJobAndResults(kv nats.KeyValue, obs nats.ObjectStore, jobId string) (*JobRecord, []byte, error) {
	jobKey := fmt.Sprintf(jobRecordKeyTemplate, jobId)
	kve, err := kv.Get(jobKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to retrieve job record (%s): %v", jobKey, err)
	}

	job := loadJob(kve.Value())
	if job.Status != Succeeded {
		return nil, nil, fmt.Errorf("Job %s is not completed successfully (status: %s)", jobId, job.Status)
	}

	if job.Results == "" {
		return nil, nil, fmt.Errorf("Job %s has no results artifact", jobId)
	}

	obj, err := obs.Get(job.Results)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to retrieve job results: %v", err)
	}

	resultsBytes, err := io.ReadAll(obj)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read job results: %v", err)
	}

	return job, resultsBytes, nil
}

type ExperimentRow struct {
	ExperimentName string
	OldSpeed       string
	NewSpeed       string
	Delta          string
}

type InfoRow struct {
	RowName  string
	OldValue string
	NewValue string
}

type TemplateValues struct {
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
}

func createReport(ri reportInputs) error {

	j1, j2 := ri.oldJob, ri.newJob
	jp1, jp2 := j1.Parameters, j2.Parameters
	infoRows := []InfoRow{
		{"Remote", jp1.GitRemote, jp2.GitRemote},
		{"Ref", jp1.GitRef, jp2.GitRef},
		{"SHA", j1.SHA, j2.SHA},
		{"Filter", jp1.TestsFilterExpr, jp2.TestsFilterExpr},
		{"Reps", fmt.Sprintf("%d x %v", jp1.Reps, jp1.TestMinRuntime), fmt.Sprintf("%d x %v", jp2.Reps, jp2.TestMinRuntime)},
		{"Job", j1.Id, j2.Id},
	}

	c := &benchstat.Collection{
		Alpha:      0.1,
		AddGeoMean: false,
		DeltaTest:  benchstat.UTest,
		Order:      nil, // Preserve file add order
	}

	c.AddConfig("first", ri.oldResults)
	c.AddConfig("second", ri.newResults)

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

		oldVariance, _ := stats.SampleVariance(row.Metrics[0].RValues)
		newVariance, _ := stats.SampleVariance(row.Metrics[1].RValues)
		oldSpeedErrs[i] = oldVariance
		newSpeedErrs[i] = newVariance

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
			OldSpeed:       fmt.Sprintf("%.1fMB/s ± %.1f", oldSpeeds[i], oldVariance),
			NewSpeed:       fmt.Sprintf("%.1fMB/s ± %.1f", newSpeeds[i], newVariance),
			Delta:          row.Delta,
		}

		if experimentRows[i].Delta == "~" {
			experimentRows[i].Delta = "Inconclusive"
		}
	}

	tv := TemplateValues{
		Title:            template.HTML(ri.parameters.Title),
		ExperimentNames:  mustMarshal(testNames),
		OldName:          template.HTML(ri.parameters.OldJobName),
		OldSpeeds:        mustMarshal(oldSpeeds),
		OldSpeedErrors:   mustMarshal(oldSpeedErrs),
		OldSpeedLabels:   mustMarshal(oldSpeedLabels),
		NewName:          template.HTML(ri.parameters.NewJobName),
		NewSpeeds:        mustMarshal(newSpeeds),
		NewSpeedErrors:   mustMarshal(newSpeedErrs),
		NewSpeedLabels:   mustMarshal(newSpeedLabels),
		Deltas:           mustMarshal(deltas),
		DeltaLabels:      mustMarshal(deltaLabels),
		DeltaColors:      mustMarshal(deltaColors),
		ExperimentsTable: experimentRows,
		InfoTable:        infoRows,
	}

	f, err := os.Create(ri.parameters.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reportTemplate := template.Must(template.New("compare_speed").Parse(compareSpeedHtmlTmpl))
	err = reportTemplate.Execute(f, tv)
	if err != nil {
		return err
	}

	fmt.Printf("Created report: %s", ri.parameters.OutputPath)
	return nil
}

func mustMarshal(v any) template.JS {
	encoded, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Failed to encode %v: %v", v, err))
	}
	return template.JS(encoded)
}
