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

type kafkaProduceFactory struct {
	logger log.Logger
}

func NewKafkaProduceFactory(logger log.Logger) Factory {
	f := &kafkaProduceFactory{
		logger: logger,
	}

	return f
}

func (f *kafkaProduceFactory) CreateRequest(u *url.URL) (Request, error) {
	pieces := strings.Split(u.RawQuery, "=")
	if len(pieces) != 2 {
		return nil, errors.New("invalid kafka produce url; provide only the message after the '?'")
	}

	bootstrapServer := u.Host
	topic := strings.Trim(u.Path, "/")
	message := pieces[1]

	producer := kafka.NewProducer(bootstrapServer, topic, f.logger)

	return KafkaProduceRequest{
		BootstrapServer: bootstrapServer,
		Topic:           topic,
		Message:         message,

		producer: producer,
		count:    parseCountFromURL(u),
	}, nil

}

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
