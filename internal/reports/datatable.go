package reports

import (
	// _ "embed"
	"fmt"
	// "github.com/montanaflynn/stats"
	"golang.org/x/perf/benchstat"
	// "html/template"
	// "os"

	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
)

type DataTable interface {
	HasSpeed() bool
}

type dataTableImpl struct {
	jobs        []*core.JobRecord
	jobLabels   []string
	collection  benchstat.Collection
	timeOpTable *benchstat.Table
	speedTable  *benchstat.Table
}

func (dt *dataTableImpl) HasSpeed() bool {
	return dt.speedTable != nil
}

func CreateDataTable(client client.Client, jobIds ...string) (DataTable, error) {
	if len(jobIds) == 0 {
		return nil, fmt.Errorf("No jobs provided")
	} else if countUnique(jobIds) != len(jobIds) {
		return nil, fmt.Errorf("The list of job IDs contains duplicates")
	}

	dataTable := dataTableImpl{
		jobs: make([]*core.JobRecord, len(jobIds)),
		collection: benchstat.Collection{
			Alpha:      kDeltaTestAlpha,
			AddGeoMean: false,
			DeltaTest:  benchstat.UTest,
			Order:      nil, // Preserve order
		},
	}

	for i, jobId := range jobIds {
		job, results, err := loadJobAndResults(client, jobId)
		if err != nil {
			return nil, err
		}
		dataTable.jobs[i] = job
		dataTable.collection.AddConfig(jobId, results)
	}

	dataTable.jobLabels = createJobLabels(dataTable.jobs)

	if len(dataTable.collection.Tables()) == 0 {
		return nil, fmt.Errorf("Jobs don't overlap in benchmarks,")
	}

	for _, table := range dataTable.collection.Tables() {
		switch table.Metric {
		case string(TimeOp):
			dataTable.timeOpTable = table
		case string(Speed):
			dataTable.speedTable = table
		default:
			fmt.Printf("Ignoring results metric '%s'\n", table.Metric)
		}
	}

	return &dataTable, nil
}
