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
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"github.com/sunchao1/hi-im-gateway/internal/conn"
	"github.com/sunchao1/hi-im-gateway/internal/protocol"
)

// Publisher sends uplink frames via hubclient FORWARD.
type Publisher interface {
	Publish(cmd uint32, payload []byte) error
}

// Kicker closes a connection by cid.
type Kicker interface {
	Kick(cid uint64)
}

// Dispatcher routes uplink cmds to handlers (beehive MesgRegister).
type Dispatcher struct {
	cfg     config.Config
	tab     *chattab.Table
	pub     Publisher
	kick    Kicker
	online  *OnlineHandler
	offline *OfflineHandler
	ping    *PingHandler
	comm    *CommHandler
	log     *slog.Logger
}

// NewDispatcher wires M5 uplink handlers.
func NewDispatcher(cfg config.Config, tab *chattab.Table, pub Publisher, kick Kicker) *Dispatcher {
	d := &Dispatcher{
		cfg:  cfg,
		tab:  tab,
		pub:  pub,
		kick: kick,
		log:  slog.Default(),
	}
	d.online = NewOnlineHandler(cfg, tab, pub)
	d.offline = NewOfflineHandler(kick)
	d.ping = NewPingHandler(cfg, pub, kick)
	d.comm = NewCommHandler(cfg, pub, kick)
	return d
}

// Handle dispatches an uplink IM frame from a WebSocket connection.
func (d *Dispatcher) Handle(c *conn.Conn, frame []byte) {
	if len(frame) < header.Size {
		return
	}
	hdr, err := header.Unmarshal(frame[:header.Size])
	if err != nil {
		d.log.Warn("uplink: bad header", "cid", c.Cid(), "err", err)
		return
	}

	switch hdr.Cmd {
	case cmd.CMD_ONLINE:
		d.online.Handle(c, hdr.Cmd, frame)
	case protocol.CMDOffline:
		d.offline.Handle(c, hdr.Cmd, frame)
	case cmd.CMD_PING:
		d.ping.Handle(c, hdr.Cmd, frame)
	default:
		d.comm.Handle(c, hdr.Cmd, frame)
	}
}
