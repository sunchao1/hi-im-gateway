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

// Package conn holds per-WebSocket session state (beehive LsndConnExtra).
package conn

import (
	"sync"
	"sync/atomic"
)

// Status mirrors beehive connection states.
type Status int

const (
	StatusReady Status = iota + 1
	StatusCheck
	StatusLogin
	StatusLogout
)

// Conn holds per-WebSocket session state.
type Conn struct {
	cid       uint64
	sid       uint64
	status    atomic.Int32
	seq       uint64
	writeC    chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

// New allocates a connection with the given cid.
func New(cid uint64) *Conn {
	c := &Conn{
		cid:    cid,
		writeC: make(chan []byte, 2048),
		done:   make(chan struct{}),
	}
	c.SetStatus(StatusReady)
	return c
}

// Cid returns the process-local connection id.
func (c *Conn) Cid() uint64 { return c.cid }

// Sid returns the session id from ONLINE.
func (c *Conn) Sid() uint64 { return c.sid }

// SetSid stores the session id from ONLINE header.
func (c *Conn) SetSid(sid uint64) { c.sid = sid }

// GetStatus returns the current connection state.
func (c *Conn) GetStatus() Status { return Status(c.status.Load()) }

// IsStatus reports whether conn is in the given state.
func (c *Conn) IsStatus(s Status) bool {
	return c.GetStatus() == s
}

// SetStatus updates connection state.
func (c *Conn) SetStatus(s Status) {
	c.status.Store(int32(s))
}

// Seq returns the current monotonic sequence.
func (c *Conn) Seq() uint64 {
	return atomic.LoadUint64(&c.seq)
}

// InitSeq sets seq when server ack seq does not advance past client seq (demo/reconnect).
func (c *Conn) InitSeq(seq uint64) {
	if seq > 0 {
		atomic.StoreUint64(&c.seq, seq)
	}
}

// SetSeq initializes or advances seq (CAS monotonic).
func (c *Conn) SetSeq(seq uint64) bool {
	for {
		cur := atomic.LoadUint64(&c.seq)
		if seq <= cur {
			return false
		}
		if atomic.CompareAndSwapUint64(&c.seq, cur, seq) {
			return true
		}
	}
}

// EnqueueWrite sends a frame to the write goroutine (copies data).
// Non-blocking: returns false when the per-conn queue is full so downlinkLoop
// cannot stall on one slow WebSocket consumer.
func (c *Conn) EnqueueWrite(data []byte) bool {
	buf := append([]byte(nil), data...)
	select {
	case <-c.done:
		return false
	case c.writeC <- buf:
		return true
	default:
		return false
	}
}

// Close signals connection shutdown (idempotent).
func (c *Conn) Close() {
	c.closeOnce.Do(func() { close(c.done) })
}

// WriteChan exposes the write channel for the server write loop.
func (c *Conn) WriteChan() <-chan []byte { return c.writeC }

// Closed returns a channel signaled on connection close.
func (c *Conn) Closed() <-chan struct{} { return c.done }
