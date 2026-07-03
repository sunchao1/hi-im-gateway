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

package chattab_test

import (
	"testing"

	"github.com/sunchao1/hi-im-gateway/internal/chattab"
)

func TestSessionSetAndGet(t *testing.T) {
	tab := chattab.New()
	conn := "conn-a"
	tab.SessionSetParam(100, 1, conn)
	got := tab.SessionGetParam(100, 1)
	if got != conn {
		t.Fatalf("SessionGetParam = %v, want %v", got, conn)
	}
}

func TestSessionSetCidAndGetCidBySid(t *testing.T) {
	tab := chattab.New()
	tab.SessionSetCid(200, 42)
	if got := tab.GetCidBySid(200); got != 42 {
		t.Fatalf("GetCidBySid = %d, want 42", got)
	}
}

func TestSessionDelRemovesMapping(t *testing.T) {
	tab := chattab.New()
	tab.SessionSetParam(300, 7, "conn")
	tab.SessionSetCid(300, 7)
	tab.SessionDel(300, 7)
	if tab.SessionGetParam(300, 7) != nil {
		t.Fatal("SessionGetParam should be nil after del")
	}
	if tab.GetCidBySid(300) != 0 {
		t.Fatal("GetCidBySid should be 0 after del")
	}
}

func TestSidConflictDetection(t *testing.T) {
	tab := chattab.New()
	tab.SessionSetParam(400, 10, "conn-old")
	tab.SessionSetCid(400, 10)

	tab.SessionSetParam(400, 20, "conn-new")
	oldCid := tab.GetCidBySid(400)
	if oldCid != 10 {
		t.Fatalf("existing sid should map to cid 10, got %d", oldCid)
	}

	tab.SessionSetCid(400, 20)
	if got := tab.GetCidBySid(400); got != 20 {
		t.Fatalf("sid should now map to cid 20, got %d", got)
	}
}
