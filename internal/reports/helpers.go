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

// Given a list of job Ids, return a list with duplicate removed (maintaining order)
func filterDuplicates(jobIds []string) []string {
	counts := make(map[string]struct{}, len(jobIds))
	for _, jobId := range jobIds {
		if _, present := counts[jobId]; present {
			fmt.Printf("Warning, duplicate job: %s\n", jobId)
		} else {
			counts[jobId] = struct{}{}
		}
	}
	uniqueJobIds := make([]string, 0, len(counts))
	for jobId := range counts {
		uniqueJobIds = append(uniqueJobIds, jobId)
	}
	return uniqueJobIds
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
		// Try GitRef
		func(job *core.JobRecord) string { return job.Parameters.GitRef },
		// Try GitRef + SHA
		func(job *core.JobRecord) string { return fmt.Sprintf("%s [%s]", job.Parameters.GitRef, job.SHA[0:7]) },
		// Try GitRef + Go version
		func(job *core.JobRecord) string { return fmt.Sprintf("%s [%s]", job.Parameters.GitRef, job.GoVersion) },
		// Last resort.. use job ID
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
