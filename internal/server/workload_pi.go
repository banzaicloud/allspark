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
	"math"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

const PIWorkloadName = "PI"

type PIWorkload struct {
	name   string
	count uint

	logger log.Logger
}

func NewPIWorkload(count uint, logger log.Logger) Workload {
	return &PIWorkload{
		name:   PIWorkloadName,
		count: count,

		logger: logger,
	}
}

func (w *PIWorkload) GetName() string {
	return w.name
}

func (w *PIWorkload) Execute() (string, string, error) {
	w.logger.WithField("n", w.count).Info("calculating pi")
	w.pi(w.count)

	return "ok", "text/plain", nil
}

// pi launches n goroutines to compute an
// approximation of pi.
func (w *PIWorkload) pi(n uint) float64 {
	ch := make(chan float64)
	for k := uint(0); k <= n; k++ {
		go w.term(ch, float64(k))
	}
	f := 0.0
	for k := uint(0); k <= n; k++ {
		f += <-ch
	}
	return f
}

func (w *PIWorkload) term(ch chan float64, k float64) {
	ch <- 4 * math.Pow(-1, k) / (2*k + 1)
}
