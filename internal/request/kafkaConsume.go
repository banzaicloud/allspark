package request

import (
	"context"
	"net/http"

	"github.com/banzaicloud/allspark/internal/kafka"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/google/uuid"
)

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
