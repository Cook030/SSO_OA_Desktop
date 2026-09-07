import { Api, ApiError, ServerEvent } from './api';
import './workbench.css';

type Role = { id: number; name: string; code: string; is_builtin: number };
type Message = { id: string; role: string; content: string };
type Plan = { id: string; status: string; expires_at: string; payload: {
  employee: { name: string; account: string; department: string };
  before_roles: Role[]; added_roles: Role[]; after_roles: Role[];
} };
type Task = { id: string; status: string; error?: string; plans?: Plan[]; audit?: { action: string; created_at: string }[] };
type Conversation = { id: string; title: string; messages: Message[]; tasks: Task[] };

const root = document.querySelector<HTMLDivElement>('#root')!;
const api = new Api(localStorage.getItem('agent.serviceUrl') || 'http://127.0.0.1:8090');
let conversationId = '', activeTask: Task | null = null, stream: AbortController | null = null;
let generation = 0, busy = false, actorLabel = '', modelReady = false;
const completedMessages = new Set<string>();
const labels: Record<string, string> = { queued: '任务排队中', running: '正在处理', waiting_confirmation: '等待你的确认',
  succeeded: '已完成', failed: '需要关注', cancelled: '已取消', pending: '等待确认', approved: '已确认',
  executing: '正在执行', rejected: '已拒绝', conflict: '数据已变化', unknown: '执行结果待核查', expired: '已过期' };
const toolLabels: Record<string, string> = { search_employees: '查询员工', list_roles: '查询角色目录', get_user_roles: '查询员工角色',
  get_role_permissions: '查询角色权限', list_permissions: '查询权限目录', list_platforms: '查询平台', list_departments: '查询部门',
  compare_employee_access: '比较员工角色', plan_add_user_roles: '准备变更预览' };

function el<K extends keyof HTMLElementTagNameMap>(tag: K, className = '', text = ''): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag); node.className = className; node.textContent = text; return node;
}
function $(id: string) { return document.getElementById(id)!; }
function button(text: string, action: () => void, className = '') {
  const node = el('button', className, text); node.type = 'button'; node.addEventListener('click', action); return node;
}
function showError(error: unknown) {
  if (error instanceof DOMException && error.name === 'AbortError') return;
  if (error instanceof ApiError && error.status === 401) { api.token = ''; loginScreen(); }
  const notice = document.getElementById('notice');
  if (notice) { notice.textContent = error instanceof Error ? error.message : '操作失败，请重试'; notice.hidden = false; }
}
function resetStream() { generation++; stream?.abort(); stream = null; }

function loginScreen() {
  resetStream(); activeTask = null; conversationId = ''; busy = false; completedMessages.clear();
  // Only static markup. All user/model/OA text is inserted using textContent.
  root.innerHTML = `<main class="login"><section class="login-card"><div class="brand-icon">智</div>
    <p class="eyebrow">企业内部工作台</p><h1>让事务，在对话中完成。</h1>
    <p class="muted">使用企业账号登录，查询员工、核查权限并处理角色分配。</p>
    <form id="login-form"><label>工作台服务地址<input id="service-url" type="url" required autocomplete="url"></label>
    <label>企业账号<input id="account" required autocomplete="username" maxlength="100"></label>
    <label>密码<input id="password" type="password" required autocomplete="current-password"></label>
    <button class="primary" id="login-button">登录工作台 →</button></form>
    <p class="hint">使用现有企业 SSO 验证身份。凭据不会保存在本机。</p>
    <p id="notice" class="notice" role="alert" hidden></p></section></main>`;
  ($('service-url') as HTMLInputElement).value = api.baseUrl;
  $('login-form').addEventListener('submit', async event => {
    event.preventDefault(); const submit = $('login-button') as HTMLButtonElement; submit.disabled = true;
    try {
      const address = new URL(($('service-url') as HTMLInputElement).value);
      if (!['http:', 'https:'].includes(address.protocol) || address.username || address.password || address.search || address.hash || address.pathname !== '/') throw new Error('请填写有效的服务根地址');
      if (address.protocol === 'http:' && !['localhost', '127.0.0.1', '[::1]'].includes(address.hostname)) throw new Error('远程服务请使用 HTTPS 地址');
      api.baseUrl = address.origin;
      const password = $('password') as HTMLInputElement;
      const credentials = { account: ($('account') as HTMLInputElement).value.trim(), password: password.value };
      password.value = '';
      const result = await api.request<{ token: string; actor: { user_id: number; is_admin: boolean } }>('/api/auth/login', credentials);
      api.token = result.token;
      actorLabel = result.actor.is_admin ? '企业管理员' : `企业成员 · ${result.actor.user_id}`;
      localStorage.setItem('agent.serviceUrl', api.baseUrl);
      const status = await api.request<{ model_configured: boolean }>('/api/me');
      modelReady = status.model_configured;
      shell(); await loadConversations();
    } catch (error) { showError(error); }
    finally { submit.disabled = false; }
  });
}

function shell() {
  root.innerHTML = `<main class="shell"><aside class="sidebar"><div class="brand"><span class="brand-icon">智</span><strong>智工工作台</strong></div>
    <button id="new-chat" class="new-chat">＋ 新建任务</button><p class="section-label">最近会话</p><nav id="conversations" aria-label="最近会话"></nav>
    <footer><span id="actor"></span><button id="logout">退出</button></footer></aside>
    <section class="main"><header><div><strong>OA 工作助手</strong><span class="muted">员工 · 角色 · 权限</span></div><span id="connection" class="badge"></span></header>
    <div id="messages" class="messages" role="log" aria-label="对话内容"></div>
    <p id="notice" class="notice" role="alert" hidden></p>
    <form id="composer" class="composer"><label class="sr-only" for="draft">描述需要处理的工作</label><textarea id="draft" rows="3" maxlength="8000" placeholder="描述需要处理的工作，例如：查一下张三的角色和平台权限"></textarea>
    <div class="composer-footer"><span id="composer-hint">角色变更会先展示预览，确认后执行</span><button id="send" class="primary">发送 ↑</button></div></form></section>
    <aside class="task-panel"><div class="panel-heading"><h2>当前任务</h2><button id="cancel" hidden>停止</button></div><p id="task-status" class="muted">任务开始后在这里查看进度</p>
    <div id="progress" aria-live="polite"></div><div id="plan"></div><details id="audit-section" hidden><summary>操作记录</summary><div id="audit"></div></details></aside></main>`;
  $('actor').textContent = actorLabel;
  $('connection').textContent = modelReady ? '● 服务已连接' : '○ 模型待配置';
  $('new-chat').addEventListener('click', () => newConversation().catch(showError));
  $('logout').addEventListener('click', async () => {
    try { const result = await api.request<{ warning?: string }>('/api/auth/logout', {}); api.token = ''; loginScreen(); if (result.warning) showError(new Error(result.warning)); }
    catch (error) { api.token = ''; loginScreen(); showError(error); }
  });
  $('cancel').addEventListener('click', async () => {
    if (!activeTask) return;
    try { await api.request(`/api/tasks/${activeTask.id}/cancel`, {}); await selectConversation(conversationId); } catch (error) { showError(error); }
  });
  $('composer').addEventListener('submit', event => { event.preventDefault(); send().catch(showError); });
  $('draft').addEventListener('keydown', event => {
    if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) { event.preventDefault(); ($('composer') as HTMLFormElement).requestSubmit(); }
  });
  welcome(); setBusy(false);
}

function welcome() {
  const box = el('div', 'welcome');
  box.append(el('p', 'eyebrow', '你的企业工作助手'), el('h1', '', '今天，有什么需要处理？'), el('p', 'muted', '从查询开始，把角色与权限事务交给对话。'));
  const examples = el('div', 'suggestions');
  for (const text of ['查一下张三有哪些角色', '列出可以分配的角色', '比较张三和李四的角色差异']) {
    examples.append(button(text + ' ↗', () => { ($('draft') as HTMLTextAreaElement).value = text; $('draft').focus(); }));
  }
  box.append(examples); $('messages').replaceChildren(box);
}

async function loadConversations() {
  const list = await api.request<Conversation[]>('/api/conversations');
  const nav = document.getElementById('conversations'); if (!nav) return; nav.replaceChildren();
  for (const item of list) {
    const node = button(item.title, () => selectConversation(item.id).catch(showError), 'conversation' + (item.id === conversationId ? ' selected' : ''));
    if (item.id === conversationId) node.setAttribute('aria-current', 'page');
    nav.append(node);
  }
}

async function newConversation() {
  if (busy) return;
  const current = generation;
  setBusy(true);
  try {
    const item = await api.request<Conversation>('/api/conversations', {});
    if (current !== generation) return;
    setBusy(false);
    await selectConversation(item.id);
  } finally { setBusy(false); }
}

async function selectConversation(id: string) {
  if (busy) return;
  resetStream(); const current = generation;
  conversationId = id; activeTask = null; setBusy(true);
  let item: Conversation;
  try { item = await api.request<Conversation>(`/api/conversations/${id}`); }
  catch (error) { if (current === generation) setBusy(false); throw error; }
  if (generation !== current) return;
  $('messages').replaceChildren(); $('progress').replaceChildren(); $('plan').replaceChildren();
  $('notice').hidden = true; $('audit-section').hidden = true;
  if (!item.messages.length) welcome(); else for (const message of item.messages) renderMessage(message);
  activeTask = item.tasks[item.tasks.length - 1] || null;
  setBusy(false);
  if (activeTask) {
    const task = await api.request<Task>(`/api/tasks/${activeTask.id}`);
    if (generation !== current) return;
    renderTask(task);
    if (['queued', 'running'].includes(task.status)) watch(task.id, current).catch(showError);
  } else $('task-status').textContent = '任务开始后在这里查看进度';
  await loadConversations();
}

function renderMessage(message: Message, append = false) {
  if (append && completedMessages.has(message.id)) return;
  if (!append) completedMessages.add(message.id);
  const area = $('messages');
  area.querySelector('.welcome')?.remove();
  let node = Array.from(area.querySelectorAll<HTMLElement>('article')).find(x => x.dataset.messageId === message.id);
  if (!node) {
    node = el('article', 'message ' + (message.role === 'user' ? 'user' : 'assistant'));
    node.dataset.messageId = message.id;
    node.append(el('div', 'message-author', message.role === 'user' ? '你' : '智工'), el('div', 'message-content'));
    area.append(node);
  }
  const content = node.querySelector('.message-content')!;
  content.textContent = append ? content.textContent + message.content : message.content;
  if (area.scrollHeight - area.scrollTop - area.clientHeight < 240 || message.role === 'user') area.scrollTop = area.scrollHeight;
}

function setBusy(value: boolean) {
  busy = value;
  if (!document.getElementById('send')) return;
  const waiting = activeTask && ['queued', 'running', 'waiting_confirmation'].includes(activeTask.status);
  ($('send') as HTMLButtonElement).disabled = busy || !!waiting || !modelReady;
  ($('draft') as HTMLTextAreaElement).disabled = busy || !!waiting;
  $('cancel').hidden = !waiting;
  $('composer-hint').textContent = waiting ? '请先完成或停止当前任务' : '角色变更会先展示预览，确认后执行';
}

async function send() {
  const draft = $('draft') as HTMLTextAreaElement;
  const content = draft.value.trim();
  if (!content || busy || activeTask && ['queued', 'running', 'waiting_confirmation'].includes(activeTask.status)) return;
  const currentRequest = generation;
  setBusy(true);
  try {
    if (!conversationId) {
      const item = await api.request<Conversation>('/api/conversations', {});
      if (currentRequest !== generation) return;
      conversationId = item.id;
    }
    const task = await api.request<Task>(`/api/conversations/${conversationId}/messages`, { content });
    if (currentRequest !== generation) return;
    draft.value = ''; resetStream(); const current = generation;
    $('notice').hidden = true; $('progress').replaceChildren(); $('plan').replaceChildren();
    renderMessage({ id: 'user-' + task.id, role: 'user', content });
    renderTask(task); await loadConversations(); watch(task.id, current).catch(showError);
  } finally { setBusy(false); }
}

function renderTask(task: Task) {
  activeTask = task;
  $('task-status').textContent = labels[task.status] || task.status;
  if (task.error) { $('notice').textContent = task.error; $('notice').hidden = false; }
  $('plan').replaceChildren();
  for (const plan of task.plans || []) renderPlan(plan);
  $('audit').replaceChildren();
  for (const entry of task.audit || []) $('audit').append(el('p', 'audit-entry', `${entry.created_at.slice(11, 19)} · ${entry.action}`));
  $('audit-section').hidden = !task.audit?.length;
  setBusy(false);
}

function renderPlan(plan: Plan) {
  const card = el('section', 'plan-card'); const data = plan.payload;
  card.append(el('span', 'badge', labels[plan.status] || plan.status), el('h3', '', '增加员工角色'),
    el('p', '', `${data.employee.name} · ${data.employee.account}`), el('p', 'muted', data.employee.department || '未设置部门'));
  card.append(el('p', 'section-label', '现有角色'), el('p', '', data.before_roles.map(r => r.name).join('、') || '无'));
  card.append(el('p', 'section-label', '本次新增'));
  for (const role of data.added_roles) card.append(el('p', 'role-add', '+ ' + role.name + (role.is_builtin ? '（内置角色）' : '')));
  card.append(el('p', 'hint', '保留已有角色。请核对员工与新增角色后确认。'), el('small', 'plan-id', plan.id));
  if (plan.status === 'pending') {
    card.append(el('p', 'hint', '有效期至 ' + new Date(plan.expires_at).toLocaleTimeString()));
    const actions = el('div', 'plan-actions');
    const decide = async (approved: boolean) => {
      const currentView = generation;
      for (const b of Array.from(actions.querySelectorAll('button'))) b.disabled = true;
      try {
        const result = await api.request<{ task_id: string }>(`/api/operation-plans/${plan.id}/decision`, { approved });
        if (currentView !== generation) return;
        resetStream(); $('plan').replaceChildren(); $('progress').replaceChildren();
        if (activeTask) renderTask({ ...activeTask, status: 'queued', plans: [] });
        watch(result.task_id, generation).catch(showError);
      } catch (error) { showError(error); for (const b of Array.from(actions.querySelectorAll('button'))) b.disabled = false; }
    };
    actions.append(button('确认分配', () => void decide(true), 'primary'), button('取消', () => void decide(false), 'secondary')); card.append(actions);
  }
  $('plan').append(card);
}

async function watch(taskId: string, current: number) {
  stream?.abort(); const controller = new AbortController(); stream = controller;
  let cursor = 0, failures = 0;
  while (generation === current && !controller.signal.aborted) {
    try {
      await api.events(taskId, cursor, controller.signal, (event: ServerEvent) => {
        if (generation !== current || event.id <= cursor) return;
        cursor = event.id;
        const data = event.data;
        if (event.kind === 'task.running') $('task-status').textContent = '正在处理';
        if (event.kind === 'message.delta') renderMessage({ ...data, role: 'assistant' }, true);
        if (event.kind === 'message.completed') renderMessage(data);
        if (event.kind === 'tool.started') $('progress').append(el('p', 'progress-step', '◌ ' + (toolLabels[data.name] || '处理任务')));
        if (event.kind === 'tool.completed') {
          const node = $('progress').lastElementChild;
          if (node) node.textContent = (data.success ? '✓ ' : '! ') + (toolLabels[data.name] || '工具调用') + (data.error ? `：${data.error}` : '');
        }
        if (event.kind === 'operation.executing') $('task-status').textContent = '正在执行角色分配';
      });
      if (generation !== current || controller.signal.aborted) return;
      const task = await api.request<Task>(`/api/tasks/${taskId}`, undefined, controller.signal);
      if (generation !== current) return;
      renderTask(task); failures = 0;
      if (!['queued', 'running'].includes(task.status)) return;
    } catch (error) {
      if (controller.signal.aborted || generation !== current) return;
      if (error instanceof ApiError && error.status === 401) { showError(error); return; }
      failures++;
      if (failures >= 5) { showError(new Error('连接中断。服务端任务可能仍在执行，请重新打开此会话查看。')); return; }
      $('task-status').textContent = '连接中断，正在重新连接…';
      await new Promise(resolve => window.setTimeout(resolve, Math.min(1000 * 2 ** failures, 10000)));
    }
  }
}

loginScreen();
