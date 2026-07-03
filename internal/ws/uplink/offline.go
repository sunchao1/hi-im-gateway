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

package uplink

import (
	"log/slog"

	"github.com/sunchao1/hi-im-gateway/internal/conn"
)

// OfflineHandler processes CMD_OFFLINE (beehive LsndMesgOfflineHandler).
//
// 原作者: Qifeng.zou · 2017.03.06 · hi-im 迁移: sunchao1 · 2026.07.03
type OfflineHandler struct {
	kick Kicker
	log  *slog.Logger
}

// NewOfflineHandler creates an OFFLINE uplink handler.
func NewOfflineHandler(kick Kicker) *OfflineHandler {
	return &OfflineHandler{kick: kick, log: slog.Default()}
}

// Handle kicks the connection on client OFFLINE.
func (h *OfflineHandler) Handle(c *conn.Conn, _ uint32, _ []byte) {
	if !c.IsStatus(conn.StatusLogin) {
		h.kick.Kick(c.Cid())
		return
	}
	c.SetStatus(conn.StatusLogout)
	h.kick.Kick(c.Cid())
}
