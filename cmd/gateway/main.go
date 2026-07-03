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

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	imhttp "github.com/sunchao1/hi-im-gateway/internal/http"
	"github.com/sunchao1/hi-im-gateway/internal/hub"
	"github.com/sunchao1/hi-im-gateway/internal/hub/downlink"
	"github.com/sunchao1/hi-im-gateway/internal/ws"
	"github.com/sunchao1/hi-im-gateway/internal/ws/uplink"
)

func main() {
	cfg, err := config.ConfigFromEnv()
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}

	log := newLogger(cfg.LogLevel)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tab := chattab.New()

	hubCfg, err := hub.ConfigFromEnv(cfg.NID, cfg.SubCmds, cfg.AuthUser, cfg.AuthPass, cfg.ForwardAddr)
	if err != nil {
		log.Error("hub config failed", "err", err)
		os.Exit(1)
	}
	hubClient, err := hub.NewClient(*hubCfg)
	if err != nil {
		log.Error("hub client init failed", "err", err)
		os.Exit(1)
	}

	kickFn := &kickAdapter{}
	dispatch := uplink.NewDispatcher(cfg, tab, hubClient, kickFn)
	wsServer := ws.NewServer(cfg, tab, dispatch)
	kickFn.kick = wsServer.Kick

	dl := downlink.NewRegistry(tab, wsServer)
	for cmd, handler := range dl.Handlers() {
		_ = hubClient.RegisterDownlink(cmd, handler)
	}
	_ = hubClient.RegisterDefaultDownlink(dl.DefaultHandler())

	if err := hubClient.Start(ctx); err != nil {
		log.Error("hub start failed", "err", err)
		os.Exit(1)
	}
	defer func() { _ = hubClient.Close() }()

	readyCtx, readyCancel := context.WithTimeout(ctx, 30*time.Second)
	defer readyCancel()
	if err := hubClient.WaitReady(readyCtx); err != nil {
		log.Error("hub wait ready failed", "err", err)
		os.Exit(1)
	}
	log.Info("hub FORWARD ready", "nid", cfg.NID, "addr", cfg.ForwardAddr)

	router := imhttp.NewRouter(imhttp.Deps{
		Config:   cfg,
		WSServer: wsServer,
		HubReady: hubClient.Ready,
	})

	httpSrv := &http.Server{
		Addr:    cfg.HTTPListen,
		Handler: router,
	}

	go func() {
		log.Info("gateway started", "http", cfg.HTTPListen, "ws", cfg.WSPath, "nid", cfg.NID)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http serve failed", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

// kickAdapter breaks the init cycle between uplink dispatcher and ws.Server.
type kickAdapter struct {
	kick func(uint64)
}

func (k *kickAdapter) Kick(cid uint64) {
	if k.kick != nil {
		k.kick(cid)
	}
}
