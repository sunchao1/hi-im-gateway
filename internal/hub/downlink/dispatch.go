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

	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/protocol"
)

// Sender delivers downlink frames to WebSocket connections.
type Sender interface {
	Send(cid uint64, data []byte) bool
	Kick(cid uint64)
}

// Registry wires M5 downlink handlers (beehive UpMesgRegister).
type Registry struct {
	tab       *chattab.Table
	sender    Sender
	onlineAck *OnlineAckHandler
	kickH     *KickHandler
	comm      *CommHandler
	log       *slog.Logger
}

// NewRegistry creates downlink handlers.
func NewRegistry(tab *chattab.Table, sender Sender) *Registry {
	r := &Registry{tab: tab, sender: sender, log: slog.Default()}
	r.onlineAck = NewOnlineAckHandler(tab, sender)
	r.kickH = NewKickHandler(tab, sender)
	r.comm = NewCommHandler(tab, sender)
	return r
}

// Handlers returns cmd→handler map for hub registration.
func (r *Registry) Handlers() map[uint32]func(uint32, uint32, []byte) {
	return map[uint32]func(uint32, uint32, []byte){
		cmd.CMD_ONLINE_ACK: r.onlineAck.Handle,
		cmd.CMD_PONG:       r.comm.Handle,
		protocol.CMDKick:   r.kickH.Handle,
		protocol.CMDSubAck: r.comm.Handle,
	}
}

// DefaultHandler forwards unknown downlink frames via ChatTab lookup.
func (r *Registry) DefaultHandler() func(uint32, uint32, []byte) {
	return r.comm.Handle
}
