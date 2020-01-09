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

	"github.com/google/uuid"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

type HTTPRequest struct {
	URL string `json:"URL"`

	count uint
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

	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", request.URL, nil)
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
