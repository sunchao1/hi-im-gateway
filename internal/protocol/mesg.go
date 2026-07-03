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

package protocol

// Kick is the CMD 0x0110 body (beehive mesg.MesgKick).
//
// 原作者: Qifeng.zou · hi-im 迁移: sunchao1 · 2026.07.03
type Kick struct {
	Code   uint32 `protobuf:"varint,1,opt,name=code,proto3"`
	Errmsg string `protobuf:"bytes,2,opt,name=errmsg,proto3"`
}
