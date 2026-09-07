import httpx
import pytest
from cryptography.fernet import Fernet
from langchain_core.messages import AIMessage, ToolMessage
from sqlalchemy import inspect

from app.config import Settings
from app.graph import AgentGraph
from app.lease import single_instance
from app.main import create_app
from app.store import Store


async def test_startup_creates_schema_and_service_starts_without_model(tmp_path, monkeypatch):
    monkeypatch.setenv("AGENT_ENCRYPTION_KEY", Fernet.generate_key().decode())
    monkeypatch.setenv("AGENT_DATABASE_URL", f"sqlite:///{tmp_path / 'agent.db'}")
    monkeypatch.setenv("AGENT_CHECKPOINT_URL", str(tmp_path / "graph.db"))
    monkeypatch.setenv("AGENT_MODEL_NAME", "")
    monkeypatch.setenv("AGENT_MODEL_API_KEY", "")
    app = create_app(Settings(_env_file=None))
    async with app.router.lifespan_context(app):
        tables = set(inspect(app.state.runtime.store.engine).get_table_names())
        assert {
            "agent_sessions",
            "conversations",
            "tasks",
            "messages",
            "operation_plans",
            "task_events",
            "agent_audit_logs",
        } <= tables
        async with httpx.AsyncClient(
            transport=httpx.ASGITransport(app=app), base_url="http://agent"
        ) as client:
            assert (await client.get("/healthz")).json() == {"status": "ok"}
            assert (await client.get("/api/conversations")).status_code == 401


def test_second_worker_refused_for_same_database(tmp_path):
    store = Store(f"sqlite:///{tmp_path / 'agent.db'}")
    with single_instance(store.engine):
        with pytest.raises(RuntimeError, match="one Agent server"):
            with single_instance(store.engine):
                pytest.fail("Second worker acquired same database")
    with single_instance(store.engine):
        pass  # Lock is released after shutdown.
    store.engine.dispose()


async def test_cancelled_tool_call_history_is_repaired_before_next_model_request(env):
    client, rt = env["client"], env["runtime"]
    conversation_id = (await client.post("/api/conversations", json={})).json()["id"]
    graph = AgentGraph(rt, "unused", conversation_id, "unused").compile()
    await graph.aupdate_state(
        {"configurable": {"thread_id": conversation_id}},
        {
            "messages": [
                AIMessage(
                    content="",
                    tool_calls=[
                        {"id": "interrupted-call", "name": "search_employees", "args": {"keyword": "张三"}}
                    ],
                )
            ],
        },
        as_node="reason",
    )
    env["model"].script.append("可以重新查询。")
    response = await client.post(
        f"/api/conversations/{conversation_id}/messages", json={"content": "继续查询"}
    )
    assert response.status_code == 202
    await rt.jobs[response.json()["id"]]
    messages = env["model"].inputs[-1]
    assert any(isinstance(m, ToolMessage) and m.tool_call_id == "interrupted-call" for m in messages)
    assert env["oa"].write_count == 0


async def test_two_messages_for_same_conversation_are_serialized(env):
    import asyncio

    client = env["client"]
    env["model"].flow()
    conversation = (await client.post("/api/conversations", json={})).json()["id"]
    first, second = await asyncio.gather(
        *[
            client.post(f"/api/conversations/{conversation}/messages", json={"content": "给张三增加角色"})
            for _ in range(2)
        ]
    )
    assert sorted([first.status_code, second.status_code]) == [202, 409]
    accepted = first if first.status_code == 202 else second
    await env["runtime"].jobs[accepted.json()["id"]]
    assert env["oa"].write_count == 0
