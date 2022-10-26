package reports

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
)

func loadJobAndResults(client client.Client, jobId string) (*core.JobRecord, []byte, error) {
	job, _, err := client.LoadJob(jobId)
	if err != nil {
		return nil, nil, err
	}

	if job.Status != core.Succeeded {
		return nil, nil, fmt.Errorf("Job %s status is %v", job.Id, job.Status)
	}

	fmt.Printf("Loading job %s\n", jobId)
	results, err := client.LoadResultsArtifact(job)
	if err != nil {
		return nil, nil, err
	}

	return job, results, nil
}

func mustMarshal(v any) template.JS {
	encoded, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Failed to encode %v: %v", v, err))
	}
	return template.JS(encoded)
}

func mustMarshalIndent(v any, indentSpaces, leftMarginSpaces int) template.JS {
	prefix := strings.Repeat(" ", leftMarginSpaces)
	indent := strings.Repeat(" ", indentSpaces)
	encoded, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		panic(fmt.Sprintf("Failed to encode %v: %v", v, err))
	}
	return template.JS(encoded)
}
