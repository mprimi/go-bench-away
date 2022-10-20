package core

// Namespace to allow multiple sets of jobs on the same instance
// TODO: allow setting this, currently hardcoded
const Namespace = "default"

// TODO: rename
const jobRecordsStoreName = Namespace + "-jobs"

// Name of stream where jobs are submitted
const jobStreamName = Namespace + "-jobs"

// Name of object store where jobs artifacts are stored
const artifactsStoreName = Namespace + "-artifacts"

// Name of consumer used by worker
const consumerName = Namespace + "-worker"

// Subject used when submitting job
const jobSubmitSubjectTemplate = Namespace + "jobs.submit.%s"

// KV key containing job record
const jobRecordKeyTemplate = "jobs/%s"

// Header name containing job id when submitting
const jobIdHeader = "x-job-id"

// Name of logfile
const logFileName = "log.txt"

// Key of log file artifact
const logKeyTemplate = "jobs/%s/log.txt"

// Name of benchmarks results file
const resultsFileName = "results.txt"

// Key of results file artifact
const resultsKeyTemplate = "jobs/%s/results.txt"
