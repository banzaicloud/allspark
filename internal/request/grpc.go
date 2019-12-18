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

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/banzaicloud/allspark/internal/pb"
	"github.com/banzaicloud/allspark/internal/platform/log"
)

type GRPCRequest struct {
	Host    string `json:"host"`
	Service string `json:"service"`
	Method  string `json:"method"`

	count uint
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

	conn, err := grpc.Dial(request.Host, grpc.WithInsecure())
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
