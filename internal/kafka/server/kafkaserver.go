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

package server

import (
	"context"
	"net/http"
	"sync"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	segmentiokafka "github.com/segmentio/kafka-go"

	"github.com/banzaicloud/allspark/internal/kafka"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/banzaicloud/allspark/internal/request"
	"github.com/banzaicloud/allspark/internal/sql"
	"github.com/banzaicloud/allspark/internal/workload"
)

type Server struct {
	// the kafka Server acts as a consumer to handle incoming messages
	consumer *kafka.Consumer

	requests request.Requests
	workload workload.Workload

	sqlClient sql.Client

	errorHandler emperror.Handler
	logger       log.Logger
}

func New(consumer *kafka.Consumer, logger log.Logger, errorHandler emperror.Handler) *Server {
	logger = logger.WithField("server", "kafka")
	return &Server{
		consumer:     consumer,
		requests:     make(request.Requests, 0),
		logger:       logger,
		errorHandler: errorHandler,
	}
}

func (s *Server) SetWorkload(workload workload.Workload) {
	s.logger.WithField("name", workload.GetName()).Info("set workload")
	s.workload = workload
}

func (s *Server) SetRequests(requests request.Requests) {
	s.requests = requests
}

func (s *Server) SetSQLClient(client sql.Client) {
	s.sqlClient = client
}

func (s *Server) Run() {
	defer func() {
		err := s.consumer.Close()
		if err != nil {
			s.logger.Error(errors.WrapIf(err, "could not close kafka server"))
			return
		}
	}()
	for {
		message, err := s.consumer.Consume(context.Background())
		if err != nil {
			s.errorHandler.Handle(errors.WrapIf(err, "could not consume"))
			return
		}
		go s.Incoming(message)
	}
}

func (s *Server) Incoming(_ *segmentiokafka.Message) {
	s.logger.Info("incoming kafka consumer message")

	go s.doRequests(nil)

	if s.sqlClient != nil {
		go func() {
			query, err := s.sqlClient.RunQuery(s.logger)
			if err != nil {
				s.logger.WithFields(log.Fields{
					"query": query,
				}).Error(err)
			}
		}()
	}

	if s.workload == nil {
		return
	}

	_, _, err := s.workload.Execute()
	if err != nil {
		s.errorHandler.Handle(errors.WrapIf(err, "could not run workload"))
	}
}

func (s *Server) doRequests(incomingRequestHeaders http.Header) {
	var wg sync.WaitGroup

	for _, r := range s.requests {
		var i uint
		for i = 0; i < r.Count(); i++ {
			wg.Add(1)
			go func(request request.Request) {
				defer wg.Done()
				request.Do(incomingRequestHeaders, s.logger)
			}(r)
		}
	}

	wg.Wait()
}
