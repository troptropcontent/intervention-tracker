package jobs

// BackgroundJobRunner defines the interface for running background jobs
// This allows us to swap between async (production) and sync (testing) implementations
type BackgroundJobRunner interface {
	Perform(job func())
}

// AsyncJobRunner executes jobs asynchronously using goroutines (for production)
type AsyncJobRunner struct{}

func NewAsyncJobRunner() *AsyncJobRunner {
	return &AsyncJobRunner{}
}

func (r *AsyncJobRunner) Perform(job func()) {
	go job()
}

// SyncJobRunner executes jobs synchronously (for testing)
type SyncJobRunner struct{}

func NewSyncJobRunner() *SyncJobRunner {
	return &SyncJobRunner{}
}

func (r *SyncJobRunner) Perform(job func()) {
	// Execute synchronously - no goroutine
	job()
}
