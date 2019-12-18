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

package errorhandler

import (
	"emperror.dev/emperror"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

type Handler struct {
	logger log.Logger
}

// NewHandler returns a handler which logs errors using the platform logger
func NewHandler(logger log.Logger) *Handler {
	return &Handler{logger: logger}
}

// Handle logs an error
func (h *Handler) Handle(err error) {
	var ctx map[string]interface{}

	// Extract context from the error and attach it to the log
	if kvs := emperror.Context(err); len(kvs) > 0 {
		ctx = ToMap(kvs)
	}

	logger := h.logger.WithFields(log.Fields(ctx))

	type errorCollection interface {
		Errors() []error
	}

	if errs, ok := err.(errorCollection); ok {
		for _, e := range errs.Errors() {
			ctx = nil
			// Extract context from the error and attach it to the log
			if kvs := emperror.Context(e); len(kvs) > 0 {
				ctx = ToMap(kvs)
			}
			h.logger.WithFields(log.Fields(ctx)).Error(e.Error())
		}
	} else {
		logger.Error(err.Error())
	}
}
