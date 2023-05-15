package kafka

import (
	"context"

	"emperror.dev/errors"
	kafka "github.com/segmentio/kafka-go"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

type Producer struct {
	writer *kafka.Writer

	bootstrapServer string
	topic           string

	logger log.Logger
}

func NewProducer(bootStrapServer string, topic string, logger log.Logger) *Producer {
	return &Producer{
		bootstrapServer: bootStrapServer,
		topic:           topic,
		logger:          logger,
	}
}

func (p *Producer) Produce(ctx context.Context, message string) error {
	if p.writer == nil {
		p.writer = &kafka.Writer{
			Topic:  p.topic,
			Addr:   kafka.TCP(p.bootstrapServer),
			Logger: p.logger,
		}
	}

	err := p.writer.WriteMessages(ctx, kafka.Message{
		Value: []byte(message),
	})
	if err != nil {
		return errors.WrapIf(err, "could not write kafka message")
	}

	return nil
}

func (p *Producer) SetLogger(log log.Logger) {
	p.logger = log
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
