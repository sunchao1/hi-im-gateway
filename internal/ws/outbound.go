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

type downlinkItem struct {
	cid  uint64
	data []byte
	meta string
}

// PostDownlink queues a frame for async WS delivery (must not block hub handlers).
func (s *Server) PostDownlink(cid uint64, data []byte, meta string) bool {
	if s.downlinkQ == nil {
		return s.Send(cid, data)
	}
	buf := append([]byte(nil), data...)
	select {
	case s.downlinkQ <- downlinkItem{cid: cid, data: buf, meta: meta}:
		return true
	default:
		s.log.Warn("downlink queue full", "cid", cid, "meta", meta)
		return false
	}
}

func (s *Server) downlinkLoop() {
	for item := range s.downlinkQ {
		s.log.Info("downlink queued", "cid", item.cid, "meta", item.meta)
		if !s.Send(item.cid, item.data) {
			s.log.Warn("downlink ws enqueue failed", "cid", item.cid, "meta", item.meta)
			continue
		}
	}
}
