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

package grpcserver

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"github.com/banzaicloud/allspark/internal/pb"
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
	logger = logger.WithField("server", "grpc")
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
	s.logger.WithField("address", s.listenAddress).Info("starting GRPC server")

	gs := grpc.NewServer(
		grpc.ConnectionTimeout(time.Second),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Second * 10,
			Timeout:           time.Second * 20,
		}),
		grpc.KeepaliveEnforcementPolicy(
			keepalive.EnforcementPolicy{
				MinTime:             time.Second,
				PermitWithoutStream: true,
			}),
		grpc.MaxConcurrentStreams(5),
	)
	pb.RegisterAllsparkServer(gs, s)

	// Register reflection service on gRPC server.
	reflection.Register(gs)
	if err := gs.Serve(lis); err != nil {
		s.errorHandler.Handle(errors.WrapIf(err, "could not serve"))
	}
}

func (s *Server) Incoming(ctx context.Context, x *pb.Params) (*pb.Msg, error) {
	s.logger.Info("incoming request")

	headers := make(http.Header)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for h, vs := range md {
			for _, v := range vs {
				headers.Add(h, v)
			}
		}
	}

	s.doRequests(headers)

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

	if s.workload == nil {
		return &pb.Msg{}, nil
	}

	response, _, err := s.workload.Execute()
	if err != nil {
		return &pb.Msg{}, errors.WrapIf(err, "could not run workload")
	}

	return &pb.Msg{
		Response: response,
	}, nil
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
