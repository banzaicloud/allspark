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
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/goph/emperror"
	"github.com/pkg/errors"

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

func New(config Config, logger log.Logger, errorHandler emperror.Handler) *Server {
	return &Server{
		requests: make([]Request, 0),

		listenAddress: config.ListenAddress,
		endpoint:      config.Endpoint,

		errorHandler: errorHandler,
		logger:       logger,
	}
}

func (s *Server) SetWorkload(workload Workload) {
	s.logger.WithField("name", workload.GetName()).Info("set workload")
	s.workload = workload
}

func (s *Server) AddRequest(request Request) error {
	u, err := url.Parse(request.URL)
	if err == nil && (u.Scheme == "" || u.Host == "") {
		return emperror.With(errors.New("invalid URL"), "url", request.URL)
	}
	if err != nil {
		return err
	}

	s.logger.WithFields(log.Fields{
		"url":   request.URL,
		"count": request.Count,
	}).Info("request added")

	s.requests = append(s.requests, request)

	return nil
}

func (s *Server) AddRequestsFromStringSlice(reqs []string) {
	var request Request

	for _, req := range reqs {
		pieces := strings.Split(req, "#")
		if len(pieces) == 0 {
			continue
		}

		if len(pieces) == 1 {
			request = Request{
				URL: pieces[0],
			}
		}

		if len(pieces) == 2 {
			count, err := strconv.ParseUint(pieces[1], 10, 64)
			if err != nil {
				continue
			}
			request = Request{
				URL:   pieces[0],
				Count: uint(count),
			}
		}

		err := s.AddRequest(request)
		if err != nil {
			s.errorHandler.Handle(emperror.Wrap(err, "could not add request"))
		}
	}

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
	err := r.Run(s.listenAddress)
	if err != nil {
		s.errorHandler.Handle(err)
	}
}

func (s *Server) runWorkload() (string, string, error) {
	if s.workload == nil {
		return "", "", nil
	}

	response, contentType, err := s.workload.Execute()
	if err != nil {
		return "", contentType, emperror.Wrap(err, "could not run workload")
	}

	return response, contentType, nil
}

func (s *Server) doRequests(incomingRequestHeaders http.Header) {
	var wg sync.WaitGroup

	for _, request := range s.requests {
		var i uint
		for i = 0; i < request.Count; i++ {
			wg.Add(1)
			go func(srv *Server, request Request) {
				defer wg.Done()
				request.Do(incomingRequestHeaders, srv)
			}(s, request)
		}
	}

	wg.Wait()
}

func (request Request) Do(incomingRequestHeaders http.Header, s *Server) {
	s.logger.WithFields(log.Fields{
		"url": request.URL,
	}).Info("outgoing request")

	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", request.URL, nil)
	if err != nil {
		s.logger.WithFields(log.Fields{
			"url": request.URL,
		}).Error(err.Error())
		return
	}

	propagateHeaders(incomingRequestHeaders, httpReq)

	response, err := httpClient.Do(httpReq)
	if err != nil {
		s.logger.WithFields(log.Fields{
			"url": request.URL,
		}).Error(err.Error())
		return
	}

	defer response.Body.Close()
	s.logger.WithFields(log.Fields{
		"url":          request.URL,
		"responseCode": response.StatusCode,
	}).Info("response to outgoing request")
}

func propagateHeaders(incomingRequestHeaders http.Header, httpReq *http.Request) {
	for _, header := range headersToPropagate {
		val := incomingRequestHeaders.Get(header)
		if val != "" {
			httpReq.Header.Set(header, val)
		}
	}
}
