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

	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"github.com/sunchao1/hi-im-gateway/internal/conn"
)

// OnlineHandler processes CMD_ONLINE (beehive LsndMesgOnlineHandler).
//
// 原作者: Qifeng.zou · 2017.03.04 · hi-im 迁移: sunchao1 · 2026.07.03
type OnlineHandler struct {
	cfg  config.Config
	tab  *chattab.Table
	pub  Publisher
	log  *slog.Logger
}

// NewOnlineHandler creates an ONLINE uplink handler.
func NewOnlineHandler(cfg config.Config, tab *chattab.Table, pub Publisher) *OnlineHandler {
	return &OnlineHandler{cfg: cfg, tab: tab, pub: pub, log: slog.Default()}
}

// Handle forwards ONLINE to Hub FORWARD and enters CHECK state.
func (h *OnlineHandler) Handle(c *conn.Conn, cmd uint32, frame []byte) {
	if !c.IsStatus(conn.StatusReady) && !c.IsStatus(conn.StatusCheck) {
		h.log.Debug("drop online: bad status", "cid", c.Cid(), "status", c.GetStatus())
		return
	}

	hdr, err := header.Unmarshal(frame[:header.Size])
	if err != nil {
		return
	}

	if h.tab.SessionGetParam(hdr.Sid, c.Cid()) != nil {
		h.log.Debug("drop online: duplicate", "cid", c.Cid(), "sid", hdr.Sid)
		return
	}

	c.SetSid(hdr.Sid)
	h.tab.SessionSetParam(hdr.Sid, c.Cid(), c)

	hdr.Cid = c.Cid()
	hdr.Nid = h.cfg.NID
	out, err := repack(hdr, frame[header.Size:])
	if err != nil {
		h.log.Warn("online: repack failed", "err", err)
		return
	}

	if err := h.pub.Publish(cmd, out); err != nil {
		h.log.Warn("online: publish failed", "err", err)
		return
	}
	c.SetStatus(conn.StatusCheck)
}
