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

// Package chattab provides in-process session indexing (beehive lib/chat_tab).
package chattab

import (
	"sync"
)

const sessionMaxLen = 999

// Table indexes sid/cid to connection handles for downlink routing.
type Table struct {
	mu       sync.RWMutex
	sessions [sessionMaxLen]map[uint64]map[uint64]any
	sidToCid map[uint64]uint64
	imGroup  *ImGroup
}

// New creates an empty ChatTab.
func New() *Table {
	return &Table{sidToCid: make(map[uint64]uint64)}
}

func shard(sid uint64) int {
	return int(sid % sessionMaxLen)
}

// SessionSetParam binds sid+cid to conn on ONLINE.
func (t *Table) SessionSetParam(sid, cid uint64, conn any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	idx := shard(sid)
	if t.sessions[idx] == nil {
		t.sessions[idx] = make(map[uint64]map[uint64]any)
	}
	if t.sessions[idx][sid] == nil {
		t.sessions[idx][sid] = make(map[uint64]any)
	}
	t.sessions[idx][sid][cid] = conn
}

// SessionGetParam looks up conn by sid+cid.
func (t *Table) SessionGetParam(sid, cid uint64) any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	idx := shard(sid)
	if t.sessions[idx] == nil {
		return nil
	}
	if t.sessions[idx][sid] == nil {
		return nil
	}
	return t.sessions[idx][sid][cid]
}

// GetCidBySid returns the canonical cid for sid (conflict detection).
func (t *Table) GetCidBySid(sid uint64) uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.sidToCid[sid]
}

// SessionSetCid updates sid→cid mapping after ONLINE_ACK.
func (t *Table) SessionSetCid(sid, cid uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sidToCid[sid] = cid
}

// SessionFindBySid returns the first cid+conn bound to sid (ONLINE CHECK state).
func (t *Table) SessionFindBySid(sid uint64) (uint64, any) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	idx := shard(sid)
	if t.sessions[idx] == nil {
		return 0, nil
	}
	m := t.sessions[idx][sid]
	if m == nil {
		return 0, nil
	}
	for cid, conn := range m {
		return cid, conn
	}
	return 0, nil
}

// SessionDel removes sid+cid binding on connection close.
func (t *Table) SessionDel(sid, cid uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	idx := shard(sid)
	if t.sessions[idx] != nil && t.sessions[idx][sid] != nil {
		delete(t.sessions[idx][sid], cid)
		if len(t.sessions[idx][sid]) == 0 {
			delete(t.sessions[idx], sid)
		}
	}
	if t.sidToCid[sid] == cid {
		delete(t.sidToCid, sid)
	}
}
