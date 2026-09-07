export class ApiError extends Error {
  constructor(message: string, public status: number) { super(message); }
}

export type ServerEvent = { id: number; kind: string; data: any };

// Exported separately so fragmentation, CRLF and UTF-8 can be verified without DOM.
export class SSEParser {
  private buffer = '';
  constructor(private receive: (event: ServerEvent) => void) {}
  feed(text: string) {
    this.buffer += text;
    if (this.buffer.length > 2_000_000) throw new Error('事件内容过大');
    let separator: RegExpExecArray | null;
    while ((separator = /\r?\n\r?\n/.exec(this.buffer))) {
      const block = this.buffer.slice(0, separator.index);
      this.buffer = this.buffer.slice(separator.index + separator[0].length);
      let id = 0, kind = 'message';
      const data: string[] = [];
      for (const line of block.split(/\r?\n/)) {
        const colon = line.indexOf(':');
        if (colon < 1) continue;
        const key = line.slice(0, colon);
        const value = line.slice(colon + 1).replace(/^ /, '');
        if (key === 'id') id = Number(value);
        if (key === 'event') kind = value;
        if (key === 'data') data.push(value);
      }
      if (data.length) this.receive({ id, kind, data: JSON.parse(data.join('\n')) });
    }
  }
}

export class Api {
  token = '';
  constructor(public baseUrl: string) {}

  async request<T>(path: string, body?: unknown, signal?: AbortSignal): Promise<T> {
    const response = await fetch(this.baseUrl + path, {
      method: body === undefined ? 'GET' : 'POST', signal,
      headers: { ...(this.token ? { Authorization: `Bearer ${this.token}` } : {}),
        ...(body === undefined ? {} : { 'Content-Type': 'application/json' }) },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    const result = await response.json();
    if (!response.ok) throw new ApiError(typeof result.detail === 'string' ? result.detail : '请求参数无效', response.status);
    return result as T;
  }

  async events(taskId: string, after: number, signal: AbortSignal, receive: (event: ServerEvent) => void) {
    const response = await fetch(`${this.baseUrl}/api/tasks/${taskId}/events?after=${after}`, {
      signal, headers: { Authorization: `Bearer ${this.token}` },
    });
    if (!response.ok || !response.body) {
      throw new ApiError('任务连接失败，请检查服务或重新登录', response.status);
    }
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    const parser = new SSEParser(receive);
    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) { parser.feed(decoder.decode()); break; }
        parser.feed(decoder.decode(value, { stream: true }));
      }
    } finally { await reader.cancel().catch(() => undefined); reader.releaseLock(); }
  }
}
