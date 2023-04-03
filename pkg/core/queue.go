package core

type QueueStatus struct {
	SubmittedCount uint64
	RunningJob     *JobRecord
}
