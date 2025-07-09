package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NATSClient interface {
	SubscribeVerificationCreate(ctx context.Context, handler func(VerificationCreateMessage)) error
	PublishVerificationCompleted(ctx context.Context, verificationID string, status string, error string) error
	Close()
}

type natsClient struct {
	conn   *nats.Conn
	logger *zap.Logger
}

type VerificationCreateMessage struct {
	VerificationID string   `json:"verification_id"`
	INN            string   `json:"inn"`
	RequestedTypes []string `json:"requested_types"`
	AuthorEmail    string   `json:"author_email"`
}

type VerificationCompletedMessage struct {
	VerificationID string `json:"verification_id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
}

func NewNATSClient(url string, logger *zap.Logger) (NATSClient, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	logger.Info("connected to NATS", zap.String("url", url))
	return &natsClient{conn: conn, logger: logger}, nil
}

func (c *natsClient) SubscribeVerificationCreate(ctx context.Context, handler func(VerificationCreateMessage)) error {
	_, err := c.conn.Subscribe("verification.create", func(msg *nats.Msg) {
		var m VerificationCreateMessage
		if err := json.Unmarshal(msg.Data, &m); err != nil {
			c.logger.Error("failed to unmarshal verification.create", zap.Error(err))
			return
		}
		handler(m)
	})
	if err != nil {
		c.logger.Error("failed to subscribe to verification.create", zap.Error(err))
		return err
	}
	c.logger.Info("subscribed to verification.create")
	return nil
}

func (c *natsClient) PublishVerificationCompleted(ctx context.Context, verificationID string, status string, error string) error {
	msg := VerificationCompletedMessage{
		VerificationID: verificationID,
		Status:         status,
		Error:          error,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		c.logger.Error("failed to marshal verification completed message", zap.Error(err))
		return fmt.Errorf("failed to marshal verification completed message: %w", err)
	}

	err = c.conn.Publish("verification.completed", data)
	if err != nil {
		c.logger.Error("failed to publish verification completed", zap.Error(err), zap.String("verification_id", verificationID))
		return fmt.Errorf("failed to publish verification completed: %w", err)
	}

	c.logger.Info("verification completed message published", zap.String("verification_id", verificationID), zap.String("status", status))
	return nil
}

func (c *natsClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.logger.Info("NATS connection closed")
	}
}
