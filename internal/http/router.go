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
	"github.com/gin-gonic/gin"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"github.com/sunchao1/hi-im-gateway/internal/ws"
)

// Deps holds HTTP router dependencies.
type Deps struct {
	Config    config.Config
	WSServer  *ws.Server
	HubReady  func() bool
}

// NewRouter builds Gin routes for health checks and WebSocket upgrade.
func NewRouter(deps Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	health := NewHealth(deps.HubReady)
	r.GET("/healthz", func(c *gin.Context) {
		health.ServeHealthz(c.Writer, c.Request)
	})
	r.GET("/readyz", func(c *gin.Context) {
		health.ServeReadyz(c.Writer, c.Request)
	})
	r.GET(deps.Config.WSPath, gin.WrapH(deps.WSServer))
	return r
}
