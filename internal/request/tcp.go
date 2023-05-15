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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

type DialerFunc func(ctx context.Context, network string, address string) (net.Conn, error)

type tcpFactory struct {
	dialerFunc DialerFunc
}

type TCPFactoryOption func(*tcpFactory)

func WithTCPDialerFunc(dialerFunc DialerFunc) TCPFactoryOption {
	return func(f *tcpFactory) {
		f.dialerFunc = dialerFunc
	}
}

func NewTCPFactory(opts ...TCPFactoryOption) Factory {
	f := &tcpFactory{
		dialerFunc: (&net.Dialer{}).DialContext,
	}

	for _, o := range opts {
		o(f)
	}

	return f
}

func (f *tcpFactory) CreateRequest(u *url.URL) (Request, error) {
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, errors.WrapIf(err, "could not convert port to int")
	}

	return tcpRequest{
		Host:        u.Hostname(),
		Port:        port,
		PayloadSize: parseCountFromURL(u) * 1024 * 1024,

		dialerFunc: f.dialerFunc,
	}, nil
}

type tcpRequest struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	PayloadSize uint   `json:"payloadSize"`

	dialerFunc DialerFunc
}

func (request tcpRequest) Count() uint {
	return 1
}

func (request tcpRequest) Do(incomingRequestHeaders http.Header, logger log.Logger) {
	conn, err := request.dialerFunc(context.Background(), "tcp", fmt.Sprintf("%s:%d", request.Host, request.Port))
	if err != nil {
		logger.Error(errors.WrapIf(err, "could not connect"))
		return
	}
	defer func() {
		conn.Close()
	}()

	s := strings.Repeat(".", 16384)

	logger.WithField("payloadSize", request.PayloadSize).Info("sending data")

	conn.SetWriteDeadline(time.Now().Add(time.Second))

	var sum int
	i := 0
	for {
		i++
		sent, err := conn.Write([]byte(s))
		if err != nil {
			logger.Error(errors.WrapIf(err, "could not send data"))
			break
		}
		sum += sent
		if uint(sum) >= request.PayloadSize {
			break
		}
		// fmt.Printf("sending [%d] ...\n", i)
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	reply := make([]byte, 4096)
	if _, err = conn.Read(reply); err != nil {
		logger.Error(errors.WrapIf(err, "could not read data"))
	}

	logger.WithField("bytes", sum).Info("data sent")
}
