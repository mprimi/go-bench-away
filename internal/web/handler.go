package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/mprimi/go-bench-away/v1/core"
	"github.com/mprimi/go-bench-away/v1/reports"
)

//go:embed html/index.html.tmpl
var indexTmpl string

//go:embed html/queue.html.tmpl
var queueTmpl string

var jobResourceRegexp = regexp.MustCompile(`^/job/([[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12})/(log|script|results|record|plot|cancel)/?$`)

type handler struct {
	client        WebClient
	indexTemplate *template.Template
	queueTemplate *template.Template
}

func NewHandler(c WebClient) http.Handler {
	return &handler{
		client:        c,
		indexTemplate: template.Must(template.New("index").Parse(indexTmpl)),
		queueTemplate: template.Must(template.New("queue").Parse(queueTmpl)),
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Reject anything that is not a GET
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Invalid request method: %s", r.Method), http.StatusMethodNotAllowed)
		return
	}

	url := r.URL
	path := url.Path

	fmt.Printf(" > %s %s\n", r.Method, path)

	var err error
	if path == "" || path == "/" {
		err = h.serveIndex(w)
	} else if path == "/queue" || path == "/queue/" {
		err = h.serveQueue(w)
	} else if strings.HasPrefix(path, "/job/") {
		groupMatches := jobResourceRegexp.FindStringSubmatch(path)
		if groupMatches == nil || len(groupMatches) != 3 {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		jobId, resource := groupMatches[1], groupMatches[2]

		err = h.serveJobResource(w, jobId, resource)
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
	} else {
		fmt.Printf("Ok\n")
	}
}

func (h *handler) serveIndex(w http.ResponseWriter) error {

	qs, err := h.client.GetQueueStatus()
	if err != nil {
		return err
	}

	return h.indexTemplate.Execute(w, qs)
}

func (h *handler) serveQueue(w http.ResponseWriter) error {

	jobRecords, err := h.client.LoadRecentJobs(10)
	if err != nil {
		return err
	}

	tv := struct {
		QueueName string
		Jobs      []*core.JobRecord
	}{
		QueueName: h.client.QueueName(),
		Jobs:      jobRecords,
	}
	return h.queueTemplate.Execute(w, tv)
}

func (h *handler) serveJobResource(w http.ResponseWriter, jobId, resourceType string) error {

	jobRecord, _, err := h.client.LoadJob(jobId)
	if err != nil {
		return fmt.Errorf("Failed to load job '%s': %v", jobId, err)
	}

	switch resourceType {
	case "log":
		err = h.client.LoadLogArtifact(jobRecord, w)
	case "script":
		err = h.client.LoadScriptArtifact(jobRecord, w)
	case "results":
		err = h.client.LoadResultsArtifact(jobRecord, w)
	case "record":
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		err = e.Encode(jobRecord)
	case "plot":
		err = h.serveJobResultsPlot(jobId, w)
	case "cancel":
		err = h.client.CancelJob(jobId)
		if err == nil {
			fmt.Fprintf(w, "Job %s cancelled", jobId)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to load '%s': %v", resourceType, err)
	}

	return nil
}

func (h *handler) serveJobResultsPlot(jobId string, w http.ResponseWriter) error {

	dataTable, err := reports.CreateDataTable(h.client, jobId)
	if err != nil {
		return err
	}

	cfg := reports.ReportConfig{
		Title: fmt.Sprintf("Results report for job %s", jobId),
	}

	cfg.AddSections(
		reports.JobsTable(),
	)

	if dataTable.HasSpeed() {
		cfg.AddSections(
			reports.HorizontalBoxChart("", reports.Speed, ""),
			reports.ResultsTable(reports.Speed, "", true),
		)
	}

	cfg.AddSections(
		reports.HorizontalBoxChart("", reports.TimeOp, ""),
		reports.ResultsTable(reports.TimeOp, "", true),
	)

	err = reports.WriteReport(&cfg, dataTable, w)
	if err != nil {
		return err
	}
	return nil
}
