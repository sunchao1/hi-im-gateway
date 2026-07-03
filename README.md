# hi-im-gateway

hi-im 生态 **WebSocket 接入层**（Gin + gorilla/websocket + hubclient FORWARD）。契约定义见 [hi-im-api](https://github.com/sunchao1/hi-im-api)。

**作者**：sunchao1 · **许可证**：Apache License 2.0

## 依赖

- **hi-im-api** v0.1.0 — header、proto、cmd
- **hi-im-hubclient** v0.1.0 — Hub FORWARD 平面
- **hi-im-usrsvr** v0.1.0 — ONLINE 业务（BACKEND）

## 快速开始

```bash
export HIIM_FORWARD_ADDR="127.0.0.1:28888"
export HIIM_NID="20001"
export HIIM_AUTH_USER="websocket"
export HIIM_AUTH_PASS="websocket"
export HIIM_SUB_CMDS="0x0102,0x0106,0x0110"

make build && ./bin/gateway
```

客户端流程：先通过 usrsvr HTTP `register` + `iplist` 获取 token 与 WS 地址，再连接 gateway WebSocket 发送 ONLINE。

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `HIIM_HTTP_LISTEN` | `:8080` | Gin HTTP（含 WS） |
| `HIIM_WS_PATH` | `/ws` | WebSocket Upgrade 路径 |
| `HIIM_FORWARD_ADDR` | — | Hub FORWARD（必填） |
| `HIIM_NID` | `20001` | 本实例 NID |
| `HIIM_SHARD_ID` | `0` | Hub 分片（M8） |
| `HIIM_AUTH_USER` / `HIIM_AUTH_PASS` | `websocket` | Hub 认证 |
| `HIIM_SUB_CMDS` | `0x0102,0x0106,0x0110` | 下行 SUB 命令集 |
| `HIIM_MAX_CONN` | `100000` | 连接上限（软限） |
| `HIIM_LOG_LEVEL` | `info` | slog 级别 |

## 健康检查

- `GET /healthz` — 进程存活
- `GET /readyz` — hubclient FORWARD Ready

## 测试

```bash
make test
make test-integration   # 可选，含 //go:build integration
```

## Docker

```bash
make docker
docker run --rm -p 8080:8080 \
  -e HIIM_FORWARD_ADDR=hub:28888 \
  -e HIIM_NID=20001 \
  hi-im-gateway:latest
```

## Compose 对接（hi-im 主仓 M5+）

主仓 `profile biz` 将加入 **gateway ×2**（不同 NID）：

```yaml
hi-im-gateway-1:
  image: hi-im-gateway:latest
  environment:
    HIIM_FORWARD_ADDR: hub:28888
    HIIM_NID: "20001"
    HIIM_AUTH_USER: websocket
    HIIM_AUTH_PASS: websocket
    HIIM_SUB_CMDS: "0x0102,0x0106,0x0110"
  ports:
    - "8080:8080"
  depends_on:
    hub:
      condition: service_healthy
    hi-im-usrsvr:
      condition: service_healthy

hi-im-gateway-2:
  image: hi-im-gateway:latest
  environment:
    HIIM_FORWARD_ADDR: hub:28888
    HIIM_NID: "20002"
    HIIM_AUTH_USER: websocket
    HIIM_AUTH_PASS: websocket
    HIIM_SUB_CMDS: "0x0102,0x0106,0x0110"
  ports:
    - "8081:8080"
  depends_on:
    hub:
      condition: service_healthy
    hi-im-usrsvr:
      condition: service_healthy
```

`profile demo` 提供静态 Demo 页，依赖 `gateway` + `usrsvr` + `hub` + `redis` + `seqsvr`。

## 里程碑

| 阶段 | 本仓交付 |
|------|----------|
| **M5** | WS + ONLINE 全链路 + 双 gateway 实例 |
| **M6** | ImGroup + GROUP-CHAT 第二段 fan-out |
| **M8** | HPA、Prometheus、分片 NID |

## 文档

- [技术设计文档](doc/技术设计文档.md)
- [M1 实施清单](doc/M1-实施清单.md)
