采用 Redis 为会话权威状态的最终方案如下。它不依赖 MySQL Outbox、不引入 Kafka/gRPC，也不把 WebSocket 变成认证服务；核心正确性由 Redis Lua 原子脚本和 Redis Streams 提供。

前提选型固定为：

- Redis 使用单主节点 + Sentinel/副本高可用，**不使用 Redis Cluster**。
- 开启 AOF，`appendfsync always`；会话状态和下线事件以 Redis 为权威数据。
- SSO 是唯一能修改 session 状态的服务。
- WebSocket Gateway 只维护连接、消费事件、投递与 ACK，不修改登录状态。
- 使用 Redis Streams，**不使用 Pub/Sub** 作为关键下线消息通道。

不选 Redis Cluster 的原因是本方案需要一个 Lua 脚本原子修改“用户索引、多个会话、多个 refresh token、事件 Stream”。跨 slot 脚本会破坏这一原子性；在这里，单主 Redis 的正确性和可验证性优先于水平拆分的优雅性。

---

## 最终架构

```text
浏览器
  │ HTTPS / WSS
  ├───────────── SSO
  │               ├─ 登录、JWT、session 校验
  │               └─ Redis Lua：置换登录状态 + 写事件
  │
  └───────────── WebSocket Gateway
                  ├─ 连接管理、Ping/Pong、重连、ACK
                  ├─ 调用 SSO 内部鉴权接口
                  └─ 消费 Redis Stream 并向目标 session 推送

Redis（权威）
  ├─ session / refresh token / user-session index
  ├─ sso:stream:session-events
  ├─ sso:delivery:{sessionId}
  └─ ws:connections:{sessionId}
```

---

## 1. 单设备登录的原子操作

不要继续沿用当前“创建新 session 后再撤销”的普通 Go 调用链。将登录成功后的会话操作收敛成一个 Lua 脚本：

```text
1. 读取 user:{userId}:sessions 中全部 session
2. 将全部旧 session 标记为 revoked/replaced
3. 将旧 session 下所有 active refresh token 标记 revoked
4. 写入新 session 和新 refresh token
5. 重置 user:{userId}:sessions，只保留新 session
6. 写入 user:{userId}:current_session = newSessionId
7. 对每个旧 session，XADD 一条 session_replaced 事件
8. 一次 EVAL 完成；任一步失败，整个操作不生效
```

新的 Redis key 约定：

```text
sso:session:{sessionId}                 session JSON
sso:rt:{tokenHash}                      refresh token JSON
sso:user:{userId}:sessions              Set，当前用户会话集合
sso:user:{userId}:current_session       String，唯一有效 session
sso:session:{sessionId}:tokens          Set，该 session 的 refresh token hash
sso:stream:session-events               Stream，SSO 领域事件
```

JWT 校验必须同时满足：

```text
JWT 签名、过期时间合法
AND JWT.sessionId == sso:user:{userId}:current_session
AND sso:session:{sessionId}.status == active
```

这样即使旧客户端未收到任何 WebSocket 消息，旧 access token 也立即失效；这才是单设备登录的安全边界。

现有 [session_service.go](F:/GoCode/mh_oa_manage/sso/server/internal/service/session_service.go) 中 `CreateLoginSession`、`revokeAllUserSessions` 等分步调用应由这个脚本替代。现有 [session.go](F:/GoCode/mh_oa_manage/sso/server/internal/cache/session.go) 的 JSON 结构可以保留。

---

## 2. WebSocket 心跳和连接管理

WebSocket Gateway 对每条连接维护唯一的读协程和写协程。

```text
服务端每 25 秒 Ping
客户端自动 Pong
45 秒未收到 Pong：关闭连接
客户端以带抖动的指数退避重连，最大间隔 30 秒
```

连接状态只放在 Gateway 所在 Redis 实例共享的短 TTL key 中：

```text
ws:connections:{sessionId}
```

内容记录 `gatewayInstanceId`、`connectionId`、最近心跳时间；TTL 为 60 秒，每次 Pong/业务心跳续期。它是路由辅助数据，不是安全状态。

每个连接的写队列必须有上限，例如 128 条：

- 普通通知：队列满时允许丢弃或合并。
- `session_replaced`：不能静默丢弃；关闭该慢连接即可。
- 关闭后，客户端重连将因旧 session 已失效而被拒绝，并从 HTTP 的 401 错误中获得下线原因。

这比无限排队更安全：无限队列会被慢客户端拖垮 Gateway。

---

## 3. 消息触达语义

这里选择并明确实现“至少一次投递”，不承诺不可能做到的“设备离线时立刻展示”。

### SSO → Gateway：可靠消费

SSO 在同一个 Lua 脚本中写状态和 Stream：

```text
XADD sso:stream:session-events *
  eventId=uuid
  type=session_replaced
  userId=...
  targetSessionId=oldSessionId
  reason=replaced_by_new_login
  createdAt=...
```

Gateway 通过 Consumer Group 消费：

```text
XREADGROUP GROUP ws-gateway {instanceId}
```

Gateway 崩溃前未确认的 Stream 消息会保留在 Pending Entries List，可由同一实例恢复或其他实例通过 `XAUTOCLAIM` 接管。因此不会因 Gateway 短暂下线而丢消息。

### Gateway → 客户端：ACK 和重投

Gateway 接到 Stream 事件后，必须先把它原子保存至：

```text
sso:delivery:{targetSessionId}
```

然后才 `XACK sso:stream:session-events`。

客户端协议：

```json
{
  "type": "event",
  "eventId": "0d8f...",
  "eventType": "session_replaced",
  "reason": "replaced_by_new_login"
}
```

客户端处理事件后回复：

```json
{
  "type": "ack",
  "eventId": "0d8f..."
}
```

Gateway 收到 ACK 后才从 `sso:delivery:{sessionId}` 删除对应消息。未 ACK 的消息在连接重建时重新发送。

客户端须按 `eventId` 幂等处理：同一 `session_replaced` 即使收到多次，也只执行一次“停止业务请求、提示、跳转登录页”。

### 旧设备已经离线怎么办

旧 session 已被撤销，所以它不应该被允许使用旧凭证重建 WebSocket。此时不能要求它“补收 WebSocket 下线事件”。

正确行为是：

```text
旧设备下一次 API 调用 / 刷新 token / 尝试重连
→ SSO 返回 401，业务码 SESSION_REPLACED
→ 前端显示“账号已在其他设备登录”并跳转登录页
```

因此，WebSocket 的成功送达只优化在线设备体验；HTTP/SSO 的 session 校验确保离线、断网、崩溃场景仍然安全。

---

## 4. 独立 WebSocket 服务的对内接口

不提供“业务服务直连某台 Gateway 的推送接口”。这会引入服务发现、连接归属和失败重试问题，也会让业务服务与 Gateway 可用性耦合。

对内接口只保留两类：

| 接口 | 使用者 | 用途 |
|---|---|---|
| Redis Stream `sso:stream:session-events` | SSO → Gateway | 认证、安全类事件；必须可靠 |
| SSO 内部鉴权 HTTP 接口 | Gateway → SSO | WebSocket 握手和重连时验证 access token/session |

Gateway 与 SSO 的内部鉴权接口建议为：

```text
POST /internal/v1/sessions/introspect
Authorization: Bearer {accessToken}
```

响应：

```json
{
  "active": true,
  "userId": 1001,
  "sessionId": "session_xxx",
  "expiresAt": "2026-09-08T..."
}
```

约束：

- 只允许 Gateway 的内网地址访问。
- mTLS 或服务间签名认证。
- 不允许公网访问。
- Gateway 不持有 JWT 签名密钥，不连接用户数据库。
- Gateway 只接受 SSO 返回 `active=true` 的连接。

现有 `/api/v1/auth/introspect` 可调整为这个内部接口，但必须补上服务间鉴权；当前不应把它作为公开接口。

---

## 5. WebSocket 对客户端的最小协议

```text
WSS /api/v1/realtime/ws
```

握手认证通过 Cookie 中现有 access token；Gateway 将该 token 转给 SSO 内部 introspect。不要将 access token 放在 URL query 参数中。

服务端消息：

```json
{ "type": "event", "eventId": "...", "eventType": "session_replaced", "reason": "replaced_by_new_login" }
```

客户端消息：

```json
{ "type": "ack", "eventId": "..." }
```

控制帧：

```text
Ping / Pong：连接保活
```

不额外设计业务 heartbeat JSON；Ping/Pong 已能正确处理断网和半开连接。若未来需要设备展示信息，再增加业务心跳，但它不应参与本次单设备登录正确性。

---

## 6. 必须验证的验收场景

1. A 在线，B 登录同一账号：A 在一个心跳周期内收到 `session_replaced` 并退出。
2. A 在线但 Gateway 断开：B 登录后 A 的任一 API 请求立即 401，错误码为 `SESSION_REPLACED`。
3. A 离线：B 登录后 A 恢复网络，旧 token 无法调用 API、无法建立 WebSocket。
4. SSO 在 Lua 执行前失败：旧会话保持有效，新会话不存在。
5. SSO 在 Lua 成功后、HTTP 响应前失败：系统最多保留最后一次成功登录的唯一 session，不会多端有效。
6. Gateway 在读取 Stream 后崩溃：事件能从 Pending Entries List 被恢复消费。
7. 客户端收到事件但 ACK 丢失：重复事件不会导致异常，客户端幂等退出。
8. 两个登录请求并发：Lua 串行执行，最终只保留最后完成脚本的 session。
9. Redis 主从切换：AOF 恢复后 session 状态与 Stream 事件一致；登录替换不会出现“状态变了、事件未写”的半成功。

这套设计的边界很清楚：Redis 可用时，登录置换和下线事件具有原子一致性；Redis 故障时，SSO 应拒绝登录和鉴权，不能退化为“放行但不校验会话”。