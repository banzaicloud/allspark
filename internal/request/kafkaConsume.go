package request

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"emperror.dev/errors"
	"github.com/google/uuid"

	"github.com/banzaicloud/allspark/internal/kafka"
	"github.com/banzaicloud/allspark/internal/platform/log"
)

type kafkaConsumeFactory struct {
	logger log.Logger
}

func NewKafkaConsumeFactory(logger log.Logger) Factory {
	f := &kafkaConsumeFactory{
		logger: logger,
	}

	return f
}

func (f *kafkaConsumeFactory) CreateRequest(u *url.URL) (Request, error) {
	pieces := strings.Split(u.RawQuery, "=")
	if len(pieces) != 2 {
		return nil, errors.New("invalid kafka consume url; provide only the consumer group after the '?'")
	}

	bootstrapServer := u.Host
	topic := strings.Trim(u.Path, "/")
	consumerGroup := pieces[1]

	consumer := kafka.NewConsumer(bootstrapServer, topic, consumerGroup, f.logger)

	return KafkaConsumeRequest{
		BootstrapServer: bootstrapServer,
		Topic:           topic,
		ConsumerGroup:   consumerGroup,

		consumer: consumer,
		count:    parseCountFromURL(u),
	}, nil
}

type KafkaConsumeRequest struct {
	BootstrapServer string `json:"host"`
	Topic           string `json:"topic"`
	ConsumerGroup   string `json:"consumerGroup"`

	consumer *kafka.Consumer
	count    uint
}

func (request KafkaConsumeRequest) Count() uint {
	return request.count
}

func (request KafkaConsumeRequest) SetConsumer(consumer *kafka.Consumer) {
	request.consumer = consumer
}

func (request KafkaConsumeRequest) Do(incomingRequestHeaders http.Header, logger log.Logger) {
	correlationID := uuid.New()
	loggerWithFields := logger.WithFields(log.Fields{
		"correlationID":   correlationID,
		"bootstrapServer": request.BootstrapServer,
		"topic":           request.Topic,
		"consumerGroup":   request.ConsumerGroup,
	})

	request.consumer.SetLogger(loggerWithFields)

	message, err := request.consumer.Consume(context.Background())
	if err != nil {
		loggerWithFields.Error(err.Error())
		return
	}

	loggerWithFields.WithField("message", message).Info("message received")
}
