package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"scoring_worker/internal/config"
	"scoring_worker/internal/credinform"
	"scoring_worker/internal/logger"
	"scoring_worker/internal/messaging"
	"scoring_worker/internal/repository"
	"scoring_worker/internal/service"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	cfg, err := setupConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to setup config: %v", err))
	}

	log, err := setupLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to setup logger: %v", err))
	}
	defer log.Sync()

	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("failed to setup database", zap.Error(err))
	}
	defer db.Close()

	natsClient, err := setupNATSClient(cfg, log)
	if err != nil {
		log.Fatal("failed to setup nats client", zap.Error(err))
	}
	defer natsClient.Close()

	cacheRepo := repository.NewDataCacheRepository(db, log)
	repo := repository.NewVerificationRepository(db, cacheRepo, log)
	credinformClient := credinform.NewClient(&cfg.Credinform, log)
	verificationService := service.NewVerificationService(credinformClient, repo, log)

	worker := NewWorker(log, repo, verificationService, natsClient, cfg.WorkerConcurrency)
	worker.Run()
}

func setupConfig() (*config.Config, error) {
	return config.Load()
}

func setupLogger(cfg *config.Config) (*zap.Logger, error) {
	return logger.New(cfg.Log.Level, cfg.Log.JSON)
}

func setupDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), cfg.DatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func setupNATSClient(cfg *config.Config, log *zap.Logger) (messaging.NATSClient, error) {
	return messaging.NewNATSClient(cfg.NATS.URL, log)
}

type Worker struct {
	log                 *zap.Logger
	repo                repository.VerificationRepository
	verificationService service.VerificationService
	natsClient          messaging.NATSClient
	concurrencyCh       chan struct{}
}

func NewWorker(log *zap.Logger, repo repository.VerificationRepository, verificationService service.VerificationService, natsClient messaging.NATSClient, concurrency int) *Worker {
	if concurrency <= 0 {
		concurrency = 5
	}
	return &Worker{
		log:                 log,
		repo:                repo,
		verificationService: verificationService,
		natsClient:          natsClient,
		concurrencyCh:       make(chan struct{}, concurrency),
	}
}

func (w *Worker) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.resumeProcessing(ctx)

	err := w.subscribeToVerifications(ctx)
	if err != nil {
		w.log.Fatal("Failed to subscribe to NATS", zap.Error(err))
	}

	w.log.Info("Worker started, waiting for messages...")
	w.waitForShutdownSignal()
}

func (w *Worker) resumeProcessing(ctx context.Context) {
	w.log.Info("Checking for stale verifications to resume...")
	for {
		verification, err := w.repo.AcquireStaleVerification(ctx)
		if err != nil {
			w.log.Error("Failed to acquire stale verification", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		if verification == nil {
			w.log.Info("No more stale verifications to resume.")
			break
		}

		go func(v repository.Verification) {
			w.concurrencyCh <- struct{}{}
			defer func() { <-w.concurrencyCh }()

			w.log.Info("Resuming verification processing", zap.String("id", v.ID), zap.String("inn", v.Inn))
			if err := w.verificationService.ProcessVerification(context.Background(), v.ID, v.Inn, v.RequestedDataTypes); err != nil {
				w.log.Error("Failed to process resumed verification", zap.Error(err), zap.String("id", v.ID))
				w.natsClient.PublishVerificationCompleted(ctx, v.ID, "ERROR", err.Error())
				return
			}
			w.log.Info("Resumed verification processing completed", zap.String("id", v.ID))
			err = w.natsClient.PublishVerificationCompleted(ctx, v.ID, "COMPLETED", "")
			if err != nil {
				w.log.Error("Failed to publish completion notification for resumed verification", zap.Error(err), zap.String("id", v.ID))
			}
		}(*verification)
	}
}

func (w *Worker) subscribeToVerifications(ctx context.Context) error {
	return w.natsClient.SubscribeVerificationCreate(ctx, func(msg messaging.VerificationCreateMessage) {
		w.log.Info("Received verification.create", zap.String("id", msg.VerificationID), zap.String("inn", msg.INN))

		err := w.repo.Create(ctx, msg.VerificationID, msg.INN, msg.RequestedTypes, msg.AuthorEmail)
		if err != nil {
			w.log.Error("Failed to create verification in DB", zap.Error(err), zap.String("id", msg.VerificationID))
			w.natsClient.PublishVerificationCompleted(ctx, msg.VerificationID, "ERROR", err.Error())
			return
		}

		go func(id string, inn string, requestedTypes []string) {
			w.concurrencyCh <- struct{}{}
			defer func() { <-w.concurrencyCh }()

			w.log.Info("Starting verification processing", zap.String("id", id), zap.String("inn", inn))

			if err := w.verificationService.ProcessVerification(context.Background(), id, inn, requestedTypes); err != nil {
				w.log.Error("Failed to process verification", zap.Error(err), zap.String("id", id))
				w.natsClient.PublishVerificationCompleted(ctx, id, "ERROR", err.Error())
				return
			}

			w.log.Info("Verification processing completed", zap.String("id", id))

			err = w.natsClient.PublishVerificationCompleted(ctx, id, "COMPLETED", "")
			if err != nil {
				w.log.Error("Failed to publish completion notification", zap.Error(err), zap.String("id", id))
			}
		}(msg.VerificationID, msg.INN, msg.RequestedTypes)
	})
}

func (w *Worker) waitForShutdownSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	w.log.Info("Shutting down worker")
}
