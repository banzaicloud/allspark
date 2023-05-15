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
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/banzaicloud/allspark/internal/pb"
	"github.com/banzaicloud/allspark/internal/platform/log"
)

type GRPCFactoryOption func(*grpcFactory)

func WithGRPCDialOptions(opts ...grpc.DialOption) GRPCFactoryOption {
	return func(f *grpcFactory) {
		f.dialOptions = opts
	}
}

type grpcFactory struct {
	dialOptions []grpc.DialOption
}

func NewGRPCFactory(opts ...GRPCFactoryOption) Factory {
	f := &grpcFactory{
		dialOptions: []grpc.DialOption{grpc.WithInsecure()},
	}

	for _, o := range opts {
		o(f)
	}

	return f
}

func (f *grpcFactory) CreateRequest(u *url.URL) (Request, error) {
	p := strings.SplitN(u.Path, "/", 3)
	if len(p) != 3 {
		return nil, errors.New("invalid grpc url; service and/or method is missing")
	}

	return GRPCRequest{
		Host:    u.Host,
		Service: p[1],
		Method:  p[2],
		count:   parseCountFromURL(u),

		dialOptions: f.dialOptions,
	}, nil
}

type GRPCRequest struct {
	Host    string `json:"host"`
	Service string `json:"service"`
	Method  string `json:"method"`

	count uint

	dialOptions []grpc.DialOption
}

func (request GRPCRequest) Count() uint {
	return request.count
}

func (request GRPCRequest) Do(incomingRequestHeaders http.Header, logger log.Logger) {
	correlationID := uuid.New()
	log := logger.WithFields(log.Fields{
		"host":          request.Host,
		"service":       request.Service,
		"method":        request.Method,
		"correlationID": correlationID,
	})
	log.Info("outgoing request")

	ctx := propagateGRPCHeaders(context.Background(), incomingRequestHeaders)

	do := []grpc.DialOption{}
	do = append(do, request.dialOptions...)
	do = append(do, grpc.WithTimeout(time.Second*3))

	conn, err := grpc.Dial(request.Host, do...)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	c := pb.NewAllsparkClient(conn)
	_, err = c.Incoming(ctx, &pb.Params{})
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info("response to outgoing request")
}
