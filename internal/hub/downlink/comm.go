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

// CommHandler forwards downlink frames to WS (beehive LsndUpMesgCommHandler).
//
// 原作者: Qifeng.zou · 2017.03.06 · hi-im 迁移: sunchao1 · 2026.07.03
type CommHandler struct {
	tab    *chattab.Table
	sender Sender
	log    *slog.Logger
}

// NewCommHandler creates the default downlink forwarder.
func NewCommHandler(tab *chattab.Table, sender Sender) *CommHandler {
	return &CommHandler{tab: tab, sender: sender, log: slog.Default()}
}

// Handle looks up conn by sid/cid and writes the frame to WS.
func (h *CommHandler) Handle(_ uint32, _ uint32, payload []byte) {
	if len(payload) < header.Size {
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		return
	}
	if err := hdr.Validate(header.RequireSid); err != nil {
		return
	}

	cid := hdr.Cid
	if cid == 0 {
		cid = h.tab.GetCidBySid(hdr.Sid)
	}
	if cid == 0 {
		return
	}
	if h.tab.SessionGetParam(hdr.Sid, cid) == nil {
		return
	}
	_ = h.sender.Send(cid, payload)
}
