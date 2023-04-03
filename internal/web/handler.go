package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/mprimi/go-bench-away/pkg/core"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

//go:embed html/index.html.tmpl
var indexTmpl string

//go:embed html/queue.html.tmpl
var queueTmpl string

var jobResourceRegexp = regexp.MustCompile(`^/job/([[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12})/(log|script|results|record)/?$`)

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
		Jobs []*core.JobRecord
	}{
		Jobs: jobRecords,
	}
	return h.queueTemplate.Execute(w, tv)
}

func (h *handler) serveJobResource(w http.ResponseWriter, jobId, resourceType string) error {

	jobRecord, _, err := h.client.LoadJob(jobId)
	if err != nil {
		return fmt.Errorf("Failed to load job '%s': %v", jobId, err)
	}

	var content []byte

	switch resourceType {
	case "log":
		content, err = h.client.LoadLogArtifact(jobRecord)
	case "script":
		content, err = h.client.LoadScriptArtifact(jobRecord)
	case "results":
		content, err = h.client.LoadResultsArtifact(jobRecord)
	case "record":
		content, err = json.MarshalIndent(jobRecord, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to load '%s': %v", resourceType, err)
	}

	_, err = w.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write response: %v", err)
	}

	return nil
}
