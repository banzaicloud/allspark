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

package tcpserver

import (
	"net"
	"net/http"
	"sync"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/banzaicloud/allspark/internal/request"
	"github.com/banzaicloud/allspark/internal/sql"
	"github.com/banzaicloud/allspark/internal/workload"
)

type Server struct {
	requests request.Requests
	workload workload.Workload

	sqlCient *sql.Client

	listenAddress string

	errorHandler emperror.Handler
	logger       log.Logger
}

func New(config Config, logger log.Logger, errorHandler emperror.Handler) *Server {
	logger = logger.WithField("server", "tcp")
	return &Server{
		requests: make(request.Requests, 0),

		listenAddress: config.ListenAddress,

		errorHandler: errorHandler,
		logger:       logger,
	}
}

func (s *Server) SetWorkload(workload workload.Workload) {
	s.logger.WithField("name", workload.GetName()).Info("set workload")
	s.workload = workload
}

func (s *Server) SetRequests(requests request.Requests) {
	s.requests = requests
}

func (s *Server) SetSQLClient(client *sql.Client) {
	s.sqlCient = client
}

func (s *Server) Run() {
	lis, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		s.errorHandler.Handle(errors.WrapIf(err, "could not listen"))
		return
	}
	s.logger.WithField("address", s.listenAddress).Info("starting TCP server")

	for {
		c, err := lis.Accept()
		if err != nil {
			s.errorHandler.Handle(errors.WrapIf(err, "could not accept connection"))
			return
		}
		go s.Incoming(c)
	}
}

func (s *Server) Incoming(c net.Conn) {
	s.logger.Info("incoming TCP request")
	defer func() {
		c.Close()
	}()

	go s.doRequests(nil)

	if s.sqlCient != nil {
		go func() {
			query, err := s.sqlCient.RunQuery(s.logger)
			if err != nil {
				s.logger.WithFields(log.Fields{
					"query": query,
				}).Error(err)
			}
		}()
	}

	tmp := make([]byte, 4096)
	for {
		_, err := c.Read(tmp)
		if err != nil {
			break
		}
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
