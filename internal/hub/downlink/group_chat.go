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

package downlink

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	imv1 "github.com/sunchao1/hi-im-api/gen/go/im/v1"
	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-gateway/internal/chattab"
	"github.com/sunchao1/hi-im-gateway/internal/config"
	"google.golang.org/protobuf/proto"
)

// GroupMemberAckHandler handles GROUP-CREAT/JOIN ACK and registers ImGroup (beehive LsndUpMesgGroupMemberAckHandler).
type GroupMemberAckHandler struct {
	tab    *chattab.Table
	comm   *CommHandler
	log    *slog.Logger
}

// NewGroupMemberAckHandler creates the handler.
func NewGroupMemberAckHandler(tab *chattab.Table, comm *CommHandler) *GroupMemberAckHandler {
	return &GroupMemberAckHandler{tab: tab, comm: comm, log: slog.Default()}
}

// Handle parses ACK, ImGroupJoin on success, then forwards to WS.
func (h *GroupMemberAckHandler) Handle(cmdID, _ uint32, payload []byte) {
	if len(payload) < header.Size {
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		return
	}
	if err := hdr.Validate(header.RequireSid); err != nil {
		h.comm.Handle(cmdID, 0, payload)
		return
	}

	ack := &imv1.GroupJoinAck{}
	if err := proto.Unmarshal(payload[header.Size:], ack); err != nil {
		h.comm.Handle(cmdID, 0, payload)
		return
	}

	if ack.GetCode() == 0 {
		if gid := parseGIDFromAck(ack.GetErrmsg()); gid > 0 {
			if cid := resolveConnCid(h.tab, hdr); cid != 0 {
				h.tab.ImGroupJoin(gid, hdr.Sid, cid)
				h.log.Info("ImGroupJoin", "gid", gid, "sid", hdr.Sid, "cid", cid, "cmd", fmt.Sprintf("0x%04X", cmdID))
			} else {
				h.log.Warn("group member ack: missing cid for ImGroupJoin", "sid", hdr.Sid, "gid", gid, "cmd", fmt.Sprintf("0x%04X", cmdID))
			}
		}
	}
	h.comm.Handle(cmdID, 0, payload)
}

// resolveConnCid picks the live WS cid for sid (prefer post-ONLINE mapping over echoed hdr.Cid).
func resolveConnCid(tab *chattab.Table, hdr *header.Header) uint64 {
	if hdr == nil {
		return 0
	}
	if cid := tab.GetCidBySid(hdr.Sid); cid != 0 {
		return cid
	}
	if hdr.Cid != 0 {
		return hdr.Cid
	}
	cid, _ := tab.SessionFindBySid(hdr.Sid)
	return cid
}

func parseGIDFromAck(errmsg string) uint64 {
	if strings.HasPrefix(errmsg, "Ok:") {
		gid, err := strconv.ParseUint(strings.TrimPrefix(errmsg, "Ok:"), 10, 64)
		if err == nil {
			return gid
		}
	}
	return 0
}

// GroupChatHandler fan-outs GROUP-CHAT to all local ImGroup members (beehive LsndUpMesgGroupChatHandler).
type GroupChatHandler struct {
	cfg    config.Config
	tab    *chattab.Table
	poster Poster
	log    *slog.Logger
}

// NewGroupChatHandler creates a GROUP-CHAT downlink handler.
func NewGroupChatHandler(cfg config.Config, tab *chattab.Table, poster Poster) *GroupChatHandler {
	return &GroupChatHandler{cfg: cfg, tab: tab, poster: poster, log: slog.Default()}
}

// Handle delivers GROUP-CHAT to every conn in the gid ImGroup table.
func (h *GroupChatHandler) Handle(cmdID, _ uint32, payload []byte) {
	if len(payload) < header.Size {
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		return
	}
	body := append([]byte(nil), payload[header.Size:]...)
	req := &imv1.GroupChat{}
	if err := proto.Unmarshal(body, req); err != nil {
		h.log.Warn("group-chat downlink: bad body", "err", err)
		return
	}
	gid := req.GetGid()
	if gid == 0 {
		return
	}
	text := req.GetText()
	members := 0
	h.tab.TravImGroupSession(gid, func(sid, imCid uint64) {
		members++
		cid := resolveConnCid(h.tab, &header.Header{Sid: sid, Cid: imCid})
		if cid == 0 {
			h.log.Warn("group-chat downlink: no live cid",
				"gid", gid, "sid", sid, "imCid", imCid, "text", text, "seq", hdr.Seq)
			return
		}
		frame, err := repackWSFrame(h.cfg.NID, cmdID, hdr, cid, body)
		if err != nil {
			h.log.Warn("group-chat downlink: repack failed", "err", err, "seq", hdr.Seq)
			return
		}
		meta := fmt.Sprintf("gid=%d sid=%d seq=%d text=%s", gid, sid, hdr.Seq, text)
		if !h.poster.PostDownlink(cid, frame, meta) {
			h.log.Warn("group-chat downlink: queue full",
				"gid", gid, "sid", sid, "cid", cid, "text", text, "seq", hdr.Seq)
		}
	})
	if members == 0 {
		h.log.Warn("group-chat downlink: no ImGroup members", "gid", gid, "text", text, "seq", hdr.Seq)
		return
	}
	h.log.Info("group-chat downlink: recv",
		"gid", gid, "uid", req.GetUid(), "text", text, "seq", hdr.Seq, "members", members)
}

func repackWSFrame(gatewayNID, cmdID uint32, hdr *header.Header, cid uint64, body []byte) ([]byte, error) {
	if cmdID == 0 {
		cmdID = cmd.CMD_GROUP_CHAT
	}
	out := &header.Header{
		Cmd:    cmdID,
		Length: uint32(len(body)),
		Sid:    hdr.Sid,
		Cid:    cid,
		Nid:    gatewayNID,
		Seq:    hdr.Seq,
	}
	headBuf, err := out.Pack()
	if err != nil {
		return nil, err
	}
	frame := make([]byte, len(headBuf)+len(body))
	copy(frame, headBuf)
	copy(frame[len(headBuf):], body)
	return frame, nil
}

// GroupChatAckHandler forwards GROUP-CHAT-ACK to sender via CommHandler path.
type GroupChatAckHandler struct {
	comm *CommHandler
}

// NewGroupChatAckHandler creates GROUP-CHAT-ACK handler.
func NewGroupChatAckHandler(comm *CommHandler) *GroupChatAckHandler {
	return &GroupChatAckHandler{comm: comm}
}

// Handle forwards ack to sender connection.
func (h *GroupChatAckHandler) Handle(cmdID, origNid uint32, payload []byte) {
	h.comm.Handle(cmdID, origNid, payload)
}

// Ensure compile-time cmd references.
var (
	_ = cmd.CMD_GROUP_CHAT
	_ = cmd.CMD_GROUP_CHAT_ACK
)
