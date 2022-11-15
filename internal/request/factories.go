// Copyright © 2022 Banzai Cloud
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
	"net/url"

	"emperror.dev/errors"

	"github.com/banzaicloud/allspark/internal/platform/log"
)

var (
	ErrorInvalidURL    = errors.New("invalid URL")
	ErrorInvalidScheme = errors.New("invalid request scheme")
)

type Factory interface {
	CreateRequest(u *url.URL) (Request, error)
}

var registeredRequestFactories = map[string]Factory{}

func RegisterFactory(scheme string, factory Factory) {
	registeredRequestFactories[scheme] = factory
}

func GetFactoryByScheme(scheme string) (Factory, bool) {
	if f, ok := registeredRequestFactories[scheme]; ok {
		return f, ok
	}

	return nil, false
}

func ParseRequests(rawRequests []string, logger log.Logger) (Requests, error) {
	requests := make(Requests, 0)

	for _, rawRequest := range rawRequests {
		if request, err := ParseRequest(rawRequest); err != nil {
			return nil, err
		} else { //nolint:revive
			requests = append(requests, request)
			logger.WithFields(log.Fields{
				"request": rawRequest,
			}).Info("request added")
		}
	}

	return requests, nil
}

func ParseRequest(req string) (Request, error) {
	u, err := url.Parse(req)
	if err == nil && (u.Scheme == "" || u.Host == "") {
		return nil, errors.WrapIfWithDetails(ErrorInvalidURL, "url", req)
	}
	if err != nil {
		return nil, err
	}

	if f, ok := GetFactoryByScheme(u.Scheme); ok {
		if req, err := f.CreateRequest(u); err != nil {
			return nil, err
		} else { //nolint:revive
			return req, nil
		}
	}

	return nil, errors.WithDetails(ErrorInvalidScheme, "scheme", u.Scheme)
}
