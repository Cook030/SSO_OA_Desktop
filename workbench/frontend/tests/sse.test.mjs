import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/api.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ES2022 } });
const { SSEParser, Api } = await import('data:text/javascript;base64,' + Buffer.from(outputText).toString('base64'));

test('fragmented CRLF and Chinese payload reconstruct a single SSE event', () => {
  const events = [];
  const parser = new SSEParser(event => events.push(event));
  const source = 'id: 12\r\nevent: message.delta\r\ndata: {"content":"正在查询员工"}\r\n\r\n';
  for (const char of source) parser.feed(char);
  assert.deepEqual(events, [{ id: 12, kind: 'message.delta', data: { content: '正在查询员工' } }]);
});

test('heartbeats do not emit events and multiple records retain order', () => {
  const events = [];
  const parser = new SSEParser(event => events.push(event));
  parser.feed(': heartbeat\n\nid: 2\nevent: tool.started\ndata: {"name":"search"}\n\nid: 3\nevent: tool.completed\ndata: {"success":true}\n\n');
  assert.deepEqual(events.map(x => [x.id, x.kind]), [[2, 'tool.started'], [3, 'tool.completed']]);
});

test('multiline JSON is joined according to the SSE protocol', () => {
  const events = [];
  const parser = new SSEParser(event => events.push(event));
  parser.feed('id: 4\ndata: {\ndata: "ok": true\ndata: }\n\n');
  assert.equal(events[0].data.ok, true);
});

test('streaming fetch keeps bearer out of URL and handles split UTF-8 bytes', async t => {
  const text = 'id: 7\nevent: message.delta\ndata: {"content":"张三"}\n\n';
  const bytes = new TextEncoder().encode(text);
  t.mock.method(globalThis, 'fetch', async (url, options) => {
    assert.equal(url, 'http://test-agent/api/tasks/task-1/events?after=6');
    assert.equal(options.headers.Authorization, 'Bearer secret');
    return new Response(new ReadableStream({ start(controller) {
      for (const byte of bytes) controller.enqueue(new Uint8Array([byte]));
      controller.close();
    } }));
  });
  const api = new Api('http://test-agent'); api.token = 'secret';
  const events = [];
  await api.events('task-1', 6, new AbortController().signal, x => events.push(x));
  assert.deepEqual(events, [{ id: 7, kind: 'message.delta', data: { content: '张三' } }]);
});
