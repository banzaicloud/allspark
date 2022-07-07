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
	reader *kafka.Reader
	dialer *kafka.Dialer

	logger log.Logger

	BootstrapServer string `mapstructure:"bootstrap_server"`
	Topic           string `mapstructure:"topic"`
	ConsumerGroup   string `mapstructure:"consumer_group"`
}

func NewConsumer(bootStrapServer string, topic string, consumerGroup string, logger log.Logger) *Consumer {
	dialer := kafka.DefaultDialer

	progName := filepath.Base(os.Args[0])
	hostName, _ := os.Hostname()
	dialer.ClientID = fmt.Sprintf("%s@%s", progName, hostName)

	consumer := &Consumer{
		BootstrapServer: bootStrapServer,
		Topic:           topic,
		ConsumerGroup:   consumerGroup,
		logger:          logger,
		dialer:          dialer,
	}

	consumer, _ = consumer.Validate()

	return consumer
}

func (c *Consumer) Consume(ctx context.Context) (*kafka.Message, error) {
	if c.reader == nil {
		c.reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{c.BootstrapServer},
			Topic:   c.Topic,
			GroupID: c.ConsumerGroup,
			Logger:  c.logger,
			Dialer:  c.dialer,
		})
	}

	// the `ReadMessage` method blocks until we receive the next event
	message, err := c.reader.ReadMessage(ctx)
	if err != nil {
		if err := c.reader.Close(); err != nil {
			return nil, errors.WrapIf(err, "failed to close kafka reader")
		}
		return nil, errors.WrapIf(err, "could not read kafka message")
	}

	return &message, nil
}

func (c *Consumer) SetLogger(log log.Logger) {
	c.logger = log
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Validate() (*Consumer, error) {
	if c.BootstrapServer == "" {
		c.BootstrapServer = "kafka-all-broker.kafka:29092"
	}

	if c.ConsumerGroup == "" {
		c.ConsumerGroup = "allspark-consumer-group"
	}

	if c.Topic == "" {
		c.Topic = "example-topic"
	}

	return c, nil
}
