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

	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
)

// KickHandler processes CMD_KICK (beehive LsndUpMesgKickHandler).
//
// 原作者: Qifeng.zou · 2017.04.29 · hi-im 迁移: sunchao1 · 2026.07.03
type KickHandler struct {
	tab    *chattab.Table
	sender Sender
	log    *slog.Logger
}

// NewKickHandler creates a KICK downlink handler.
func NewKickHandler(tab *chattab.Table, sender Sender) *KickHandler {
	return &KickHandler{tab: tab, sender: sender, log: slog.Default()}
}

// Handle delivers KICK frame and closes the connection.
func (h *KickHandler) Handle(_ uint32, _ uint32, payload []byte) {
	if len(payload) < header.Size {
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		return
	}
	if hdr.Nid == 0 {
		return
	}

	cid := hdr.Cid
	if cid == 0 {
		cid = h.tab.GetCidBySid(hdr.Sid)
	}
	if cid == 0 {
		h.log.Warn("kick: cid not found", "sid", hdr.Sid)
		return
	}

	_ = h.sender.Send(cid, payload)
	h.sender.Kick(cid)
}
