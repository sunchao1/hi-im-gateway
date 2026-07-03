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

package hub

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/sunchao1/hi-im-hubclient/pkg/hubclient"
)

// Handler processes a downlink IM frame from Hub async_send.
type Handler func(cmd, origNid uint32, payload []byte)

// Client wraps hubclient FORWARD lifecycle and downlink handler registry.
type Client struct {
	inner    *hubclient.Client
	handlers map[uint32]Handler
	defaultH Handler
	mu       sync.RWMutex
	log      *slog.Logger
}

// NewClient builds a hub FORWARD client.
func NewClient(cfg hubclient.Config) (*Client, error) {
	c, err := hubclient.New(&cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		inner:    c,
		handlers: make(map[uint32]Handler),
		log:      slog.Default(),
	}, nil
}

// RegisterDownlink registers a handler for a specific downlink cmd.
func (c *Client) RegisterDownlink(cmd uint32, h Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[cmd] = h
	return c.inner.RegisterHandler(cmd, c.dispatch)
}

// RegisterDefaultDownlink registers the fallback handler for unknown cmds.
func (c *Client) RegisterDefaultDownlink(h Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultH = h
	return c.inner.RegisterHandler(0, c.dispatch)
}

func (c *Client) dispatch(cmd, origNid uint32, payload []byte) {
	c.mu.RLock()
	h := c.handlers[cmd]
	if h == nil {
		h = c.defaultH
	}
	c.mu.RUnlock()
	if h == nil {
		c.log.Debug("drop downlink: no handler", "cmd", fmt.Sprintf("0x%04X", cmd))
		return
	}
	h(cmd, origNid, payload)
}

// Start launches the hub client goroutines.
func (c *Client) Start(ctx context.Context) error {
	return c.inner.Start(ctx)
}

// WaitReady blocks until AUTH+SUB complete.
func (c *Client) WaitReady(ctx context.Context) error {
	return c.inner.WaitReady(ctx)
}

// Ready reports handshake completion.
func (c *Client) Ready() bool {
	return c.inner.Ready()
}

// Publish enqueues an uplink IM frame via FORWARD publish (destNid=0).
func (c *Client) Publish(cmd uint32, payload []byte) error {
	return c.inner.AsyncSend(cmd, 0, payload)
}

// Close stops the hub client.
func (c *Client) Close() error {
	return c.inner.Close()
}

// ConfigFromEnv builds hubclient.Config for FORWARD plane only.
func ConfigFromEnv(nid uint32, subCmdsCSV, authUser, authPass, forwardAddr string) (*hubclient.Config, error) {
	cfg := hubclient.DefaultConfig()
	cfg.Addr = forwardAddr
	cfg.NID = nid
	cfg.User = authUser
	cfg.Pass = authPass
	cmds, err := parseHexList(subCmdsCSV)
	if err != nil {
		return nil, fmt.Errorf("HIIM_SUB_CMDS: %w", err)
	}
	cfg.Subscribe = cmds
	return cfg, nil
}

func parseHexList(s string) ([]uint32, error) {
	parts := strings.Split(s, ",")
	out := make([]uint32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseUint(p, 0, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, uint32(v))
	}
	return out, nil
}
