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
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"github.com/sunchao1/hi-im-gateway/internal/conn"
)

// CommHandler forwards LOGIN-state business cmds (beehive LsndMesgCommHandler).
//
// 原作者: Qifeng.zou · 2017.03.04 · hi-im 迁移: sunchao1 · 2026.07.03
type CommHandler struct {
	cfg  config.Config
	pub  Publisher
	kick Kicker
	log  *slog.Logger
}

// NewCommHandler creates the default uplink forwarder.
func NewCommHandler(cfg config.Config, pub Publisher, kick Kicker) *CommHandler {
	return &CommHandler{cfg: cfg, pub: pub, kick: kick, log: slog.Default()}
}

// Handle publishes the frame when conn is LOGIN.
func (h *CommHandler) Handle(c *conn.Conn, cmd uint32, frame []byte) {
	if !c.IsStatus(conn.StatusLogin) {
		h.kick.Kick(c.Cid())
		h.log.Debug("kick: not login", "cid", c.Cid(), "cmd", cmd)
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
	_ = h.pub.Publish(cmd, out)
}

func repack(hdr *header.Header, body []byte) ([]byte, error) {
	buf, err := hdr.Pack()
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return buf, nil
	}
	out := make([]byte, len(buf)+len(body))
	copy(out, buf)
	copy(out[len(buf):], body)
	return out, nil
}
