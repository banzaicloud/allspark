package kafka

import (
	"context"
	"strconv"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/banzaicloud/allspark/internal/workload"
	kafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	bootstrapServer string
	topic           string
	workload        workload.Workload

	logger log.Logger
}

func NewProducer(bootStrapServer string, topic string, wl workload.Workload, logger log.Logger) (*Producer, error) {
	logger = logger.WithField("server", "producer")

	if wl.GetName() != workload.AirportWorkloadName && wl.GetName() != workload.EchoWorkloadName {
		return nil, errors.Errorf("invalid workload: '%s'", wl.GetName())
	}

	if bootStrapServer == "" {
		bootStrapServer = "kafka-all-broker.kafka:29092"
	}

	return &Producer{
		bootstrapServer: bootStrapServer,
		topic:           topic,
		workload:        wl,
		logger:          logger,
	}, nil
}

func (p *Producer) Produce(ctx context.Context) {
	w := kafka.Writer{
		Addr:   kafka.TCP(p.bootstrapServer),
		Topic:  p.topic,
		Logger: p.logger,
	}

	// counter used for message key
	var i int

	for {
		message, _, _ := p.workload.Execute()

		// each kafka message has a key and value. The key is used
		// to decide which partition (and consequently, which broker)
		// the message gets published on
		err := w.WriteMessages(ctx, kafka.Message{
			Key:   []byte(strconv.Itoa(i)),
			Value: []byte(message),
		})
		if err != nil {
			panic(errors.WrapIf(err, "could not write kafka message"))
		}

		// log a confirmation once the message is written
		p.logger.Infof("writes:", message)
		i++

		time.Sleep(200 * time.Millisecond)
	}
}
