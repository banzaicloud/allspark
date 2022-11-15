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

func parseCountFromURL(u *url.URL) uint {
	if count, err := strconv.ParseUint(u.Fragment, 10, 64); err == nil {
		return uint(count)
	}

	return 1
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
