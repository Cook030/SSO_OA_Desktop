import { FormEvent, useEffect, useRef, useState } from "react";
import { AgentStatus, AppName, CurrentSession, ListConversations, ListMessages, Login, Logout, NewConversation, SendAgentMessage } from "../wailsjs/go/main/App";
import "./App.css";

type Session = { account: string; displayName: string; role: string };
type Conversation = { id: number; title: string; updatedAt: string };
type Message = { id: number; conversationId: number; role: "user" | "assistant"; content: string; createdAt: string };
type Status = { oaMode: string; oaConfigured: boolean; toolCount: number };

function App() {
  const [session, setSession] = useState<Session | null>(null);
  const [appName, setAppName] = useState("智工工作台");
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [conversationId, setConversationId] = useState(0);
  const [messages, setMessages] = useState<Message[]>([]);
  const [draft, setDraft] = useState("");
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [login, setLogin] = useState({ account: "admin", password: "admin123" });
  const [status, setStatus] = useState<Status>({ oaMode: "dry-run", oaConfigured: false, toolCount: 1 });
  const chatEnd = useRef<HTMLDivElement>(null);

  const show = (text: string) => { setNotice(text); window.setTimeout(() => setNotice(""), 3500); };
  const refreshConversations = async () => setConversations(await ListConversations() as Conversation[]);
  const selectConversation = async (id: number) => { setConversationId(id); setMessages(await ListMessages(id) as Message[]); };
  const createConversation = async () => { const conversation = await NewConversation() as Conversation; await refreshConversations(); await selectConversation(conversation.id); };

  useEffect(() => {
    AppName().then(setAppName).catch(() => undefined);
    CurrentSession().then(value => setSession(value as Session | null)).catch(() => undefined);
    AgentStatus().then(value => setStatus(value as Status)).catch(() => undefined);
  }, []);
  useEffect(() => { if (session) createConversation().catch(error => show(error.message)); }, [session]);
  useEffect(() => { chatEnd.current?.scrollIntoView({ behavior: "smooth" }); }, [messages, busy]);

  const submitLogin = async (event: FormEvent) => { event.preventDefault(); try { setSession(await Login(login.account, login.password) as Session); } catch (error) { show(error instanceof Error ? error.message : "登录失败"); } };
  const send = async (event: FormEvent) => {
    event.preventDefault(); if (!draft.trim() || busy) return;
    setBusy(true); const content = draft; setDraft("");
    try { const reply = await SendAgentMessage({ conversationId, content }); setConversationId(reply.conversationId); setMessages(reply.messages as Message[]); await refreshConversations(); }
    catch (error) { show(error instanceof Error ? error.message : "任务执行失败"); setDraft(content); }
    finally { setBusy(false); }
  };

  if (!session) return <LoginScreen appName={appName} login={login} setLogin={setLogin} submit={submitLogin} notice={notice} />;
  return <main className="agent-shell">
    <aside className="agent-sidebar">
      <div className="agent-brand"><span>✦</span><strong>{appName}</strong></div>
      <button className="new-chat" onClick={createConversation}>＋ 新会话</button>
      <div className="side-section"><div className="section-label">工作区</div><div className="workspace active">▣ OA 权限助手</div></div>
      <div className="side-section conversations"><div className="section-label">会话</div>{conversations.map(item => <button key={item.id} className={item.id === conversationId ? "conversation active" : "conversation"} onClick={() => selectConversation(item.id)}>{item.title}</button>)}</div>
      <div className="side-footer"><div><strong>{session.displayName}</strong><small>OA 管理员</small></div><button onClick={() => { Logout(); setSession(null); }}>退出</button></div>
    </aside>
    <section className="agent-main">
      <header className="agent-header"><div><strong>OA 权限助手</strong><span>仅连接 OA 平台权限能力</span></div><div className={`connection ${status.oaMode === "execute" ? "live" : "dry"}`}>{status.oaMode === "execute" ? "● OA 自动执行" : "◌ OA 演练模式"}</div></header>
      <div className="chat-area">
        {!messages.length && <Welcome status={status} setDraft={setDraft} />}
        {messages.map(message => <article className={`message ${message.role}`} key={message.id}><div className="avatar">{message.role === "assistant" ? "✦" : "你"}</div><div className="message-body"><div className="message-role">{message.role === "assistant" ? "OA 权限助手" : "你"}</div><div className="message-content">{message.content}</div></div></article>)}
        {busy && <article className="message assistant"><div className="avatar">✦</div><div className="message-body"><div className="message-role">OA 权限助手</div><div className="thinking"><i></i><i></i><i></i> 正在查询 OA 并执行任务…</div></div></article>}
        <div ref={chatEnd} />
      </div>
      <form className="composer" onSubmit={send}><textarea value={draft} onChange={event => setDraft(event.target.value)} placeholder="例如：给张三分配 A 平台权限" onKeyDown={event => { if (event.key === "Enter" && !event.shiftKey) { event.preventDefault(); event.currentTarget.form?.requestSubmit(); } }} /><div className="composer-bottom"><span>工具：OA 权限管理</span><button type="submit" disabled={busy || !draft.trim()}>发送 ↑</button></div></form>
      {notice && <div className="notice">{notice}</div>}
    </section>
  </main>;
}

function Welcome({ status, setDraft }: { status: Status; setDraft: (value: string) => void }) { return <div className="welcome"><div className="welcome-mark">✦</div><p className="eyebrow">OA AGENT WORKBENCH</p><h1>今天需要处理什么权限？</h1><p>我会只调用 OA 的员工、平台和权限接口完成任务。</p><div className="suggestions"><button onClick={() => setDraft("给张三分配 A 平台权限")}>给张三分配 A 平台权限 <span>→</span></button><button onClick={() => setDraft("给李四开通 OA 平台权限")}>给李四开通 OA 平台权限 <span>→</span></button></div><div className="tool-note"><strong>OA 权限工具已加载</strong><span>{status.oaMode === "execute" ? "执行模式：任务会自动写入 OA" : "演练模式：先查询并展示拟执行操作，不会写入 OA"}</span></div></div> }
function LoginScreen({ appName, login, setLogin, submit, notice }: any) { return <main className="login"><section className="login-card"><div className="login-logo">✦</div><p className="eyebrow">OA AGENT WORKBENCH</p><h1>{appName}</h1><p className="muted">本地桌面端只提供 OA 权限助手能力。</p><form onSubmit={submit}><label>本地管理账号<input value={login.account} onChange={e => setLogin({ ...login, account: e.target.value })} /></label><label>密码<input type="password" value={login.password} onChange={e => setLogin({ ...login, password: e.target.value })} /></label><button className="primary" type="submit">进入工作台</button></form><p className="hint">MVP 初始账号：admin / admin123</p>{notice && <div className="notice">{notice}</div>}</section></main> }
export default App;
