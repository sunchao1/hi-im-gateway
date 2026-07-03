# hi-im-gateway 文档

> **hi-im-gateway** 是 hi-im 生态 **L3 WebSocket 接入服务**（Gin + WS + hubclient FORWARD）；对应必嗨 websocket。  
> **状态**：**M5 可联调**（WS + ONLINE 全链路 + 双 gateway 实例）  
> **许可证**：Apache License 2.0（见仓库根目录 `LICENSE`）  
> **作者**：sunchao1

---

## 阅读顺序

| 顺序 | 文档 | 内容 |
|------|------|------|
| 1 | [技术设计文档.md](技术设计文档.md) | WS 接入、FORWARD 上下行、ChatTab、M5/M6 边界 |
| 2 | [M1-实施清单.md](M1-实施清单.md) | 生态 **M5** 任务拆解（依赖 M4 usrsvr） |

---

## 生态对照

| 文档 | 说明 |
|------|------|
| [hi-im/doc/hi-im-档C技术方案设计.md](https://github.com/sunchao1/hi-im/blob/main/doc/hi-im-档C技术方案设计.md) | 生态总方案 §3 双平面、§7 gateway、§11 M5 |
| [hi-im-hubclient/doc/技术设计文档.md](https://github.com/sunchao1/hi-im-hubclient/blob/main/doc/技术设计文档.md) | FORWARD AsyncSend / RegisterHandler |
| [hi-im-usrsvr/doc/技术设计文档.md](https://github.com/sunchao1/hi-im-usrsvr/blob/main/doc/技术设计文档.md) | ONLINE / iplist / token |
| beehive-im `src/golang/exec/websocket/` | 必嗨 websocket 对照 |

---

## 角色对照

```text
档 C 总方案     →  hi-im/doc/hi-im-档C技术方案设计.md
契约            →  hi-im-api（header / proto / cmd）
Hub 传输        →  hi-im-hubclient（FORWARD 平面）
WS 接入         →  hi-im-gateway（本仓库）
用户/会话       →  hi-im-usrsvr（ONLINE 业务）
群聊第一段      →  hi-im-msgsvr（M6）
群聊第二段      →  gateway ChatTab ImGroup（M6）
```
