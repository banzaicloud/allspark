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

package httpserver

import (
	"net/http"
	"sync"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/gin-gonic/gin"

	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/banzaicloud/allspark/internal/request"
	"github.com/banzaicloud/allspark/internal/workload"
)

type Server struct {
	requests request.Requests
	workload workload.Workload

	listenAddress string
	endpoint      string

	errorHandler emperror.Handler
	logger       log.Logger
}

func New(config Config, logger log.Logger, errorHandler emperror.Handler) *Server {
	logger = logger.WithField("server", "http")
	return &Server{
		requests: make([]request.Request, 0),

		listenAddress: config.ListenAddress,
		endpoint:      config.Endpoint,

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

func (s *Server) Run() {
	r := gin.New()
	r.GET(s.endpoint, func(c *gin.Context) {
		s.doRequests(c.Request.Header)
		response, contentType, err := s.runWorkload()
		if err != nil {
			ginErr := c.AbortWithError(503, err)
			if ginErr != nil {
				s.errorHandler.Handle(ginErr)
			}
			return
		}
		c.Data(http.StatusOK, contentType, []byte(response))
	})
	s.logger.WithField("address", s.listenAddress).Info("starting HTTP server")
	err := r.Run(s.listenAddress)
	if err != nil {
		s.errorHandler.Handle(err)
	}
}

func (s *Server) runWorkload() (string, string, error) {
	if s.workload == nil {
		return "ok", "text/plain", nil
	}

	response, contentType, err := s.workload.Execute()
	if err != nil {
		return "", contentType, errors.WrapIf(err, "could not run workload")
	}

	return response, contentType, nil
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
