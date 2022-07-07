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

package workload

import (
	"bytes"
	"encoding/csv"
	"math/rand"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/allspark/assets"
	"github.com/banzaicloud/allspark/internal/platform/log"
)

var _ Workload = &KafkaWorkload{}

const KafkaWorkloadName = "Kafka"

type KafkaWorkload struct {
	name     string
	airports [][]string

	logger log.Logger
}

func NewKafkaWorkload(logger log.Logger) Workload {
	rand.Seed(time.Now().Unix())

	csvReader := csv.NewReader(bytes.NewReader(assets.AirportCodes))
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(errors.WrapIf(err, "Unable to parse Airport asset CSV"))
	}

	return &KafkaWorkload{
		name:     KafkaWorkloadName,
		airports: records,

		logger: logger,
	}
}

func (w *KafkaWorkload) GetName() string {
	return w.name
}

func (w *KafkaWorkload) Execute() (string, string, error) {
	randomAirportInfo := w.airports[rand.Intn(len(w.airports))]
	result := strings.Join(randomAirportInfo, ",")

	return result, "text/plain", nil
}
