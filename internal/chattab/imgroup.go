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

package chattab

import "sync"

type sessionKey struct {
	sid uint64
	cid uint64
}

// ImGroup tracks gid membership for second-stage fan-out (beehive chat_tab ImGroup).
type ImGroup struct {
	mu      sync.RWMutex
	members map[uint64]map[sessionKey]struct{} // gid -> sessions
}

func newImGroup() *ImGroup {
	return &ImGroup{members: make(map[uint64]map[sessionKey]struct{})}
}

// ImGroupJoin records sid+cid in gid.
func (t *Table) ImGroupJoin(gid, sid, cid uint64) {
	if t.imGroup == nil {
		t.imGroup = newImGroup()
	}
	t.imGroup.join(gid, sid, cid)
}

// ImGroupQuit removes sid+cid from gid.
func (t *Table) ImGroupQuit(gid, sid, cid uint64) {
	if t.imGroup == nil {
		return
	}
	t.imGroup.quit(gid, sid, cid)
}

// TravImGroupSession invokes fn for each local session in gid.
func (t *Table) TravImGroupSession(gid uint64, fn func(sid, cid uint64)) {
	if t.imGroup == nil {
		return
	}
	t.imGroup.trav(gid, fn)
}

func (g *ImGroup) join(gid, sid, cid uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.members[gid] == nil {
		g.members[gid] = make(map[sessionKey]struct{})
	}
	// Replace stale cid for the same sid (reconnect / duplicate ONLINE).
	for k := range g.members[gid] {
		if k.sid == sid {
			delete(g.members[gid], k)
		}
	}
	g.members[gid][sessionKey{sid: sid, cid: cid}] = struct{}{}
}

func (g *ImGroup) quit(gid, sid, cid uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	m, ok := g.members[gid]
	if !ok {
		return
	}
	delete(m, sessionKey{sid: sid, cid: cid})
	if len(m) == 0 {
		delete(g.members, gid)
	}
}

func (g *ImGroup) trav(gid uint64, fn func(sid, cid uint64)) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for k := range g.members[gid] {
		fn(k.sid, k.cid)
	}
}
