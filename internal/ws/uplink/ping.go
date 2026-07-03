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

	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"github.com/sunchao1/hi-im-gateway/internal/conn"
)

// PingHandler processes CMD_PING (beehive LsndMesgPingHandler).
//
// 原作者: Qifeng.zou · 2017.03.04 · hi-im 迁移: sunchao1 · 2026.07.03
type PingHandler struct {
	cfg  config.Config
	pub  Publisher
	kick Kicker
	log  *slog.Logger
}

// NewPingHandler creates a PING uplink handler.
func NewPingHandler(cfg config.Config, pub Publisher, kick Kicker) *PingHandler {
	return &PingHandler{cfg: cfg, pub: pub, kick: kick, log: slog.Default()}
}

// Handle forwards PING to Hub and replies with local PONG.
func (h *PingHandler) Handle(c *conn.Conn, cmdVal uint32, frame []byte) {
	if !c.IsStatus(conn.StatusLogin) {
		h.kick.Kick(c.Cid())
		return
	}

	hdr, err := header.Unmarshal(frame[:header.Size])
	if err != nil {
		return
	}

	hdr.Cid = c.Cid()
	hdr.Nid = h.cfg.NID
	out, err := repack(hdr, frame[header.Size:])
	if err != nil {
		return
	}
	_ = h.pub.Publish(cmdVal, out)

	pongHdr := &header.Header{
		Cmd:    cmd.CMD_PONG,
		Length: 0,
		Sid:    hdr.Sid,
		Cid:    hdr.Cid,
		Nid:    hdr.Nid,
		Seq:    hdr.Seq,
	}
	pongFrame, err := pongHdr.Pack()
	if err != nil {
		return
	}
	_ = c.EnqueueWrite(pongFrame)
}
