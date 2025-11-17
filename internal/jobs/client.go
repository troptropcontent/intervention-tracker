package jobs

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	attachementsJobs "github.com/troptropcontent/qr_code_maintenance/internal/jobs/attachements"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
)

func Workers() *river.Workers {
	db, err := database.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize S3 storage service (reads config from environment variables)
	storageService, err := storage.NewS3StorageService(nil)
	if err != nil {
		log.Fatalf("failed to initialize storage service: %v", err)
	}
	workers := river.NewWorkers()

	river.AddWorker(workers, &attachementsJobs.UploadWorker{
		DB:             db,
		StorageService: storageService,
	})

	return workers
}

type RiverEnqueuer struct {
	RiverClient *river.Client[pgx.Tx]
	DbPool      *pgxpool.Pool
}
type RiverRunner struct {
	RiverClient *river.Client[pgx.Tx]
}

func GetDbPool(ctx context.Context) *pgxpool.Pool {
	dsn := database.GetDatabaseDSN()
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to initialize dbPool: %v", err)
	}
	return dbPool
}

func NewEnqueuer() *RiverEnqueuer {
	ctx := context.Background()
	dbPool := GetDbPool(ctx)
	workers := Workers()
	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Workers: workers,
	})
	if err != nil {
		log.Fatalf("Failed to initialize to riverClient: %v", err)
	}
	return &RiverEnqueuer{
		RiverClient: riverClient,
		DbPool:      dbPool,
	}
}

func (e *RiverEnqueuer) Enqueue(ctx context.Context, args river.JobArgs) error {
	tx, err := e.DbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = e.RiverClient.InsertTx(ctx, tx, args, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Close closes the database pool and cleans up resources
func (e *RiverEnqueuer) Close() {
	if e.DbPool != nil {
		e.DbPool.Close()
	}
}

func NewRunner(ctx context.Context) *RiverRunner {
	dbPool := GetDbPool(ctx)
	workers := Workers()
	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		log.Fatalf("Failed to initialize to riverClient: %v", err)
	}
	return &RiverRunner{
		RiverClient: riverClient,
	}
}

func (r *RiverRunner) Start(ctx context.Context) {
	if err := r.RiverClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start to riverClient: %v", err)
	}
}

func (r *RiverRunner) Stop(ctx context.Context) {
	if err := r.RiverClient.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop to riverClient: %v", err)
	}
}
