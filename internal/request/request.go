// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package request

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/banzaicloud/allspark/internal/kafka"
	"google.golang.org/grpc/metadata"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

var (
	headersToPropagate = []string{
		"X-Request-Id",
		"X-B3-Parentspanid",
		"X-B3-Traceid",
		"X-B3-Spanid",
		"X-B3-Sampled",
		"X-B3-Flags",
		"X-Ot-Span-Context",
		"X-Datadog-Trace-Id",
		"X-Datadog-Parent-Id",
		"X-Datadog-Sampled",
		"End-User",
		"User-Agent",
	}
)

type Request interface {
	Do(incomingRequestHeaders http.Header, logger log.Logger)
	Count() uint
}

type Requests []Request

func CreateRequestsFromStringSlice(reqs []string, logger log.Logger) (Requests, error) {
	var request Request

	requests := make(Requests, 0)

	for _, req := range reqs {
		pieces := strings.Split(req, "#")
		if len(pieces) == 0 {
			continue
		}

		if len(pieces) == 1 {
			request = HTTPRequest{
				URL: pieces[0],
			}
		}

		if len(pieces) == 2 {
			count, err := strconv.ParseUint(pieces[1], 10, 64)
			if err != nil {
				continue
			}
			request = HTTPRequest{
				URL:   pieces[0],
				count: uint(count),
			}
		}

		err := requests.AddRequest(request.(HTTPRequest), logger)
		if err != nil {
			return nil, errors.WrapIf(err, "could not add request")
		}
	}

	return requests, nil
}

func (r *Requests) AddRequest(request HTTPRequest, logger log.Logger) error {
	u, err := url.Parse(request.URL)
	if err == nil && (u.Scheme == "" || u.Host == "") {
		return emperror.With(errors.New("invalid URL"), "url", request.URL)
	}
	if err != nil {
		return err
	}

	var req Request

	switch u.Scheme {
	case "http", "https":
		req = request
	case "grpc":
		p := strings.SplitN(u.Path, "/", 3)
		if len(p) != 3 {
			return errors.New("invalid grpc url; service and/or method is missing")
		}

		req = GRPCRequest{
			Host:    u.Host,
			Service: p[1],
			Method:  p[2],
			count:   request.Count(),
		}
	case "tcp":
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return errors.WrapIf(err, "could not convert port to int")
		}
		req = TCPRequest{
			Host:        u.Hostname(),
			Port:        port,
			PayloadSize: request.Count() * 1024 * 1024,
		}
	case "kafka-consume":
		pieces := strings.Split(u.RawQuery, "=")
		if len(pieces) != 2 {
			return errors.New("invalid kafka consume url; provide only the consumer group after the '?'")
		}

		bootstrapServer := u.Host
		topic := strings.Trim(u.Path, "/")
		consumerGroup := pieces[1]

		consumer := kafka.NewConsumer(bootstrapServer, topic, consumerGroup, logger)

		req = KafkaConsumeRequest{
			consumer: consumer,
			count:    request.Count(),
		}
	case "kafka-produce":
		pieces := strings.Split(u.RawQuery, "=")
		if len(pieces) != 2 {
			return errors.New("invalid kafka produce url; provide only the message after the '?'")
		}

		bootstrapServer := u.Host
		topic := strings.Trim(u.Path, "/")
		message := pieces[1]

		producer := kafka.NewProducer(bootstrapServer, topic, logger)

		req = KafkaProduceRequest{
			producer: producer,
			Message:  message,
			count:    request.Count(),
		}
	}

	logger.WithFields(log.Fields{
		"url":   request.URL,
		"count": request.Count(),
	}).Info("request added")

	*r = append(*r, req)

	return nil
}

func propagateHeaders(incomingRequestHeaders http.Header, httpReq *http.Request) {
	for _, header := range headersToPropagate {
		val := incomingRequestHeaders.Get(header)
		if val != "" {
			httpReq.Header.Set(header, val)
		}
	}
}

func propagateGRPCHeaders(ctx context.Context, incomingRequestHeaders http.Header) context.Context {
	for _, header := range headersToPropagate {
		val := incomingRequestHeaders.Get(header)
		if val != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, header, val)
		}
	}

	return ctx
}
