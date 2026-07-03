// Copyright 2026 sunchao1
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"net/http"
)

// Health serves liveness and readiness probes.
type Health struct {
	hubReady func() bool
}

// NewHealth creates health handlers.
func NewHealth(hubReady func() bool) *Health {
	return &Health{hubReady: hubReady}
}

// ServeHealthz responds 200 when the process is alive.
func (h *Health) ServeHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// ServeReadyz responds 200 when hubclient FORWARD is ready.
func (h *Health) ServeReadyz(w http.ResponseWriter, _ *http.Request) {
	if h.hubReady != nil && h.hubReady() {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("not ready"))
}
