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
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"emperror.dev/errors"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

type TCPRequest struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	PayloadSize uint   `json:"payloadSize"`
}

func (request TCPRequest) Count() uint {
	return 1
}

func (request TCPRequest) Do(incomingRequestHeaders http.Header, logger log.Logger) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", request.Host, request.Port))
	if err != nil {
		logger.Error(errors.WrapIf(err, "could not connect"))
		return
	}

	defer conn.Close()

	s := strings.Repeat(".", 1024)

	logger.WithField("payloadSize", request.PayloadSize).Info("sending data")

	var sum int
	for {
		sent, err := conn.Write([]byte(s))
		if err != nil {
			logger.Error(errors.WrapIf(err, "could not send response"))
			break
		}
		sum += sent
		if uint(sum) >= request.PayloadSize {
			break
		}
	}

	// read the response from the server
	tmp := make([]byte, 256)
	var sumRead int
	for {
		read, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				logger.Error("read error", err)
			}
			break
		}
		sumRead += read
		// the server returns payloadsize / 2 bytes
		if uint(sumRead) >= request.PayloadSize/2 {
			break
		}
	}

	logger.WithField("bytes", sum).Info("data sent")
}
