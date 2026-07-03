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

package uplink_test

import (
	"testing"

	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"github.com/sunchao1/hi-im-gateway/internal/conn"
	"github.com/sunchao1/hi-im-gateway/internal/ws/uplink"
)

type mockPub struct {
	lastCmd uint32
	lastLen int
}

func (m *mockPub) Publish(cmd uint32, payload []byte) error {
	m.lastCmd = cmd
	m.lastLen = len(payload)
	return nil
}

type mockKick struct {
	kicked []uint64
}

func (m *mockKick) Kick(cid uint64) { m.kicked = append(m.kicked, cid) }

func TestPingLocalPong(t *testing.T) {
	tab := chattab.New()
	pub := &mockPub{}
	kick := &mockKick{}
	cfg := config.Config{NID: 20001}
	d := uplink.NewDispatcher(cfg, tab, pub, kick)

	c := conn.New(1)
	c.SetSid(99)
	c.SetStatus(conn.StatusLogin)

	pingHdr := &header.Header{
		Cmd:    cmd.CMD_PING,
		Length: 0,
		Sid:    99,
		Cid:    1,
		Nid:    20001,
		Seq:    5,
	}
	frame, err := pingHdr.Pack()
	if err != nil {
		t.Fatal(err)
	}

	d.Handle(c, frame)

	if pub.lastCmd != cmd.CMD_PING {
		t.Fatalf("publish cmd = 0x%X, want PING", pub.lastCmd)
	}

	select {
	case pong := <-c.WriteChan():
		hdr, err := header.Unmarshal(pong)
		if err != nil {
			t.Fatal(err)
		}
		if hdr.Cmd != cmd.CMD_PONG {
			t.Fatalf("local pong cmd = 0x%X", hdr.Cmd)
		}
		if hdr.Seq != 5 {
			t.Fatalf("pong seq = %d, want 5", hdr.Seq)
		}
	default:
		t.Fatal("expected local PONG on write channel")
	}
}

func TestPingKicksWhenNotLogin(t *testing.T) {
	tab := chattab.New()
	kick := &mockKick{}
	cfg := config.Config{NID: 20001}
	d := uplink.NewDispatcher(cfg, tab, &mockPub{}, kick)

	c := conn.New(2)
	pingHdr := &header.Header{Cmd: cmd.CMD_PING, Cid: 2, Nid: 20001}
	frame, _ := pingHdr.Pack()
	d.Handle(c, frame)

	if len(kick.kicked) != 1 || kick.kicked[0] != 2 {
		t.Fatalf("expected kick cid 2, got %v", kick.kicked)
	}
}
