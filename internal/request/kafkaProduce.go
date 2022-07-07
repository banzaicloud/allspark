package request

import (
	"context"
	"net/http"

	"github.com/banzaicloud/allspark/internal/kafka"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/google/uuid"
)

type KafkaProduceRequest struct {
	BootstrapServer string `json:"host"`
	Topic           string `json:"topic"`
	Message         string `json:"message"`

	producer *kafka.Producer
	count    uint
}

func (request KafkaProduceRequest) Count() uint {
	return request.count
}

func (request KafkaProduceRequest) Do(incomingRequestHeaders http.Header, logger log.Logger) {
	correlationID := uuid.New()
	loggerWithFields := logger.WithFields(log.Fields{
		"correlationID":   correlationID,
		"bootstrapServer": request.BootstrapServer,
		"topic":           request.Topic,
	})

	request.producer.SetLogger(loggerWithFields)

	err := request.producer.Produce(context.Background(), request.Message)
	if err != nil {
		loggerWithFields.Error(err.Error())
		return
	}

	loggerWithFields.WithField("message", request.Message).Info("message sent")
}
