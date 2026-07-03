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

package downlink

import (
	"log/slog"

	imv1 "github.com/sunchao1/hi-im-api/gen/go/im/v1"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/conn"
	"google.golang.org/protobuf/proto"
)

// OnlineAckHandler processes CMD_ONLINE_ACK (beehive LsndUpMesgOnlineAckHandler).
//
// 原作者: Qifeng.zou · 2017.03.07 · hi-im 迁移: sunchao1 · 2026.07.03
type OnlineAckHandler struct {
	tab    *chattab.Table
	sender Sender
	log    *slog.Logger
}

// NewOnlineAckHandler creates an ONLINE_ACK downlink handler.
func NewOnlineAckHandler(tab *chattab.Table, sender Sender) *OnlineAckHandler {
	return &OnlineAckHandler{tab: tab, sender: sender, log: slog.Default()}
}

// Handle delivers ONLINE_ACK to WS and transitions conn to LOGIN.
func (h *OnlineAckHandler) Handle(_ uint32, _ uint32, payload []byte) {
	if len(payload) < header.Size {
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		return
	}
	if err := hdr.Validate(header.RequireSid); err != nil {
		h.log.Warn("online_ack: invalid header", "err", err)
		return
	}

	cid := hdr.Cid
	var ack imv1.OnlineAck
	if err := proto.Unmarshal(payload[header.Size:], &ack); err != nil {
		h.sendAndKick(cid, payload)
		return
	}
	if ack.GetCode() != 0 {
		h.log.Warn("online failed", "cid", cid, "sid", ack.GetSid(), "code", ack.GetCode(), "errmsg", ack.GetErrmsg())
		h.sendAndKick(cid, payload)
		return
	}

	extra := h.tab.SessionGetParam(ack.GetSid(), cid)
	if extra == nil {
		h.sendAndKick(cid, payload)
		return
	}
	c, ok := extra.(*conn.Conn)
	if !ok {
		h.sendAndKick(cid, payload)
		return
	}
	if !c.SetSeq(ack.GetSeq()) {
		h.sendAndKick(cid, payload)
		return
	}

	c.SetStatus(conn.StatusLogin)

	oldCid := h.tab.GetCidBySid(ack.GetSid())
	if oldCid != 0 && oldCid != cid {
		h.sender.Kick(oldCid)
	}
	h.tab.SessionSetCid(ack.GetSid(), cid)

	_ = h.sender.Send(cid, payload)
}

func (h *OnlineAckHandler) sendAndKick(cid uint64, payload []byte) {
	_ = h.sender.Send(cid, payload)
	h.sender.Kick(cid)
}
