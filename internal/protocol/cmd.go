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

// Package protocol holds IM command constants and message bodies not yet exported
// by hi-im-api v0.1.0. Migrated from beehive lib/comm/mesg.go and lib/mesg.
package protocol

// Lifecycle and control commands (beehive lib/comm/mesg.go).
const (
	// CMDOffline — 下线请求（必嗨 0x0103）
	// 原作者: Qifeng.zou · 2017.03.06 · hi-im 迁移: sunchao1 · 2026.07.03
	CMDOffline = 0x0103

	// CMDSubAck — 订阅应答（必嗨 0x0107）
	// 原作者: Qifeng.zou · hi-im 迁移: sunchao1 · 2026.07.03
	CMDSubAck = 0x0107

	// CMDKick — 踢人请求（必嗨 0x0110）
	// 原作者: Qifeng.zou · 2017.04.29 · hi-im 迁移: sunchao1 · 2026.07.03
	CMDKick = 0x0110
)
