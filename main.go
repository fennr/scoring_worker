package main

import (
	"context"
	"os"
	"os/signal"
	"scoring_worker/internal/config"
	"scoring_worker/internal/credinform"
	"scoring_worker/internal/logger"
	"scoring_worker/internal/messaging"
	"scoring_worker/internal/repository"
	"scoring_worker/internal/service"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Log.Level, cfg.Log.JSON)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseDSN())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	repo := repository.NewVerificationRepository(db, log)

	credinformClient := credinform.NewClient(&cfg.Credinform, log)
	companyService := service.NewCompanyService(credinformClient, repo, log)

	natsClient, err := messaging.NewNATSClient(cfg.NATS.URL, log)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = natsClient.SubscribeVerificationCreate(ctx, func(msg messaging.VerificationCreateMessage) {
		log.Info("Received verification.create", zap.String("id", msg.VerificationID), zap.String("inn", msg.INN))

		err := repo.Create(ctx, msg.VerificationID, msg.INN, msg.RequestedTypes, msg.AuthorEmail)
		if err != nil {
			log.Error("Failed to create verification in DB", zap.Error(err), zap.String("id", msg.VerificationID))
			natsClient.PublishVerificationCompleted(ctx, msg.VerificationID, "ERROR", err.Error())
			return
		}

		go func(id string, inn string, requestedTypes []string) {
			log.Info("Starting verification processing", zap.String("id", id), zap.String("inn", inn))

			if err := companyService.ProcessVerification(context.Background(), id, inn, requestedTypes); err != nil {
				log.Error("Failed to process verification", zap.Error(err), zap.String("id", id))
				natsClient.PublishVerificationCompleted(ctx, id, "ERROR", err.Error())
				return
			}

			log.Info("Verification processing completed", zap.String("id", id))

			err = natsClient.PublishVerificationCompleted(ctx, id, "COMPLETED", "")
			if err != nil {
				log.Error("Failed to publish completion notification", zap.Error(err), zap.String("id", id))
			}
		}(msg.VerificationID, msg.INN, msg.RequestedTypes)
	})
	if err != nil {
		log.Fatal("Failed to subscribe to NATS", zap.Error(err))
	}

	log.Info("Worker started, waiting for messages...")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Info("Shutting down worker")
}
