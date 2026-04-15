package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/thwqsz/uptime-monitor/internal/contracts"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	log    *zap.Logger
}

func NewProducer(writer *kafka.Writer, log *zap.Logger) *Producer {
	return &Producer{writer: writer, log: log}
}

func (p *Producer) SendTask(ctx context.Context, target *models.Target) error {
	uniqID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	uniqIDStr := uniqID.String()
	now := time.Now().UTC().Format(time.RFC3339)
	taskForSent := &contracts.CheckTask{
		TaskID:      uniqIDStr,
		TargetID:    target.ID,
		URL:         target.URL,
		TimeoutSec:  target.Timeout,
		ScheduledAt: now,
	}
	taskBytes, err := json.Marshal(taskForSent)
	if err != nil {
		p.log.Error("error during making json from target", zap.Error(err))
		return err
	}
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("check_task"),
		Value: taskBytes,
	}); err != nil {
		return err
	}

	return nil
}
