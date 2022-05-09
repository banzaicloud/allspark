package kafka

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/banzaicloud/allspark/internal/platform/log"
	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	bootstrapServer string
	topic           string
	consumerGroup   string

	logger log.Logger
}

func NewConsumer(bootStrapServer string, topic string, consumerGroup string, logger log.Logger) *Consumer {
	logger = logger.WithField("server", "consumer")

	if consumerGroup == "" {
		consumerGroup = "allspark-consumer-group"
	}

	return &Consumer{
		bootstrapServer: bootStrapServer,
		topic:           topic,
		consumerGroup:   consumerGroup,
		logger:          logger,
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	dialer := kafka.DefaultDialer

	progName := filepath.Base(os.Args[0])
	hostName, _ := os.Hostname()
	dialer.ClientID = fmt.Sprintf("%s@%s", progName, hostName)

	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{c.bootstrapServer},
		Topic:   c.topic,
		GroupID: c.consumerGroup,
		Logger:  c.logger,
		Dialer:  dialer,
	})
	for {
		// the `ReadMessage` method blocks until we receive the next event
		message, err := r.ReadMessage(ctx)
		if err != nil {
			if err := r.Close(); err != nil {
				c.logger.Error(errors.WrapIf(err, "failed to close kafka reader"))
			}
			panic(errors.WrapIf(err, "could not read kafka message"))
		}

		c.logger.Infof("message at offset %d: %s = %s\n\n", message.Offset, string(message.Key), string(message.Value))
	}

}
