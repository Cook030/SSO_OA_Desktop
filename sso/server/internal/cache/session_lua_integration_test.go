package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"mh-sso-svc/internal/consts"

	"github.com/redis/go-redis/v9"
)

// 本文件验证单设备登录正确性核心 —— loginReplaceLua 等 Lua 脚本在真实 Redis 上
// 的原子语义。登录时的"撤销旧会话 + 撤销旧 refresh token + 写入新会话 +
// 重置索引 + 为每个旧会话写下线事件"必须整体生效：
//
//   - 状态变则事件必在：被替换的会话必然在事件流中存在一条对应下线事件；
//   - 并发登录最终只有一个有效会话，其余全部为 replaced 且 token 已撤销；
//   - reject 模式在已有有效会话时整体拒绝，不留任何半成功状态；
//   - 陈旧会话的登出/撤销请求不会误伤之后登录产生的新会话。

const testUserID = uint64(10001)

type sessionAndToken struct {
	sessionID string
	rtHash    string
}

// streamEvent 一条会话事件流的消息
type streamEvent struct {
	ID              string
	EventID         string
	Type            string
	UserID          string
	TargetSessionID string
	Reason          string
	CreatedAt       string
}

func newSessionRecord(sid string) *SessionRecord {
	now := time.Now()
	return &SessionRecord{
		SessionID:    sid,
		UserID:       testUserID,
		Status:       consts.SessionStatusActive,
		LastActiveAt: now,
		ExpiredAt:    now.Add(2 * time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func newTokenRecord(sid, hash string) *RefreshTokenRecord {
	now := time.Now()
	return &RefreshTokenRecord{
		TokenHash: hash,
		SessionID: sid,
		UserID:    testUserID,
		Status:    consts.RefreshTokenStatusActive,
		ExpiredAt: now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// replaceLogin 以 replace 模式为 testUserID 建立新会话（真实走 Lua）
func replaceLogin(t *testing.T, c *Cache, st sessionAndToken) *LoginReplaceResult {
	t.Helper()
	ctx := context.Background()
	res, err := c.ReplaceLoginSession(ctx, LoginReplaceInput{
		UserID:       testUserID,
		Mode:         consts.LoginModeReplace,
		Session:      newSessionRecord(st.sessionID),
		SessionTTL:   time.Hour,
		RefreshToken: newTokenRecord(st.sessionID, st.rtHash),
		RefreshTTL:   time.Hour,
		Event: SessionEventInput{
			EventType: consts.EventTypeSessionReplaced,
			Reason:    consts.EventReasonReplacedByNewLogin,
		},
	})
	if err != nil {
		t.Fatalf("ReplaceLoginSession 失败: %v", err)
	}
	return res
}

func readStream(t *testing.T, cli *redis.Client) []streamEvent {
	t.Helper()
	msgs, err := cli.XRange(context.Background(), consts.SessionEventStreamKey, "-", "+").Result()
	if err != nil {
		t.Fatalf("读取事件流失败: %v", err)
	}
	out := make([]streamEvent, 0, len(msgs))
	for _, m := range msgs {
		ev := streamEvent{ID: m.ID}
		for k, v := range m.Values {
			switch k {
			case "eventId":
				ev.EventID = fmt.Sprint(v)
			case "type":
				ev.Type = fmt.Sprint(v)
			case "userId":
				ev.UserID = fmt.Sprint(v)
			case "targetSessionId":
				ev.TargetSessionID = fmt.Sprint(v)
			case "reason":
				ev.Reason = fmt.Sprint(v)
			case "createdAt":
				ev.CreatedAt = fmt.Sprint(v)
			}
		}
		out = append(out, ev)
	}
	return out
}

func sessionStatus(t *testing.T, c *Cache, sid string) int {
	t.Helper()
	rec, err := c.GetSession(sid)
	if err != nil {
		t.Fatalf("读取会话 %s 失败: %v", sid, err)
	}
	return rec.Status
}

func tokenStatus(t *testing.T, c *Cache, hash string) int {
	t.Helper()
	rec, err := c.GetRefreshToken(hash)
	if err != nil {
		t.Fatalf("读取 refresh token 失败: %v", err)
	}
	return rec.Status
}

func TestLoginReplaceSingleDeviceAtomically(t *testing.T) {
	cli := startRedis(t)
	flushAll(t, cli)
	requireRedisAtLeast(t, cli, 6, 0)
	c := newTestCache(cli)

	// 首次登录：无旧会话可替换
	first := replaceLogin(t, c, sessionAndToken{sessionID: "s1", rtHash: "rt1"})
	if len(first.ReplacedSessionIDs) != 0 {
		t.Fatalf("首次登录不应替换任何会话，实际 %v", first.ReplacedSessionIDs)
	}
	events := readStream(t, cli)
	if len(events) != 0 {
		t.Fatalf("首次登录不应写事件，实际 %d", len(events))
	}

	// 再次登录（单设备）：s1 必须被顶下线
	second := replaceLogin(t, c, sessionAndToken{sessionID: "s2", rtHash: "rt2"})
	if len(second.ReplacedSessionIDs) != 1 || second.ReplacedSessionIDs[0] != "s1" {
		t.Fatalf("二次登录应替换 s1，实际 %v", second.ReplacedSessionIDs)
	}

	// 状态一致：s1 → replaced，s2 → active
	if got := sessionStatus(t, c, "s1"); got != consts.SessionStatusReplaced {
		t.Fatalf("s1 状态应为 replaced(%d)，实际 %d", consts.SessionStatusReplaced, got)
	}
	if got := sessionStatus(t, c, "s2"); got != consts.SessionStatusActive {
		t.Fatalf("s2 状态应为 active(%d)，实际 %d", consts.SessionStatusActive, got)
	}
	// 旧 token 撤销、新 token 有效
	if got := tokenStatus(t, c, "rt1"); got != consts.RefreshTokenStatusRevoked {
		t.Fatalf("rt1 状态应为 revoked(%d)，实际 %d", consts.RefreshTokenStatusRevoked, got)
	}
	if got := tokenStatus(t, c, "rt2"); got != consts.RefreshTokenStatusActive {
		t.Fatalf("rt2 状态应为 active(%d)，实际 %d", consts.RefreshTokenStatusActive, got)
	}
	// current_session 与用户会话索引都只指向 s2
	cur, _ := c.GetCurrentSession(context.Background(), testUserID)
	if cur != "s2" {
		t.Fatalf("current_session 应为 s2，实际 %s", cur)
	}
	members, err := cli.SMembers(context.Background(), userSessionKey(testUserID)).Result()
	if err != nil || len(members) != 1 || members[0] != "s2" {
		t.Fatalf("用户会话索引应只含 s2，实际 %v err=%v", members, err)
	}

	// 状态变则事件必在：恰好一条针对 s1 的下线事件
	events = readStream(t, cli)
	if len(events) != 1 {
		t.Fatalf("事件流应恰好 1 条，实际 %d", len(events))
	}
	ev := events[0]
	if ev.TargetSessionID != "s1" || ev.Type != consts.EventTypeSessionReplaced ||
		ev.Reason != consts.EventReasonReplacedByNewLogin || ev.UserID != fmt.Sprint(testUserID) {
		t.Fatalf("下线事件字段不正确: %+v", ev)
	}
	if ev.EventID == "" || ev.CreatedAt == "" {
		t.Fatalf("下线事件应含 eventId 与 createdAt: %+v", ev)
	}
}

// 并发登录是单设备可靠性的核心竞态：
// 无论多少次并发 ReplaceLoginSession，最终必须恰好一个有效会话，
// 且"每次被顶下线的会话"都有一条与之对应的事件（不丢事件、不出现假下线）。
func TestConcurrentLoginKeepsExactlyOneActiveSession(t *testing.T) {
	cli := startRedis(t)
	flushAll(t, cli)
	requireRedisAtLeast(t, cli, 6, 0)
	c := newTestCache(cli)
	ctx := context.Background()

	const n = 16
	// 初始登录一次，制造一个待替换的旧会话
	replaceLogin(t, c, sessionAndToken{sessionID: "s-init", rtHash: "rt-init"})

	all := make([]sessionAndToken, 0, n)
	for i := 0; i < n; i++ {
		all = append(all, sessionAndToken{
			sessionID: fmt.Sprintf("s-conc-%d", i),
			rtHash:    fmt.Sprintf("rt-conc-%d", i),
		})
	}

	// 并发 16 次登录：真实竞态，Redis 串行执行 Lua 保证整体原子
	var wg sync.WaitGroup
	results := make([]*LoginReplaceResult, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 每次登录使用独立的 session / token，模拟不同登录请求
			res, err := c.ReplaceLoginSession(ctx, LoginReplaceInput{
				UserID:       testUserID,
				Mode:         consts.LoginModeReplace,
				Session:      newSessionRecord(all[i].sessionID),
				SessionTTL:   time.Hour,
				RefreshToken: newTokenRecord(all[i].sessionID, all[i].rtHash),
				RefreshTTL:   time.Hour,
				Event: SessionEventInput{
					EventType: consts.EventTypeSessionReplaced,
					Reason:    consts.EventReasonReplacedByNewLogin,
				},
			})
			if err != nil {
				t.Errorf("并发登录 %s 失败: %v", all[i].sessionID, err)
				return
			}
			results[i] = res
		}(i)
	}
	wg.Wait()

	// 结论 1：恰好一个会话存活 —— current_session 就是胜者
	winner, err := c.GetCurrentSession(ctx, testUserID)
	if err != nil || winner == "" {
		t.Fatalf("并发登录后应恰好存在一个 current_session，实际 %q err=%v", winner, err)
	}

	// 结论 2：全部会话（含初始）中除胜者外，状态与 token 都必须是被替换/撤销
	sessionIDs := []string{"s-init"}
	rtHashes := []string{"rt-init"}
	for _, st := range all {
		sessionIDs = append(sessionIDs, st.sessionID)
		rtHashes = append(rtHashes, st.rtHash)
	}
	activeCount := 0
	for i, sid := range sessionIDs {
		status := sessionStatus(t, c, sid)
		if sid == winner {
			if status != consts.SessionStatusActive {
				t.Fatalf("胜者会话 %s 应为 active，实际 %d", sid, status)
			}
			activeCount++
			continue
		}
		if status != consts.SessionStatusReplaced {
			t.Fatalf("会话 %s 应全部被 replaced(%d)，实际 %d", sid, consts.SessionStatusReplaced, status)
		}
		if got := tokenStatus(t, c, rtHashes[i]); got != consts.RefreshTokenStatusRevoked {
			t.Fatalf("会话 %s 的 token 应已撤销(%d)，实际 %d", sid, consts.RefreshTokenStatusRevoked, got)
		}
	}
	if activeCount != 1 {
		t.Fatalf("应恰好 1 个 active 会话，实际 %d", activeCount)
	}
	if got := sessionStatus(t, c, winner); got != consts.SessionStatusActive {
		t.Fatalf("胜者 %s 状态不一致: %d", winner, got)
	}

	// 结论 3：用户会话索引只含胜者
	members, err := cli.SMembers(ctx, userSessionKey(testUserID)).Result()
	if err != nil || len(members) != 1 || members[0] != winner {
		t.Fatalf("用户会话索引应只含胜者 %s，实际 %v err=%v", winner, members, err)
	}

	// 结论 4：事件流中，被替换的每个会话恰好一条事件，且没有"无中生有"的事件
	events := readStream(t, cli)
	replacedByEvent := map[string]bool{}
	for _, ev := range events {
		if ev.TargetSessionID == "" {
			t.Fatalf("事件缺少 targetSessionId: %+v", ev)
		}
		replacedByEvent[ev.TargetSessionID] = true
	}
	expectedReplaced := n // init + (n-1) 个后续登录产物，除胜者外全被替换
	if len(events) != expectedReplaced || len(replacedByEvent) != expectedReplaced {
		t.Fatalf("事件数应与被替换会话数一致（期望 %d，事件 %d，去重 %d）",
			expectedReplaced, len(events), len(replacedByEvent))
	}
	if replacedByEvent[winner] {
		t.Fatalf("胜者会话 %s 不应出现在下线事件中", winner)
	}
	for _, sid := range sessionIDs {
		if sid != winner && !replacedByEvent[sid] {
			t.Fatalf("被替换会话 %s 缺少对应下线事件", sid)
		}
	}
	// 事件字段完整（每条都有 eventId/type/reason）
	for _, ev := range events {
		if ev.EventID == "" || ev.Type == "" || ev.Reason == "" || ev.UserID == "" {
			t.Fatalf("事件字段不完整: %+v", ev)
		}
	}
}

func TestLoginRejectModeDeniesSecondActive(t *testing.T) {
	cli := startRedis(t)
	flushAll(t, cli)
	requireRedisAtLeast(t, cli, 6, 0)
	c := newTestCache(cli)
	ctx := context.Background()

	replaceLogin(t, c, sessionAndToken{sessionID: "s1", rtHash: "rt1"})
	before := len(readStream(t, cli))

	// reject 模式：已有有效会话，新登录应整体拒绝，不留任何半成功状态
	res, err := c.ReplaceLoginSession(ctx, LoginReplaceInput{
		UserID:       testUserID,
		Mode:         consts.LoginModeReject,
		Session:      newSessionRecord("s-rejected"),
		SessionTTL:   time.Hour,
		RefreshToken: newTokenRecord("s-rejected", "rt-rejected"),
		RefreshTTL:   time.Hour,
		Event: SessionEventInput{
			EventType: consts.EventTypeSessionReplaced,
			Reason:    consts.EventReasonReplacedByNewLogin,
		},
	})
	if err != nil {
		t.Fatalf("reject 登录不应报错: %v", err)
	}
	if !res.Rejected {
		t.Fatalf("已有有效会话时 reject 模式应返回 Rejected=true")
	}

	// 状态未发生任何变化
	if cur, _ := c.GetCurrentSession(ctx, testUserID); cur != "s1" {
		t.Fatalf("current_session 不应变化，实际 %s", cur)
	}
	if got := sessionStatus(t, c, "s1"); got != consts.SessionStatusActive {
		t.Fatalf("s1 状态不应变化，实际 %d", got)
	}
	if got := tokenStatus(t, c, "rt1"); got != consts.RefreshTokenStatusActive {
		t.Fatalf("rt1 状态不应变化，实际 %d", got)
	}
	// 被拒绝的会话与 token 不应残留
	if _, err := c.GetSession("s-rejected"); err == nil {
		t.Fatalf("被拒绝的会话不应写入 Redis")
	}
	if _, err := c.GetRefreshToken("rt-rejected"); err == nil {
		t.Fatalf("被拒绝的 refresh token 不应写入 Redis")
	}
	if after := len(readStream(t, cli)); after != before {
		t.Fatalf("reject 拒绝不应写入事件，before=%d after=%d", before, after)
	}
}

// 陈旧会话的登出/撤销不能误伤之后登录产生的新会话（current_session 已指向新会话）
func TestStaleRevokeDoesNotKillCurrentSession(t *testing.T) {
	cli := startRedis(t)
	flushAll(t, cli)
	requireRedisAtLeast(t, cli, 6, 0)
	c := newTestCache(cli)
	ctx := context.Background()

	replaceLogin(t, c, sessionAndToken{sessionID: "s1", rtHash: "rt1"})
	replaceLogin(t, c, sessionAndToken{sessionID: "s2", rtHash: "rt2"}) // 顶掉 s1
	eventsAfterReplaced := len(readStream(t, cli))

	// 旧设备拿着已失效的 s1 登出：不得影响当前会话 s2
	ok, err := c.RevokeSessionWithEvent(ctx, testUserID, "s1", consts.SessionStatusLoggedOut, SessionEventInput{
		EventType: consts.EventTypeSessionTerminated,
		Reason:    consts.EventReasonLoggedOut,
	})
	if err != nil {
		t.Fatalf("陈旧会话撤销不应报错: %v", err)
	}
	if ok {
		t.Fatalf("s1 早已被替换，RevokeSessionWithEvent 应返回 false")
	}
	if cur, _ := c.GetCurrentSession(ctx, testUserID); cur != "s2" {
		t.Fatalf("陈旧登出不应影响 current_session，实际 %s", cur)
	}
	if got := sessionStatus(t, c, "s2"); got != consts.SessionStatusActive {
		t.Fatalf("s2 应保持 active，实际 %d", got)
	}
	if after := len(readStream(t, cli)); after != eventsAfterReplaced {
		t.Fatalf("陈旧会话不应触发新事件，before=%d after=%d", eventsAfterReplaced, after)
	}

	// 当前会话 s2 登出：正常生效，current_session 清空并广播事件
	ok, err = c.RevokeSessionWithEvent(ctx, testUserID, "s2", consts.SessionStatusLoggedOut, SessionEventInput{
		EventType: consts.EventTypeSessionTerminated,
		Reason:    consts.EventReasonLoggedOut,
	})
	if err != nil || !ok {
		t.Fatalf("当前会话登出应成功，ok=%v err=%v", ok, err)
	}
	if cur, _ := c.GetCurrentSession(ctx, testUserID); cur != "" {
		t.Fatalf("登出后 current_session 应被清空，实际 %s", cur)
	}
	events := readStream(t, cli)
	last := events[len(events)-1]
	if last.TargetSessionID != "s2" || last.Type != consts.EventTypeSessionTerminated {
		t.Fatalf("登出事件应指向 s2，实际 %+v", last)
	}
}

func TestRevokeUserBroadcastsPerSessionEvents(t *testing.T) {
	cli := startRedis(t)
	flushAll(t, cli)
	requireRedisAtLeast(t, cli, 6, 0)
	c := newTestCache(cli)
	ctx := context.Background()

	replaceLogin(t, c, sessionAndToken{sessionID: "s1", rtHash: "rt1"})
	replaceLogin(t, c, sessionAndToken{sessionID: "s2", rtHash: "rt2"}) // s1 已被顶下线

	// 全部撤销：只应撤销仍 active 的 s2（s1 早已 replaced，不应重复计）
	revoked, err := c.RevokeUserSessionsWithEvents(ctx, testUserID, consts.SessionStatusRevoked, SessionEventInput{
		EventType: consts.EventTypeSessionTerminated,
		Reason:    consts.EventReasonRevokedByAdmin,
	})
	if err != nil {
		t.Fatalf("撤销全部会话失败: %v", err)
	}
	if len(revoked) != 1 || revoked[0] != "s2" {
		t.Fatalf("应只撤销 s2，实际 %v", revoked)
	}
	if cur, _ := c.GetCurrentSession(ctx, testUserID); cur != "" {
		t.Fatalf("全部撤销后 current_session 应为空，实际 %s", cur)
	}
	events := readStream(t, cli)
	last := events[len(events)-1]
	if last.TargetSessionID != "s2" || last.Reason != consts.EventReasonRevokedByAdmin {
		t.Fatalf("撤销事件应指向 s2，实际 %+v", last)
	}
}
