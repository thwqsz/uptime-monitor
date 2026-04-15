package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/thwqsz/uptime-monitor/internal/contracts"
	"go.uber.org/zap"
)

type CheckResultProcessor interface {
	ProcessCheckResult(ctx context.Context, resCheck *contracts.CheckResult) error
}

type Consumer struct {
	log           *zap.Logger
	reader        *kafka.Reader
	resultService CheckResultProcessor
}

func NewConsumer(log *zap.Logger, reader *kafka.Reader, srv CheckResultProcessor) *Consumer {
	return &Consumer{log: log, reader: reader, resultService: srv}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.log.Error("error during reading", zap.Error(err))
			continue
		}
		var taskResult contracts.CheckResult
		err = json.Unmarshal(msg.Value, &taskResult)
		if err != nil {
			c.log.Error("error during making struct from json", zap.Error(err))
			continue
		}
		err = c.resultService.ProcessCheckResult(ctx, &taskResult)
		if err != nil {
			c.log.Error("error during saving check_result", zap.Error(err))
			continue
		}
	}

}
