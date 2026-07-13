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

package ws

import (
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"github.com/sunchao1/hi-im-gateway/internal/conn"
	"github.com/sunchao1/hi-im-gateway/internal/ws/uplink"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server manages WebSocket connections and uplink dispatch.
type Server struct {
	cfg      config.Config
	tab      *chattab.Table
	dispatch *uplink.Dispatcher
	log      *slog.Logger

	nextCid   atomic.Uint64
	conns     sync.Map
	connCnt   atomic.Int64
	downlinkQ chan downlinkItem
}

// NewServer creates a WebSocket server.
func NewServer(cfg config.Config, tab *chattab.Table, dispatch *uplink.Dispatcher) *Server {
	s := &Server{
		cfg:       cfg,
		tab:       tab,
		dispatch:  dispatch,
		log:       slog.Default(),
		downlinkQ: make(chan downlinkItem, 4096),
	}
	go s.downlinkLoop()
	return s
}

// ServeHTTP upgrades HTTP to WebSocket and starts read/write loops.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.connCnt.Load() >= int64(s.cfg.MaxConn) {
		http.Error(w, "too many connections", http.StatusServiceUnavailable)
		return
	}

	raw, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Warn("ws upgrade failed", "err", err)
		return
	}

	cid := s.nextCid.Add(1)
	c := conn.New(cid)
	s.conns.Store(cid, c)
	s.connCnt.Add(1)

	go s.writeLoop(raw, c)
	go s.readLoop(raw, c)
}

func (s *Server) readLoop(raw *websocket.Conn, c *conn.Conn) {
	defer s.cleanup(raw, c)

	for {
		_, data, err := raw.ReadMessage()
		if err != nil {
			return
		}
		if len(data) < header.Size {
			s.log.Warn("frame too short", "cid", c.Cid(), "len", len(data))
			continue
		}
		s.dispatch.Handle(c, data)
	}
}

func (s *Server) writeLoop(raw *websocket.Conn, c *conn.Conn) {
	defer s.cleanup(raw, c)
	for {
		select {
		case data, ok := <-c.WriteChan():
			if !ok {
				return
			}
			if err := raw.WriteMessage(websocket.BinaryMessage, data); err != nil {
				s.log.Warn("ws write failed", "cid", c.Cid(), "sid", c.Sid(), "len", len(data), "err", err)
				return
			}
			s.log.Info("ws write ok", "cid", c.Cid(), "sid", c.Sid(), "len", len(data))
		case <-c.Closed():
			return
		}
	}
}

func (s *Server) cleanup(raw *websocket.Conn, c *conn.Conn) {
	c.Close()
	s.conns.Delete(c.Cid())
	s.connCnt.Add(-1)
	if c.Sid() != 0 {
		s.tab.SessionDel(c.Sid(), c.Cid())
	}
	_ = raw.Close()
}

// DownlinkPoster is implemented by ws.Server for non-blocking group fan-out.
type DownlinkPoster interface {
	PostDownlink(cid uint64, data []byte, meta string) bool
}

// Send delivers a downlink frame to the connection identified by cid.
func (s *Server) Send(cid uint64, data []byte) bool {
	v, ok := s.conns.Load(cid)
	if !ok {
		s.log.Warn("downlink send: unknown cid", "cid", cid)
		return false
	}
	conn, ok := v.(*conn.Conn)
	if !ok {
		return false
	}
	if !conn.EnqueueWrite(data) {
		s.log.Warn("downlink send: write queue full", "cid", cid, "sid", conn.Sid())
		return false
	}
	return true
}

// Kick closes the connection identified by cid.
func (s *Server) Kick(cid uint64) {
	v, ok := s.conns.Load(cid)
	if !ok {
		return
	}
	c, ok := v.(*conn.Conn)
	if !ok {
		return
	}
	c.SetStatus(conn.StatusLogout)
	c.Close()
	s.conns.Delete(cid)
	if c.Sid() != 0 {
		s.tab.SessionDel(c.Sid(), c.Cid())
	}
}

// GetConn returns the connection for sid+cid lookup in downlink handlers.
func (s *Server) GetConn(cid uint64) *conn.Conn {
	v, ok := s.conns.Load(cid)
	if !ok {
		return nil
	}
	c, _ := v.(*conn.Conn)
	return c
}
