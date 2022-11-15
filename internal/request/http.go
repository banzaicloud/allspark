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
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

type httpFactory struct {
	transport http.RoundTripper
}

type HTTPFactoryOption func(*httpFactory)

func WithHTTPTransport(transport http.RoundTripper) HTTPFactoryOption {
	return func(f *httpFactory) {
		f.transport = transport
	}
}

func NewHTTPFactory(opts ...HTTPFactoryOption) Factory {
	f := &httpFactory{
		transport: http.DefaultTransport,
	}

	for _, o := range opts {
		o(f)
	}

	return f
}

func (f *httpFactory) CreateRequest(u *url.URL) (Request, error) {
	r := HTTPRequest{
		URL:       u,
		transport: f.transport,
		count:     parseCountFromURL(u),
	}

	return r, nil
}

type HTTPRequest struct {
	URL *url.URL

	count uint

	transport http.RoundTripper
}

func (request HTTPRequest) Count() uint {
	return request.count
}

func (request HTTPRequest) Do(incomingRequestHeaders http.Header, logger log.Logger) {
	correlationID := uuid.New()
	logger.WithFields(log.Fields{
		"url":           request.URL,
		"correlationID": correlationID,
	}).Info("outgoing request")

	httpClient := &http.Client{
		Transport: request.transport,
	}
	httpReq, err := http.NewRequest("GET", request.URL.String(), nil)
	if err != nil {
		logger.WithFields(log.Fields{
			"url": request.URL,
		}).Error(err.Error())
		return
	}
	httpReq.Close = true

	propagateHeaders(incomingRequestHeaders, httpReq)

	response, err := httpClient.Do(httpReq)
	if err != nil {
		logger.WithFields(log.Fields{
			"url": request.URL,
		}).Error(err.Error())
		return
	}
	defer response.Body.Close()

	logger.WithFields(log.Fields{
		"url":           request.URL,
		"responseCode":  response.StatusCode,
		"correlationID": correlationID,
	}).Info("response to outgoing request")
}
