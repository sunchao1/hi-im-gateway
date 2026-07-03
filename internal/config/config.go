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

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds gateway runtime settings loaded from environment variables.
type Config struct {
	HTTPListen  string
	WSPath      string
	ForwardAddr string
	NID         uint32
	ShardID     uint32
	AuthUser    string
	AuthPass    string
	SubCmds     string
	MaxConn     int
	LogLevel    string
}

// ConfigFromEnv loads configuration from HIIM_* environment variables.
func ConfigFromEnv() (Config, error) {
	cfg := Config{
		HTTPListen:  envOr("HIIM_HTTP_LISTEN", ":8080"),
		WSPath:      envOr("HIIM_WS_PATH", "/ws"),
		ForwardAddr: os.Getenv("HIIM_FORWARD_ADDR"),
		AuthUser:    envOr("HIIM_AUTH_USER", "websocket"),
		AuthPass:    envOr("HIIM_AUTH_PASS", "websocket"),
		SubCmds:     envOr("HIIM_SUB_CMDS", "0x0102,0x0106,0x0110,0x0302,0x0306,0x030B,0x030C"),
		MaxConn:     envIntOr("HIIM_MAX_CONN", 100000),
		LogLevel:    envOr("HIIM_LOG_LEVEL", "info"),
	}

	if v := os.Getenv("HIIM_NID"); v != "" {
		n, err := parseU32(v)
		if err != nil {
			return cfg, fmt.Errorf("HIIM_NID: %w", err)
		}
		cfg.NID = n
	} else {
		cfg.NID = 20001
	}

	if v := os.Getenv("HIIM_SHARD_ID"); v != "" {
		n, err := parseU32(v)
		if err != nil {
			return cfg, fmt.Errorf("HIIM_SHARD_ID: %w", err)
		}
		cfg.ShardID = n
	}

	if cfg.ForwardAddr == "" {
		return cfg, fmt.Errorf("HIIM_FORWARD_ADDR is required")
	}
	if cfg.MaxConn <= 0 {
		return cfg, fmt.Errorf("HIIM_MAX_CONN must be positive")
	}
	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOr(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseU32(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
