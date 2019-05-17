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
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/goph/emperror"
)

type Request struct {
	URL   string `json:"URL"`
	Count uint   `json:"count"`
}

type Workload interface {
	GetName() string
	Execute() (string, string, error)
}

type Server struct {
	requests []Request
	workload Workload

	listenAddress string
	endpoint      string

	errorHandler emperror.Handler
	logger       log.Logger
}

