package worker

import (
	"context"
	_ "embed"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
)

const (
	kScriptFilename    = "run.sh"
	kLogFilename       = "log.txt"
	kResultsFilename   = "results.txt"
	kShaFilename       = "sha.txt"
	kGoversionFilename = "go_version.txt"
)

//go:embed scripts/benchmark.sh.tmpl
var runScriptTmpl string

type Worker interface {
	Run(context.Context) error
}

type workerImpl struct {
	c              client.Client
	jobsDir        string
	workerInfo     core.WorkerInfo
	scriptTemplate *template.Template
}

func NewWorker(c client.Client, jobsDir string) (Worker, error) {
	// Utsname byte arrays are filled with string termination characters,
	// and naive string conversion preserves them.
	bts := func(buf []byte) string {
		stringWithNullChars := string(buf[:])
		trimmedString := strings.ReplaceAll(stringWithNullChars, "\x00", "")
		utf8String := strings.ToValidUTF8(trimmedString, "")
		return strings.TrimSpace(utf8String)
	}

	buf := unix.Utsname{}
	if err := unix.Uname(&buf); err != nil {
		return nil, err
	}

	return &workerImpl{
		c:       c,
		jobsDir: jobsDir,
		workerInfo: core.WorkerInfo{
			Hostname: bts(buf.Nodename[:]),
			Uname:    fmt.Sprintf("%s_%s-%s", bts(buf.Sysname[:]), bts(buf.Release[:]), bts(buf.Machine[:])),
			Version:  fmt.Sprintf("%s (%s)", core.Version, core.SHA),
		},
		scriptTemplate: template.Must(template.New("benchmark_script").Parse(runScriptTmpl)),
	}, nil
}

func (w *workerImpl) Run(ctx context.Context) error {

	handleJob := func(jr *core.JobRecord, revision uint64) (bool, error) {
		return w.processJob(jr, revision)
	}
	fmt.Printf("⚙️  Ready for work\n")
	return w.c.DispatchJobs(ctx, handleJob)
}

func (w *workerImpl) processJob(job *core.JobRecord, revision uint64) (bool, error) {

	if job.Status != core.Submitted {
		return false, fmt.Errorf("Cannot process job %s in status %v", job.Id, job.Status)
	}

	job.Status = core.Running
	job.Started = time.Now()
	job.WorkerInfo = w.workerInfo

	newRevision, err := w.c.UpdateJob(job, revision)
	if err != nil {
		// TODO: retry if error is transitional
		return false, fmt.Errorf("Failed to update job %s: %v", job.Id, err)
	}

	fmt.Printf("⚙️  Processing job %s\n", job.Id)

	jobTempDir, runErr := w.runJob(job)

	// Update job status to final
	job.Completed = time.Now()
	job.Status = core.Succeeded
	if runErr != nil {
		job.Status = core.Failed
	}

	// Upload artifacts
	uploadErr := w.uploadArtifacts(job, jobTempDir)
	if uploadErr != nil {
		fmt.Fprintf(os.Stderr, "Job %s artifacts upload failed: %v\n", job.Id, uploadErr)
		job.Status = core.Failed
	}

	// Remove job directory
	if jobTempDir != "" && !job.Parameters.SkipCleanup {
		defer os.RemoveAll(jobTempDir)
	}

	fmt.Printf("⚙️  Completed job %s, updating status to: %s\n", job.Id, job.Status)
	_, finalUpdateErr := w.c.UpdateJob(job, newRevision)
	if finalUpdateErr != nil {
		// TODO: retry if error is transitional
		return false, fmt.Errorf("Failed to update job %s: %v", job.Id, finalUpdateErr)
	}

	return false, nil
}

func (w *workerImpl) runJob(job *core.JobRecord) (string, error) {

	jobTempDir, err := os.MkdirTemp(w.jobsDir, fmt.Sprintf("go-bench-away-job-%s-", job.Id))
	if err != nil {
		return "", fmt.Errorf("Failed to create job directory: %v", err)
	}

	scriptPath := filepath.Join(jobTempDir, kScriptFilename)
	logPath := filepath.Join(jobTempDir, kLogFilename)
	resultsPath := filepath.Join(jobTempDir, kResultsFilename)
	shaPath := filepath.Join(jobTempDir, kShaFilename)
	goVersionPath := filepath.Join(jobTempDir, kGoversionFilename)

	scriptFile, err := os.Create(scriptPath)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to create script: %v", err)
	}

	scriptTemplateValues := struct {
		JobDirPath      string
		ResultsPath     string
		ShaPath         string
		GoVersionPath   string
		GitRemote       string
		GitRef          string
		TestsSubDir     string
		TestsFilterExpr string
		Reps            string
		MinRuntime      string
		Timeout         string
		GoPath          string
	}{
		JobDirPath:      jobTempDir,
		ResultsPath:     resultsPath,
		ShaPath:         shaPath,
		GoVersionPath:   goVersionPath,
		GitRemote:       job.Parameters.GitRemote,
		GitRef:          job.Parameters.GitRef,
		TestsSubDir:     job.Parameters.TestsSubDir,
		TestsFilterExpr: job.Parameters.TestsFilterExpr,
		Reps:            fmt.Sprintf("%d", job.Parameters.Reps),
		MinRuntime:      fmt.Sprintf("%v", job.Parameters.TestMinRuntime),
		Timeout:         fmt.Sprintf("%v", job.Parameters.Timeout),
		GoPath:          job.Parameters.GoPath,
	}

	err = w.scriptTemplate.Execute(scriptFile, scriptTemplateValues)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to write job script: %v", err)
	}
	scriptFile.Close()

	err = os.Chmod(scriptPath, 0700)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to make script executable: %v", err)
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to create logfile: %v", err)
	}
	defer logFile.Close()

	// Tee output to logfile and worker stdout
	mw := io.MultiWriter(logFile, os.Stdout)

	cmd := exec.CommandContext(context.Background(), scriptPath)

	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Start()
	if err != nil {
		return jobTempDir, fmt.Errorf("Failed to launch job %s: %w", job.Id, err)
	}

	procState, waitErr := cmd.Process.Wait()
	if waitErr != nil {
		return jobTempDir, fmt.Errorf("Error waiting for termination of job %s: %s", job.Id, waitErr)
	}

	shaBytes, err := os.ReadFile(shaPath)
	if err == nil {
		job.SHA = strings.TrimSpace(string(shaBytes))
	} else {
		job.SHA = "?"
	}

	goVersionBytes, err := os.ReadFile(goVersionPath)
	if err == nil {
		job.GoVersion = strings.TrimSpace(string(goVersionBytes))
	} else {
		job.GoVersion = "?"
	}

	if procState.ExitCode() != 0 {
		return jobTempDir, fmt.Errorf("Non-zero exit code")
	}

	return jobTempDir, nil
}

func (w *workerImpl) uploadArtifacts(job *core.JobRecord, jobDirPath string) error {

	logPath := filepath.Join(jobDirPath, kLogFilename)
	logArtifactKey, logErr := w.c.UploadLogArtifact(job.Id, logPath)
	if logErr != nil {
		fmt.Printf("Log artifact upload error: %v\n", logErr)
	} else {
		job.Log = logArtifactKey
	}

	resultsPath := filepath.Join(jobDirPath, kResultsFilename)
	resultsArtifactKey, resultsErr := w.c.UploadResultsArtifact(job.Id, resultsPath)
	if resultsErr != nil {
		fmt.Printf("Results artifact upload error: %v\n", resultsErr)
	} else {
		job.Results = resultsArtifactKey
	}

	scriptPath := filepath.Join(jobDirPath, kScriptFilename)
	scriptArtifactKey, scriptErr := w.c.UploadScriptArtifact(job.Id, scriptPath)
	if scriptErr != nil {
		fmt.Printf("Script artifact upload error: %v\n", scriptErr)
	} else {
		job.Script = scriptArtifactKey
	}

	if logErr != nil || resultsErr != nil || scriptErr != nil {
		return fmt.Errorf("Artifacts upload error")
	}

	return nil
}
