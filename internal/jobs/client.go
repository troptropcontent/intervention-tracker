package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	attachementsJobs "github.com/troptropcontent/qr_code_maintenance/internal/jobs/attachements"
	interventionsJobs "github.com/troptropcontent/qr_code_maintenance/internal/jobs/interventions"
	"github.com/troptropcontent/qr_code_maintenance/internal/jobs/types"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/pdf"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

// Config holds all shared dependencies for job processing
type Config struct {
	DB             *gorm.DB
	DBPool         *pgxpool.Pool
	StorageService storage.StorageService
	EmailService   email.EmailService
	PdfConverter   pdf.HtmlToPdfConverter
}

// NewConfig creates a new Config with all dependencies initialized
func NewConfig(ctx context.Context) (*Config, error) {
	// Initialize GORM database connection
	db, err := database.InitializeDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create pgx connection pool with limits
	dbPool, err := createDBPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	// Initialize S3 storage service
	storageService, err := storage.NewS3StorageService(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage service: %w", err)
	}

	// Initialize email service
	emailService, err := email.NewSMTPServiceFromEnv(&email.NewSMTPServiceFromEnvOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize email service: %w", err)
	}

	return &Config{
		DB:             db,
		DBPool:         dbPool,
		StorageService: storageService,
		EmailService:   emailService,
		PdfConverter:   pdf.NewGotenbergService(),
	}, nil
}

// Close closes all resources held by the config
func (c *Config) Close() {
	if c.DBPool != nil {
		c.DBPool.Close()
	}
}

// createDBPool creates a pgx connection pool with appropriate limits
func createDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := database.GetDatabaseDSN()

	// Parse the DSN into a config so we can set connection pool limits
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	// Set connection pool limits to prevent exhausting database connections
	config.MaxConns = 10                      // Maximum number of connections in the pool
	config.MinConns = 2                       // Minimum number of connections to keep warm
	config.MaxConnLifetime = time.Hour        // Maximum lifetime of a connection
	config.MaxConnIdleTime = 30 * time.Minute // Maximum idle time before closing a connection
	config.HealthCheckPeriod = time.Minute    // How often to check connection health

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// NewWorkers creates River workers with the given dependencies
// This is a pure function that doesn't create any resources
func NewWorkers(cfg *Config, jobEnqueuer types.BackgroundJobRunner) *river.Workers {
	workers := river.NewWorkers()

	river.AddWorker(workers, &attachementsJobs.UploadWorker{
		DB:             cfg.DB,
		StorageService: cfg.StorageService,
	})

	river.AddWorker(workers, &interventionsJobs.AttachReportPdfWorker{
		DB:                 cfg.DB,
		StorageService:     cfg.StorageService,
		EmailService:       cfg.EmailService,
		JobRunner:          jobEnqueuer,
		HtmlToPdfConverter: cfg.PdfConverter,
	})

	return workers
}

// RiverEnqueuer is used to enqueue jobs (typically in the web application)
type RiverEnqueuer struct {
	client *river.Client[pgx.Tx]
	pool   *pgxpool.Pool
}

// NewEnqueuerFromConfig creates a new job enqueuer using a pre-configured Config
// This is used by the background job runner where Config manages all dependencies
func NewEnqueuerFromConfig(ctx context.Context, cfg *Config) (*RiverEnqueuer, error) {
	// The enqueuer doesn't need workers configured - it just inserts into the queue
	// Workers are only needed by the runner that processes jobs
	client, err := river.NewClient(riverpgxv5.New(cfg.DBPool), &river.Config{
		// No workers needed for enqueueing
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create river client: %w", err)
	}

	return &RiverEnqueuer{
		client: client,
		pool:   cfg.DBPool,
	}, nil
}

// NewEnqueuer creates a standalone job enqueuer with its own connection pool
// This is used by the web application where dependencies are managed separately
func NewEnqueuer(ctx context.Context) (*RiverEnqueuer, error) {
	// Create just the pgx pool for job enqueueing
	dbPool, err := createDBPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	// The enqueuer doesn't need workers configured - it just inserts into the queue
	client, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		// No workers needed for enqueueing
	})
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("failed to create river client: %w", err)
	}

	return &RiverEnqueuer{
		client: client,
		pool:   dbPool,
	}, nil
}

// Enqueue adds a job to the queue
func (e *RiverEnqueuer) Enqueue(ctx context.Context, args river.JobArgs) error {
	tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = e.client.InsertTx(ctx, tx, args, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Close closes the enqueuer's resources
func (e *RiverEnqueuer) Close() error {
	if e.pool != nil {
		e.pool.Close()
	}
	return nil
}

// RiverRunner runs background jobs (used in the background job process)
type RiverRunner struct {
	client   *river.Client[pgx.Tx]
	enqueuer *RiverEnqueuer
}

// NewRunner creates a new job runner for processing background jobs
func NewRunner(ctx context.Context, cfg *Config) (*RiverRunner, error) {
	// Create a minimal enqueuer for workers to use when they need to enqueue other jobs
	// This enqueuer shares the same DBPool as the runner
	enqueuer := &RiverEnqueuer{
		pool: cfg.DBPool,
		// client will be set after we create the runner's client
	}

	// Create workers with the enqueuer reference
	workers := NewWorkers(cfg, enqueuer)

	// Create the river client that will RUN the workers
	client, err := river.NewClient(riverpgxv5.New(cfg.DBPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create river client: %w", err)
	}

	// Now complete the enqueuer by setting its client
	// This allows workers to enqueue other jobs (e.g., AttachReportPdfWorker can enqueue attachment jobs)
	enqueuer.client = client

	return &RiverRunner{
		client:   client,
		enqueuer: enqueuer,
	}, nil
}

// Start starts the job runner
func (r *RiverRunner) Start(ctx context.Context) error {
	return r.client.Start(ctx)
}

// Stop stops the job runner gracefully
func (r *RiverRunner) Stop(ctx context.Context) error {
	return r.client.Stop(ctx)
}
